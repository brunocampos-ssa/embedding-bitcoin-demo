package bounty

import "testing"

func TestTransitions(t *testing.T) {
	for _, tc := range []struct {
		from, to State
		want     bool
	}{{Draft, Open, true}, {Open, Assigned, true}, {Assigned, Submitted, true}, {Submitted, Approved, true}, {Approved, Paid, true}, {Open, Paid, false}, {Paid, Approved, false}} {
		if got := CanTransition(tc.from, tc.to); got != tc.want {
			t.Errorf("CanTransition(%s,%s)=%v", tc.from, tc.to, got)
		}
	}
}
