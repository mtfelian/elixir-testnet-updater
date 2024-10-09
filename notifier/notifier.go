package notifier

// Notifier abstracts TG bot client
type Notifier interface {
	SendBroadcastMessage(message string)
}
