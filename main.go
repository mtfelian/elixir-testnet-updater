package main

import (
	"context"
	"fmt"
	"log"
	"time"

	"github.com/mtfelian/elixir-testnet-updater/config"
	"github.com/mtfelian/elixir-testnet-updater/installer"
	"github.com/mtfelian/elixir-testnet-updater/service"
)

const delay = 3 * time.Second

var svc *service.Service

func main() {
	fmt.Printf("Waiting for %s startup delay...\n", delay)
	time.Sleep(delay)
	svc = initialize()
	svc.Notifier.SendBroadcastMessage("launcher started")
	select {}
}

func initialize() *service.Service {
	cfg, err := config.New()
	if err != nil {
		log.Fatalf("Failed to create initialize configuration: %v", err)
	}

	params := service.Params{
		TGBotToken:       cfg.TGBotToken,
		TGForceChatID:    cfg.TGForceChatID,
		User:             cfg.User,
		ContainerName:    cfg.ContainerName,
		RestartPolicy:    cfg.RestartPolicy,
		EnvFilePath:      cfg.EnvFilePath,
		ServiceName:      cfg.ServiceName,
		Port:             cfg.Port,
		DockerAPIVersion: cfg.DockerAPIVersion,
		MetricsURI:       fmt.Sprintf("%s:%s", cfg.Host, cfg.Port),
	}

	var serviceInstaller installer.Installer
	if params.ServiceName != "" {
		serviceInstaller, err = installer.NewSystemd(installer.SystemdParams{
			ServiceName: params.ServiceName,
			User:        params.User,
		})
		if err != nil {
			log.Fatalf("Failed to create systemd service: %v", err)
		}
	} else {
		serviceInstaller = &installer.Dummy{}
	}
	if !serviceInstaller.IsInstalled() {
		if err := serviceInstaller.Install(); err != nil {
			log.Fatal(err)
		}
		fmt.Printf("Service was installed. For systemd case, "+
			"use 'journalctl -u %s -n 10 -f' command to follow log\n", params.ServiceName)
	}

	ctx := context.Background()
	return service.New(ctx, params)
}
