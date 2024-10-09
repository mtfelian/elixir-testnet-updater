package service

import (
	"context"
	"log"
	"time"

	"github.com/mtfelian/elixir-testnet-updater/delixir"
	"github.com/mtfelian/elixir-testnet-updater/metrics"
	"github.com/mtfelian/elixir-testnet-updater/notifier"
	"github.com/robfig/cron/v3"
)

// Service represents service capabilities
type Service struct {
	Notifier     notifier.Notifier
	DockerClient *delixir.DockerClient
	Metrics      *metrics.Metrics
}

// Params represents service parameters
type Params struct {
	TGBotToken    string
	TGForceChatID int64

	User             string
	ContainerName    string
	RestartPolicy    string
	EnvFilePath      string
	ServiceName      string
	Port             string
	DockerAPIVersion string
	MetricsURI       string
}

// New initializes new service instance
func New(ctx context.Context, p Params) *Service {
	service := new(Service)

	envVars, envConfig, err := delixir.ParseEnvFile(p.EnvFilePath)
	if err != nil {
		log.Fatalf("Failed to parse env file: %v", err)
	}

	if p.TGBotToken != "" {
		if service.Notifier, err = notifier.NewTGBot(notifier.TGBotParams{
			BotToken:    p.TGBotToken,
			ForceChatID: p.TGForceChatID,
			InstanceID:  envConfig.DisplayName,
		}); err != nil {
			log.Fatalf("Failed to init TG bot: %v", err)
		}
	} else {
		service.Notifier = &notifier.Dummy{}
	}

	service.Metrics = metrics.New(metrics.Params{
		URI:      p.MetricsURI,
		Notifier: service.Notifier,
	})

	if service.DockerClient, err = delixir.NewDockerClient(delixir.DockerClientParams{
		EnvVars:       envVars,
		Notifier:      service.Notifier,
		APIVersion:    p.DockerAPIVersion,
		ContainerName: p.ContainerName,
		Port:          p.Port,
		RestartPolicy: p.RestartPolicy,
	}); err != nil {
		log.Fatalf("Failed to create Docker DockerClient: %v", err)
	}

	service.DockerClient.CheckAndUpdateContainer(ctx) // check once first
	service.startPeriodicUpdates(ctx)

	return service
}

func (s *Service) startPeriodicUpdates(ctx context.Context) {
	c := cron.New()

	if _, err := c.AddFunc("0 * * * *", func() { // every hour at minute 0
		log.Printf("Checking for updates at %s...", time.Now().Format(time.RFC1123))
		s.DockerClient.CheckAndUpdateContainer(ctx)
	}); err != nil {
		log.Fatalf("Failed to add periodic task: %v", err)
	}

	if _, err := c.AddFunc("*/5 * * * *", func() {
		s.Metrics.Update()
	}); err != nil {
		log.Fatalf("Failed to add metrics changed detection periodic task: %v", err)
	}

	c.Start()
}
