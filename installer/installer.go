package installer

// Installer represents system service installer
type Installer interface {
	IsInstalled() bool
	Install() error
}
