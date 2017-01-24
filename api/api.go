package api

import (
	"fmt"
	"net/http"

	"github.com/siapool/p2pool/sharechain"
)

//PoolAPI implements the http handlers
type PoolAPI struct {
	//Fee is the poolfee in 0.01%
	Fee int
	//ShareChain for getting work and posting shares
	ShareChain *sharechain.ShareChain
	//Version is the poolversion
	Version string
}

//FeeHandler writes the fee applied by the pool
func (pa *PoolAPI) FeeHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Fprintf(w, "%.2f%%", float64(pa.Fee)/100)
}

//VersionHandler writes the software version of the pool
func (pa *PoolAPI) VersionHandler(w http.ResponseWriter, r *http.Request) {
	fmt.Print(w, pa.Version)
}
