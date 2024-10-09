package installer

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"text/template"
)

// Systemd install system service
type Systemd struct {
	serviceName string
	binaryPath  string
	user        string
}

// SystemdParams represents systemd service params
type SystemdParams struct {
	ServiceName string
	User        string
}

// NewSystemd creates new systemd service installer
func NewSystemd(p SystemdParams) (*Systemd, error) {
	ss := &Systemd{
		serviceName: p.ServiceName,
		user:        p.User,
	}

	var err error
	ss.binaryPath, err = os.Executable()
	if err != nil {
		return nil, err
	}
	return ss, nil
}

// unitFilePath returns a path to systemd service definition file
func (ss *Systemd) unitFilePath() string {
	return "/etc/systemd/system/" + ss.serviceName + ".service"
}

const systemdUnitTemplate = `[Unit]
Description=Elixir Updater
After=network.target

[Service]
ExecStart={{.ExecStart}}
Restart=always
User={{.User}}
WorkingDirectory={{.WorkingDirectory}}
Environment=DOCKER_API_VERSION=1.43
StandardOutput=syslog
StandardError=syslog

[Install]
WantedBy=multi-user.target
`

// SystemdUnitFileData represents data which describes a Linux system service
type SystemdUnitFileData struct {
	ExecStart        string
	User             string
	WorkingDirectory string
}

// writeUnitFile writes the systemd unit file to the specified path
func (ss *Systemd) writeUnitFile() error {
	unitData := SystemdUnitFileData{
		ExecStart:        ss.binaryPath,
		User:             ss.user,
		WorkingDirectory: filepath.Dir(ss.binaryPath),
	}

	tmpl, err := template.New("systemd").Parse(systemdUnitTemplate)
	if err != nil {
		return fmt.Errorf("error parsing template: %v", err)
	}

	file, err := os.Create(ss.unitFilePath())
	if err != nil {
		return fmt.Errorf("error creating unit file: %v", err)
	}
	defer file.Close()

	if err = tmpl.Execute(file, unitData); err != nil {
		return fmt.Errorf("error writing unit file: %v", err)
	}

	return nil
}

// enableAndStartService enables and starts the systemd service
func (ss *Systemd) enableAndStartService() error {
	if err := exec.Command("systemctl", "daemon-reload").Run(); err != nil {
		return fmt.Errorf("error reloading systemd daemon: %v", err)
	}

	if err := exec.Command("systemctl", "enable", ss.serviceName).Run(); err != nil {
		return fmt.Errorf("error enabling service: %v", err)
	}

	if err := exec.Command("systemctl", "start", ss.serviceName).Run(); err != nil {
		return fmt.Errorf("error starting service: %v", err)
	}

	return nil
}

// IsInstalled returns whether the system service exists
func (ss *Systemd) IsInstalled() bool {
	_, err := os.Stat(ss.unitFilePath())
	return err == nil
}

// Install accepts path to binary file as binaryPath parameter and installs it as a service
func (ss *Systemd) Install() error {
	if err := ss.writeUnitFile(); err != nil {
		fmt.Printf("Failed to write systemd unit file: %v\n", err)
		return err
	}

	if err := ss.enableAndStartService(); err != nil {
		fmt.Printf("Failed to enable and start service: %v\n", err)
		return err
	}

	fmt.Println("Systemd service created, enabled, and started successfully.")
	return nil
}
