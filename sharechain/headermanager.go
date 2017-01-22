package sharechain

import "github.com/NebulousLabs/Sia/types"

// HeaderForWork returns a header that is ready for nonce grinding.
func (sc *ShareChain) HeaderForWork(payoutaddress string) (blockheader types.BlockHeader, target types.Target, err error) {
	err = sc.tg.Add()
	if err != nil {
		return
	}
	defer sc.tg.Done()

	sc.mu.Lock()
	defer sc.mu.Unlock()

	return
}
