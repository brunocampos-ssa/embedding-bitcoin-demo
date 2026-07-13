package bounty

import "time"

type State string

const (
	Draft     State = "DRAFT"
	Open      State = "OPEN"
	Assigned  State = "ASSIGNED"
	Submitted State = "SUBMITTED"
	Approved  State = "APPROVED"
	Paid      State = "PAID"
	Cancelled State = "CANCELLED"
	Expired   State = "EXPIRED"
)

type Bounty struct {
	ID          string    `json:"id"`
	Title       string    `json:"title"`
	Description string    `json:"description"`
	Format      string    `json:"format"`
	Language    string    `json:"language"`
	RewardSats  int64     `json:"rewardSats"`
	State       State     `json:"state"`
	CreatedAt   time.Time `json:"createdAt"`
}

func CanTransition(from, to State) bool {
	allowed := map[State][]State{Draft: {Open, Cancelled}, Open: {Assigned, Cancelled, Expired}, Assigned: {Submitted, Cancelled}, Submitted: {Approved, Assigned}, Approved: {Paid, Cancelled}}
	for _, v := range allowed[from] {
		if v == to {
			return true
		}
	}
	return false
}
