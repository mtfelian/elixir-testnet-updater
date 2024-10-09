package delixir

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
	"github.com/docker/go-connections/nat"
	"github.com/mtfelian/elixir-testnet-updater/notifier"
)

const imageName = "elixirprotocol/validator:v3" // Replace with your image name

// DockerClientParams represents docker client parameters
type DockerClientParams struct {
	EnvVars       []string
	Notifier      notifier.Notifier
	APIVersion    string
	ContainerName string
	Port          string
	RestartPolicy string
}

// NewDockerClient creates new Docker client
func NewDockerClient(p DockerClientParams) (*DockerClient, error) {
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithVersion(p.APIVersion))
	if err != nil {
		log.Fatalf("Failed to create Docker client: %v", err)
	}
	return &DockerClient{
		cli:           cli,
		envVars:       p.EnvVars,
		notifier:      p.Notifier,
		containerName: p.ContainerName,
		port:          p.Port,
		restartPolicy: p.RestartPolicy,
	}, nil
}

// DockerClient represents docker client
type DockerClient struct {
	cli           *client.Client
	envVars       []string
	notifier      notifier.Notifier
	containerName string
	port          string
	restartPolicy string
}

func (dc *DockerClient) pullLatestImage(ctx context.Context) error {
	reader, err := dc.cli.ImagePull(ctx, imageName, image.PullOptions{Platform: "linux/amd64"})
	if err != nil {
		return err
	}
	defer reader.Close()

	// Parse the JSON output and print the progress
	decoder := json.NewDecoder(reader)
	for {
		var msg struct {
			Status   string `json:"status"`
			Progress string `json:"progress"`
			ID       string `json:"id"`
		}

		if err := decoder.Decode(&msg); err == io.EOF {
			break
		} else if err != nil {
			return err
		}

		// Print status and progress if available
		if msg.ID != "" {
			fmt.Printf("%s: %s %s\n", msg.ID, msg.Status, msg.Progress)
		} else {
			fmt.Printf("%s\n", msg.Status)
		}
	}

	return nil
}

// CheckAndUpdateContainer image
func (dc *DockerClient) CheckAndUpdateContainer(ctx context.Context) {
	currentImageID, err := dc.getCurrentImageID(ctx)
	if err != nil {
		log.Printf("Error getting current image ID: %v", err)
	}

	fmt.Println("Pulling the latest image...")
	if err := dc.pullLatestImage(ctx); err != nil {
		log.Printf("Error pulling image: %v", err)
		return
	}

	newImageID, err := dc.getImageID(ctx)
	if err != nil {
		log.Printf("Error getting new image ID: %v", err)
		return
	}

	if currentImageID != newImageID {
		fmt.Println("New image found, updating container...")
		dc.updateContainer(ctx)
		dc.notifier.SendBroadcastMessage(fmt.Sprintf("updated image from %q to %q", currentImageID, newImageID))
	} else {
		fmt.Println("Container is already up to date.")
		//dc.notifier.SendBroadcastMessage("image is up-to-date")
	}
}

// containerExists returns containerID if it exists, otherwise returns empty string
func (dc *DockerClient) containerExists(ctx context.Context) (string, error) {
	containers, err := dc.cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return "", err
	}

	for _, cont := range containers {
		for _, name := range cont.Names {
			if name == "/"+dc.containerName {
				return cont.ID, nil
			}
		}
	}

	return "", nil
}

func (dc *DockerClient) getCurrentImageID(ctx context.Context) (string, error) {
	containers, err := dc.cli.ContainerList(ctx, container.ListOptions{All: true})
	if err != nil {
		return "", err
	}

	for _, cont := range containers {
		for _, name := range cont.Names {
			if name == "/"+dc.containerName {
				return cont.ImageID, nil
			}
		}
	}

	return "", fmt.Errorf("container %s not found", dc.containerName)
}

func (dc *DockerClient) getImageID(ctx context.Context) (string, error) {
	images, err := dc.cli.ImageList(ctx, image.ListOptions{All: true})
	if err != nil {
		return "", err
	}

	// Find the image by name and get its ID
	for _, img := range images {
		for _, tag := range img.RepoTags {
			fmt.Println(">>>", tag)
			if tag == imageName {
				return img.ID, nil
			}
		}
	}

	return "", fmt.Errorf("image %s not found", imageName)
}

func (dc *DockerClient) updateContainer(ctx context.Context) {
	fmt.Println("checking container existence...")
	containerID, err := dc.containerExists(ctx)
	if err != nil {
		log.Printf("Error checking container for existence: %v", err)
		return
	}

	if containerID != "" {
		fmt.Println("Stopping the container...")
		if err := dc.cli.ContainerStop(ctx, containerID, container.StopOptions{}); err != nil {
			log.Printf("Error stopping container: %v", err)
			return
		}

		fmt.Println("Removing the container...")
		if err := dc.cli.ContainerRemove(ctx, containerID, container.RemoveOptions{}); err != nil {
			log.Printf("Error removing container: %v", err)
			return
		}
	}

	natPort, err := nat.NewPort("tcp", dc.port)
	if err != nil {
		log.Printf("Error creating NatPort: %v", err)
		return
	}
	fmt.Println("Starting a new container with the updated image...")
	resp, err := dc.cli.ContainerCreate(ctx, &container.Config{
		Image: imageName,
		Env:   dc.envVars,
		ExposedPorts: nat.PortSet{
			natPort: struct{}{},
		},
	}, &container.HostConfig{
		PortBindings: nat.PortMap{
			natPort: []nat.PortBinding{
				{
					HostIP:   "0.0.0.0",
					HostPort: dc.port,
				},
			},
		},
		RestartPolicy: container.RestartPolicy{
			Name: container.RestartPolicyMode(dc.restartPolicy),
		},
	}, nil, nil, dc.containerName)
	if err != nil {
		log.Printf("Error creating container: %v", err)
		return
	}

	if err := dc.cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		log.Printf("Error starting container: %v", err)
		return
	}

	fmt.Println("Container updated successfully.")
}
