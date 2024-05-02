package gosnippets

import "testing"

func TestSwapPairsInvalid(t *testing.T) {
	if got := swapPairs(nil); got != nil {
		t.Errorf("swapPairs(nil) == %p, want nil", got)
	}
}

func TestSwapPairsValid(t *testing.T) {
	cases := []struct {
		in, want *ListNode
	}{
		
	}
}