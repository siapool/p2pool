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
	}
}
