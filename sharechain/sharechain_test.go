package sharechain

import (
	"math/big"
	"testing"

	"github.com/rivine/rivine/types"
)

func TestTarget(t *testing.T) {
	hashesPerShare := big.NewInt(StartHashesPerShare)
	calculatesHashesPerShare := types.RootDepth.Int().Div(types.RootDepth.Int(), StartTarget.Int())
	if hashesPerShare.Cmp(calculatesHashesPerShare) != 0 {
		t.Fail()
	}

}
