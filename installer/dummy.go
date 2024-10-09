package installer

// Dummy is a stub
type Dummy struct{}

// IsInstalled always return false
func (d Dummy) IsInstalled() bool { return false }

// Install does nothing
func (d Dummy) Install() error { return nil }
