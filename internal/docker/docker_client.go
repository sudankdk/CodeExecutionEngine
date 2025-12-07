package docker

import (
	"bytes"
	"context"
	"fmt"

	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
	"github.com/docker/docker/pkg/stdcopy"
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

func (c *Client) Run(ctx context.Context, image string, sb sandbox.Config) (Result, error) {
	resp, err := c.d.ContainerCreate(ctx,
		&container.Config{
			Image: image,
			Cmd:   sb.Cmd,
		},
		&container.HostConfig{
			AutoRemove:     false,
			Binds:          sb.Binds,
			Resources:      container.Resources{Memory: sb.Memory, NanoCPUs: sb.CPU},
			ReadonlyRootfs: sb.ReadonlyRootfs,
			NetworkMode:    "none",
		},
		nil, nil, "",
	)
	if err != nil {
		return Result{}, err
	}

	if err := c.d.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return Result{}, err
	}

	statusCh, errCh := c.d.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)

	var waitResp container.WaitResponse

	select {
	case err := <-errCh:
		if err != nil {
			return Result{}, fmt.Errorf("container wait error: %w", err)
		}
	case waitResp = <-statusCh:
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
