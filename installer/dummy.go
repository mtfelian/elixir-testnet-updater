package installer

// Dummy is a stub
type Dummy struct{}

// IsInstalled always return true
func (d Dummy) IsInstalled() bool { return true }

// Install does nothing
func (d Dummy) Install() error { return nil }
