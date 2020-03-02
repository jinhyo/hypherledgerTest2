package model

// ApprovalEvent is the log of the ApprovalEvent
type ApprovalEvent struct {
	Owner   string `json:"owner"`
	Spender string `json:"spender"`
	Amount  int    `json:"amount"`
}

// NewApprovalEvent is ...
func NewApprovalEvent(owner, spender string, amount int) *ApprovalEvent {
	return &ApprovalEvent{owner, spender, amount}
}
