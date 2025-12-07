package docker

import (
	"context"
	"fmt"
	"io"
	"log"
	"sync"
	"time"

	"github.com/docker/docker/api/types/image"
	"github.com/docker/docker/client"
)

type RunnerImage string

const (
	GoImage     RunnerImage = "cee-go"
	NodeImage   RunnerImage = "cee-node"
	PythonImage RunnerImage = "cee-py"
)

type ConatainerInfo struct {
	ID        string
	Image     RunnerImage
	Busy      bool
	LastCheck time.Time
}

type PoolManager struct {
	ID    string
	Cli   *client.Client
	mu    sync.Mutex
	Image RunnerImage
}

func (pm *PoolManager) EnsureImage(ctx context.Context, imageName string) error {
	resp, err := pm.Cli.ImageInspect(ctx, imageName)
	if err == nil {
		fmt.Printf("image found: %s\n", resp.ID)
		return nil
	}
	log.Println("pulling the image")

	out, err := pm.Cli.ImagePull(ctx, imageName, image.PullOptions{})
	if err != nil {
		return err
	}
	defer out.Close()
	io.Copy(io.Discard, out)
	return nil
}
