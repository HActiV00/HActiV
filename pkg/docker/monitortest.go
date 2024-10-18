package docker

import (
	"context"
	"fmt"
	"log"
	"os"
	"os/signal"
	"syscall"

	"github.com/docker/docker/api/types/events"
	"github.com/docker/docker/client"
)

func MonitorDockerEvents() {
	sig := make(chan os.Signal, 1)
	signal.Notify(sig, os.Interrupt, syscall.SIGTERM)
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		log.Fatal(err)
	}

	msgChan, errChan := cli.Events(context.Background(), events.ListOptions{})

	for {
		select {
		case event := <-msgChan:
			if event.Type == events.ContainerEventType {
				switch event.Action {
				case "start", "die":

					SetContainer()
				}
			}
		case err := <-errChan:
			fmt.Println("1", err)
		}
	}
}
