package notifier

// Dummy is a stub
type Dummy struct{}

// SendBroadcastMessage does nothing
func (*Dummy) SendBroadcastMessage(message string) {}
