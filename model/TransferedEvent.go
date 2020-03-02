package model

// TransferedEvent is the log of the TransferedEvent
type TransferedEvent struct {
	Sender          string `json:"sender"`
	Recipient       string `json:"recipient"`
	TransferedMoney string `json:"transferedMoney"`
}

// newTransferedEvent is ...
func newTransferedEvent(sender, recipient, transferedMoney string) *TransferedEvent {
	return &TransferedEvent{sender, recipient, transferedMoney}
}
