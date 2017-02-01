package stratum

import (
	"encoding/hex"
	"strconv"
	"testing"

	"github.com/NebulousLabs/Sia/types"
)

func TestDifficultyToTarget(t *testing.T) {
	expectedDiff, _ := strconv.ParseFloat("0.99998474121094105", 64)
	var target types.Target
	targetSlice, _ := hex.DecodeString("00000000fffffffffffefffeffff00000001000200020000fffefffcfffbfffd")
	copy(target[:], targetSlice[:])
	diff := targetToDifficulty(target)

	if diff != expectedDiff {
		t.Error(diff, "returned instead of", expectedDiff)
	}
}
