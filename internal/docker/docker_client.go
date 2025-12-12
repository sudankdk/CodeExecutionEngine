package docker

import (
	"bytes"
	"context"
	"fmt"
	"log"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/api/types/filters"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
	"github.com/docker/go-units"
	"github.com/sudankdk/ceev2/internal/sandbox"
)

type Client struct {
	d *client.Client
}

func New() *Client {
	cli, _ := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	return &Client{d: cli}
}

type Result struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

func (c *Client) CreateContainer(ctx context.Context, sb sandbox.Config, image string) (container.CreateResponse, error) {
	resp, err := c.d.ContainerCreate(ctx,
		&container.Config{
			Image: image,
			Cmd:   sb.Cmd,
		},
		&container.HostConfig{
			AutoRemove: false,
			Binds:      sb.Binds,
			Resources: container.Resources{
				Memory:   sb.Memory,
				NanoCPUs: sb.CPU,
				Ulimits: []*units.Ulimit{
					{
						Name: "nproc",
						Soft: 64,
						Hard: 128,
					},
					{
						Name: "nofile",
						Soft: 64,
						Hard: 128,
					},
					{
						Name: "core",
						Soft: 0,
						Hard: 0,
					},
					{
						// Maximum file size that can be created by the process (output file in our case)
						Name: "fsize",
						Soft: 20 * 1024 * 1024,
						Hard: 20 * 1024 * 1024,
					},
				},
			},
			ReadonlyRootfs: sb.ReadonlyRootfs,
			NetworkMode:    "none",
		},
		nil, nil, "",
	)
	return resp, err
}

func (c *Client) Run(ctx context.Context, image string, sb sandbox.Config) (Result, error) {
	resp, err := c.CreateContainer(ctx, sb, image)
	if err != nil {
		return Result{}, err
	}

	if err := c.d.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return Result{}, err
	}
	execCtx, cancel := context.WithTimeout(ctx, sb.Timeout)
	defer cancel()
	statusCh, errCh := c.d.ContainerWait(execCtx, resp.ID, container.WaitConditionNotRunning)

	var waitResp container.WaitResponse

	select {
	case err := <-errCh:
		if err != nil {
			return Result{}, fmt.Errorf("container wait error: %w", err)
		}
	case waitResp = <-statusCh:
	case <-execCtx.Done():
		c.d.ContainerKill(ctx, resp.ID, "SIGKILL")
		return Result{}, fmt.Errorf("execution timed out after %s", sb.Timeout)
	}

	logReader, err := c.d.ContainerLogs(ctx, resp.ID, container.LogsOptions{
		ShowStdout: true,
		ShowStderr: true,
		Follow:     false,
	})
	if err != nil {
		return Result{}, err
	}
	defer logReader.Close()

	var stdout, stderr bytes.Buffer
	_, err = stdcopy.StdCopy(&stdout, &stderr, logReader)
	if err != nil {
		return Result{}, err
	}

	if waitResp.StatusCode != 0 && stderr.Len() == 0 {
		stderr.WriteString(fmt.Sprintf("Container exited with code %d", waitResp.StatusCode))
	}

	return Result{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: int(waitResp.StatusCode),
	}, nil
}

func (c *Client) DeleteZombieContainer(ctx context.Context) error {
	ticker := time.NewTicker(50 * time.Second)
	for {
		select {
		case <-ctx.Done():
			log.Println("stopping the zombie cleanup process")
			return nil

		case <-ticker.C:
			f := filters.NewArgs()
			f.Add("label", "pool!=true")

			report, err := c.d.ContainersPrune(ctx, f)
			if err != nil {
				log.Println("failed to prune containers")
				return err
			}
			log.Printf("prune successful, pruned %v containers", len(report.ContainersDeleted))
		}
	}
}
