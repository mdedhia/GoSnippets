package gosnippets

import (
	"slices"
	"testing"
)

func TestSwapPairsInvalid(t *testing.T) {
	if got := swapPairs(nil); got != nil {
		t.Errorf("swapPairs(nil) == %p, want nil", got)
	}
}

func TestSwapPairsValid(t *testing.T) {
	cases := []struct {
		in, want []int
	}{
		{[]int{1,2,3,4}, []int{2,1,4,3}},
		{[]int{1,2,3}, []int{2,1,3}},
	}

	for _, c := range cases {
		got := parseList(swapPairs(makeList(c.in)))
		if !slices.Equal(got, c.want) {
			t.Errorf("swapPairs(%q) == %q, want %q", c.in, got, c.want)
		}
	}
}

func parseList(head *ListNode) []int {
	var ret []int
	for node := head; node != nil; node = node.Next {
		ret = append(ret, node.Val)
	}
	return ret
}

func makeList(vals []int) *ListNode {
	if len(vals) == 0 {
		return nil
	}

	var head, prevNode *ListNode

	for _, val := range vals {
		var node ListNode
		node.Val = val
		
		if head == nil {
			head = &node
		} else {
			prevNode.Next = &node
		}
		prevNode = &node
	}
	return head
}
