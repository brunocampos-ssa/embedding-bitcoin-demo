package submission

import "time"

type State string

const (
	Submitted State = "SUBMITTED"
	Approved  State = "APPROVED"
	Rejected  State = "REJECTED"
)

type Submission struct {
	ID          string     `json:"id"`
	BountyID    string     `json:"bountyId"`
	Actor       string     `json:"actor"`
	EvidenceURL string     `json:"evidenceUrl"`
	Notes       string     `json:"notes"`
	State       State      `json:"state"`
	CreatedAt   time.Time  `json:"createdAt"`
	ApprovedAt  *time.Time `json:"approvedAt,omitempty"`
}
