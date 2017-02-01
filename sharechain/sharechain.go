package sharechain

import (
	"math/big"

	"github.com/NebulousLabs/Sia/persist"
	siasync "github.com/NebulousLabs/Sia/sync"
	"github.com/NebulousLabs/Sia/types"
	"github.com/NebulousLabs/demotemutex"
	"github.com/siapool/p2pool/siad"
)

const (
	//ShareChainLength is the number of shares the chain can hold, given a share twice per minute, it holds 4 days worth of shares
	ShareChainLength = 2 * 1440 * 4
	//ShareTime is the target time between two shares (like block time in a normal blockchain)
	ShareTime = 30
	//StartHashesPerShare is currently set for a 1Gh/s miner to find 2 shares per day
	StartHashesPerShare = 1 * 1000 * 1000 * 1000 * 3600 * 24 / 2
)

//StartTarget is the target for a share to be accepted when starting the p2pool,
var StartTarget = types.RootDepth.MulDifficulty(big.NewRat(StartHashesPerShare, 1))

//ShareChain holds the previous shares of the pool
type ShareChain struct {

	//Siad is the handler towards the sia daemon
	Siad *siad.Siad

	// Utilities
	db         *persist.BoltDatabase
	log        *persist.Logger
	mu         demotemutex.DemoteMutex
	persistDir string

	// tg signals the Miner's goroutines to shut down and blocks until all
	// goroutines have exited before returning from Close().
	tg siasync.ThreadGroup

	Target types.Target
}

// New returns a new ShareChain.
// If there is an existing sharechain database present in the persist directory, it is loaded.
func New(siadaemon *siad.Siad, persistDir string) (sc *ShareChain, err error) {

	sc = &ShareChain{
		Siad: siadaemon,

		persistDir: persistDir,

		Target: StartTarget,
	}

	// Initialize the persistence structures.
	err = sc.initPersist()

	return
}

//Share is a block with a lower difficulty target
type Share struct {
	BlockID   types.BlockID
	ParentID  types.BlockID
	Timestamp types.Timestamp
	Miner     string
}

//GetPPLNSSummary returns a mapping between miner addresses and the number of shares they found (within the ShareChainLength last number of shares)
func (sc *ShareChain) GetPPLNSSummary() (sharesummary map[string]int, err error) {
	//TODO
	return
}
