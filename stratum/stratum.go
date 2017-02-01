//Package stratum implements the basic stratum protocol.
// This is normal jsonrpc but the go standard library is insufficient since we need features like notifications.
package stratum

import (
	"bufio"
	"encoding/hex"
	"encoding/json"
	"errors"
	"math/big"
	"math/rand"
	"net"
	"sync"
	"time"

	"github.com/NebulousLabs/Sia/types"
	log "github.com/Sirupsen/logrus"

	"github.com/siapool/p2pool/sharechain"
)

// message is the structure for both requests, responses and notifications
type message struct {
	Method string        `json:"method,omitempty"`
	Params []interface{} `json:"params,omitempty"`
	ID     uint64        `json:"id,omitempty"`
	Result interface{}   `json:"result,omitempty"`
	Error  []interface{} `json:"error,omitempty"`
}

//ErrorCallback is the type of function that be registered to be notified of errors requiring a client connection
// to be dropped and a new one to be created
type ErrorCallback func(err error)

//NotificationHandler is the signature for a function that handles notifications
type NotificationHandler func(args []interface{})

// ClientConnection maintains a connection to a stratum client and (de)serializes requests/reponses/notifications
type ClientConnection struct {
	server *Server

	socketMutex sync.Mutex // protects following
	socket      net.Conn

	seqmutex sync.Mutex // protects following
	seq      uint64

	callsMutex   sync.Mutex // protects following
	pendingCalls map[uint64]chan interface{}

	ErrorCallback        ErrorCallback
	notificationHandlers map[string]NotificationHandler

	extranonce1  []byte
	MinerVersion string
	User         string
}

//NewClientConnection creates a new ClientConnection given a socket
func (server *Server) NewClientConnection(socket net.Conn) (c *ClientConnection) {
	extranonce1 := server.generateExtraNonce1()
	return &ClientConnection{socket: socket, extranonce1: extranonce1, server: server}
}

// Server Listens on a connection for incoming connections
type Server struct {
	shareChain *sharechain.ShareChain
	difficulty float64

	laddr          string
	maxConnections int

	lismutex sync.Mutex // protects following
	lis      net.Listener

	clientconnectionmutex sync.Mutex // protects following
	connections           []*ClientConnection

	ErrorCallback        ErrorCallback
	notificationHandlers map[string]NotificationHandler
}

//NewServer creates a stratum server for listening on the local network address laddr.
// During the Accept() call, a listening socket is created ( https://golang.org/pkg/net/#Listen ) using "tcp" as network and laddr as specified.
func NewServer(laddr string, shareChain *sharechain.ShareChain) (server *Server) {
	server = &Server{laddr: laddr, shareChain: shareChain, maxConnections: 1000}
	server.difficulty = targetToDifficulty(shareChain.Target)
	return
}

func targetToDifficulty(target types.Target) (difficulty float64) {
	//target = targetone/diff
	diffOneString := "0x00000000ffff0000000000000000000000000000000000000000000000000000"
	targetOneAsBigInt := &big.Int{}
	targetOneAsBigInt.SetString(diffOneString, 0)
	targetOneAsBigRat := &big.Rat{}
	targetOneAsBigRat.SetInt(targetOneAsBigInt)
	difficulty, _ = targetOneAsBigRat.Quo(targetOneAsBigRat, target.Rat()).Float64()
	return
}

func generateRandomBytes(length int) ([]byte, error) {
	b := make([]byte, length)
	_, err := rand.Read(b)
	if err != nil {
		return b, err
	}
	return b, nil
}

func (server *Server) generateExtraNonce1() (extraNonce1 []byte) {
retry:
	for {
		extraNonce1, _ = generateRandomBytes(4)
		for _, c := range server.connections {
			if hex.EncodeToString(extraNonce1) == hex.EncodeToString(c.extranonce1) {
				continue retry
			}
		}
		return
	}
}

//Accept creates  connections on the listener and serves requests for each incoming connection.
// Accept blocks until the underlying tcp listener returns a non-nil error or Close is called on the server.
// The caller typically invokes Accept in a go statement.
func (server *Server) Accept() (err error) {
	func() {
		server.lismutex.Lock()
		defer server.lismutex.Unlock()
		server.clientconnectionmutex.Lock()
		defer server.clientconnectionmutex.Unlock()
		server.lis, err = net.Listen("tcp", server.laddr)
		server.connections = make([]*ClientConnection, 0, 10)
		log.Infoln("Listening for incoming stratum connections on", server.laddr)
	}()
	if err != nil {
		return
	}
	for {
		err = func() (err error) {
			server.lismutex.Lock()
			defer server.lismutex.Unlock()
			conn, err := server.lis.Accept()
			if err != nil {
				return
			}
			server.clientconnectionmutex.Lock()
			defer server.clientconnectionmutex.Unlock()
			c := server.NewClientConnection(conn)
			if len(server.connections) >= server.maxConnections {
				log.Errorln("Maximum number of client connections reached (", server.maxConnections, "), dropping connection request")
				c.Close()
			}

			server.connections = append(server.connections, c)
			go c.Listen()
			//TODO: clean up closed client connections
			return
		}()
		if err != nil {
			return
		}
	}
}

//Close releases the underlying tcp listener
func (server *Server) Close() {
	if server.lis != nil {
		server.lis.Close()
	}
}

//Close releases the tcp connection
func (c *ClientConnection) Close() {
	if c.socket != nil {
		c.socket.Close()
	}
}

//SetNotificationHandler registers a function to handle notification for a specific method.
// This function is not threadsafe and all notificationhandlers should be set prior to calling the Dial function
func (c *ClientConnection) SetNotificationHandler(method string, handler NotificationHandler) {
	if c.notificationHandlers == nil {
		c.notificationHandlers = make(map[string]NotificationHandler)
	}
	c.notificationHandlers[method] = handler
}

func (c *ClientConnection) dispatchNotification(n message) {
	if c.notificationHandlers == nil {
		return
	}
	if notificationHandler, exists := c.notificationHandlers[n.Method]; exists {
		notificationHandler(n.Params)
	}
}

func (c *ClientConnection) dispatch(r message) {
	if r.ID == 0 {
		c.dispatchNotification(r)
		return
	}
	c.callsMutex.Lock()
	defer c.callsMutex.Unlock()
	cb, found := c.pendingCalls[r.ID]
	var result interface{}
	if r.Error != nil {
		message := ""
		if len(r.Error) >= 2 {
			message, _ = r.Error[1].(string)
		}
		result = errors.New(message)
	} else {
		result = r.Result
	}
	if found {
		cb <- result
	} else {
		switch r.Method {
		case "mining.subscribe":
			c.MiningSubscribeHandler(r)
		case "mining.authorize":
			c.MiningAuthorizeHandler(r)
		default:
			log.Debugln("unknown json-rpc method called on stratum server:", r.Method, "-", r)
		}
	}
}

func (c *ClientConnection) dispatchError(err error) {
	if c.ErrorCallback != nil {
		c.ErrorCallback(err)
	}
}

//Listen reads data from the open connection, deserializes it and dispatches the reponses and notifications
// This is a blocking function and will continue to listen until an error occurs (io or deserialization)
func (c *ClientConnection) Listen() {
	reader := bufio.NewReader(c.socket)
	for {
		rawmessage, err := reader.ReadString('\n')
		if err != nil {
			c.dispatchError(err)
			return
		}
		r := message{}
		err = json.Unmarshal([]byte(rawmessage), &r)
		if err != nil {
			c.dispatchError(err)
			return
		}
		c.dispatch(r)
	}
}

func (c *ClientConnection) registerRequest(requestID uint64) (cb chan interface{}) {
	c.callsMutex.Lock()
	defer c.callsMutex.Unlock()
	if c.pendingCalls == nil {
		c.pendingCalls = make(map[uint64]chan interface{})
	}
	cb = make(chan interface{})
	c.pendingCalls[requestID] = cb
	return
}

func (c *ClientConnection) cancelRequest(requestID uint64) {
	c.callsMutex.Lock()
	defer c.callsMutex.Unlock()
	cb, found := c.pendingCalls[requestID]
	if found {
		close(cb)
		delete(c.pendingCalls, requestID)
	}
}

//Call invokes the named function, waits for it to complete, and returns its error status.
func (c *ClientConnection) Call(serviceMethod string, args []interface{}) (reply interface{}, err error) {
	r := message{Method: serviceMethod, Params: args}

	c.seqmutex.Lock()
	c.seq++
	r.ID = c.seq
	c.seqmutex.Unlock()

	rawmsg, err := json.Marshal(r)
	if err != nil {
		return
	}
	call := c.registerRequest(r.ID)
	defer c.cancelRequest(r.ID)

	rawmsg = append(rawmsg, []byte("\n")...)
	func() {
		c.socketMutex.Lock()
		defer c.socketMutex.Unlock()
		_, err = c.socket.Write(rawmsg)
	}()
	if err != nil {
		return
	}
	//Make sure the request is cancelled if no response is given
	go func() {
		time.Sleep(10 * time.Second)
		c.cancelRequest(r.ID)
	}()
	reply = <-call

	if reply == nil {
		err = errors.New("Timeout")
		return
	}
	err, _ = reply.(error)
	return
}

func (c *ClientConnection) Reply(ID uint64, result interface{}, errorResult []interface{}) (err error) {
	r := message{ID: ID, Result: result, Error: errorResult}

	rawmsg, err := json.Marshal(r)
	if err != nil {
		return
	}
	rawmsg = append(rawmsg, []byte("\n")...)
	func() {
		c.socketMutex.Lock()
		defer c.socketMutex.Unlock()
		_, err = c.socket.Write(rawmsg)
	}()
	if err != nil {
		return
	}
	return
}

//Notify sends a notification to the client
func (c *ClientConnection) Notify(serviceMethod string, args []interface{}) (err error) {
	r := message{Method: serviceMethod, Params: args}

	rawmsg, err := json.Marshal(r)
	if err != nil {
		return
	}
	rawmsg = append(rawmsg, []byte("\n")...)
	func() {
		c.socketMutex.Lock()
		defer c.socketMutex.Unlock()
		_, err = c.socket.Write(rawmsg)
	}()
	if err != nil {
		return
	}
	return
}
