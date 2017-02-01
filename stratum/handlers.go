package stratum

import "encoding/hex"

//MiningSubscribeHandler handles the mining.subscribe request
func (c *ClientConnection) MiningSubscribeHandler(m message) {
	if m.Params != nil && len(m.Params) > 0 {
		c.MinerVersion, _ = m.Params[0].(string)
	}
	err := c.Reply(
		m.ID,
		[]interface{}{
			[]interface{}{
				[]interface{}{"mining.set_difficulty", "b4b6693b72a50c7116db18d6497cac52"},
				[]interface{}{"mining.notify", "ae6812eb4cd7735a302a8a9dd95cf71f"},
			},
			hex.EncodeToString(c.extranonce1),
			4,
		},
		nil)
	if err != nil {
		c.Close()
		return
	}
}

//MiningAuthorizeHandler handles the mining.authorize request
func (c *ClientConnection) MiningAuthorizeHandler(m message) {
	if m.Params == nil || len(m.Params) == 0 {
		c.sendErrorAndClose(m.ID, "Mining address required")
		return
	}
	user, ok := m.Params[0].(string)
	if !ok {
		c.sendErrorAndClose(m.ID, "Invalid mining address")
		return
	}
	//TODO: validate the supplied address.rigname
	c.User = user

	err := c.Reply(m.ID, true, nil)
	if err != nil {
		c.Close()
		return
	}
	c.SendDifficulty()
}

func (c *ClientConnection) sendErrorAndClose(ID uint64, errormessage string) {
	c.Reply(ID, nil, []interface{}{errormessage})
	c.Close()
}

//SendDifficulty sends the current difficulty to the miner
func (c *ClientConnection) SendDifficulty() {
	err := c.Notify("mining.set_difficulty", []interface{}{c.server.difficulty})
	if err != nil {
		c.Close()
	}
	return
}
