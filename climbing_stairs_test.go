package gosnippets

import "testing"

func TestClimbStairsInvalid(t *testing.T) {
	cases := []struct {
		in, want int
	}{
		{0, -1},
		{46, -1},
	}

	for _, c := range cases {
		got := climbStairs(c.in)
		if got != c.want {
			t.Errorf("climbStairs(%q) == %q, want %q", c.in, got, c.want)
		}
	}
}

func TestClimbStairsValid(t *testing.T) {
	cases := []struct {
		in, want int
	}{
		{2, 2},
		{45, 1836311903},
	}

	for _, c := range cases {
		got := climbStairs(c.in)
		if got != c.want {
			t.Errorf("climbStairs(%d) == %d, want %d", c.in, got, c.want)
		}
	}
}

