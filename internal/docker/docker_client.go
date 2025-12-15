package docker

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"log"
	"os"
	"path/filepath"
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
			Cmd:   sb.ExecCmd,
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
	if err := c.d.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return resp, err
	}
	return resp, err
}

func (c *Client) Run(ctx context.Context, image string, sb sandbox.Config) (Result, error) {
	resp, err := c.CreateContainer(ctx, sb, image)
	if err != nil {
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
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			log.Println("zombie cleanup stopped")
			return nil

		case <-ticker.C:
			containers, err := c.d.ContainerList(ctx, container.ListOptions{
				All: true,
				Filters: filters.NewArgs(
					filters.Arg("status", "exited"),
					filters.Arg("label", "pool!=true"),
				),
			})
			if err != nil {
				log.Printf("container list failed: %v", err)
				continue
			}

			for _, ctr := range containers {
				err := c.d.ContainerRemove(ctx, ctr.ID, container.RemoveOptions{
					Force: true,
				})
				if err != nil {
					log.Printf("failed to remove %s: %v", ctr.ID[:12], err)
					continue
				}
				log.Printf("removed zombie container %s", ctr.ID[:12])
			}
		}
	}
}

func (c *Client) CopyFilesToContainer(ctx context.Context, containerID string, hostDir string) error {
	// Create a tar archive of the files in hostDir
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)

	// First, create the /run/code directory in the tar
	hdr := &tar.Header{
		Name:     "code/",
		Mode:     0755,
		Typeflag: tar.TypeDir,
	}
	if err := tw.WriteHeader(hdr); err != nil {
		return err
	}

	files, err := os.ReadDir(hostDir)
	if err != nil {
		return err
	}

	for _, file := range files {
		if file.IsDir() {
			continue
		}

		filePath := filepath.Join(hostDir, file.Name())
		content, err := os.ReadFile(filePath)
		if err != nil {
			return err
		}

		hdr := &tar.Header{
			Name: "code/" + file.Name(),
			Mode: 0644,
			Size: int64(len(content)),
		}

		if err := tw.WriteHeader(hdr); err != nil {
			return err
		}

		if _, err := tw.Write(content); err != nil {
			return err
		}
	}

	if err := tw.Close(); err != nil {
		return err
	}

	// Copy the tar archive to /run/ in the container (it will create /run/code/)
	return c.d.CopyToContainer(ctx, containerID, "/run", &buf, container.CopyToContainerOptions{})
}

func (c *Client) ExecInExistingContainer(
	ctx context.Context,
	containerID string,
	sb sandbox.Config,
) (Result, error) {

	execCtx, cancel := context.WithTimeout(ctx, sb.Timeout)
	defer cancel()

	execResp, err := c.d.ContainerExecCreate(execCtx, containerID, container.ExecOptions{
		Cmd:          append([]string{"sh", "-c"}, quoteCmd(sb.ExecCmd)...),
		AttachStdout: true,
		AttachStderr: true,
		AttachStdin:  true,
		Tty:          false,
		WorkingDir:   "/app",
	})
	if err != nil {
		return Result{}, err
	}

	attach, err := c.d.ContainerExecAttach(execCtx, execResp.ID, container.ExecStartOptions{})
	if err != nil {
		return Result{}, err
	}
	defer attach.Close()

	// write stdin
	go func() {
		if sb.Stdin != "" {
			attach.Conn.Write([]byte(sb.Stdin))
		}
		if closer, ok := attach.Conn.(interface{ CloseWrite() error }); ok {
			closer.CloseWrite()
		}
	}()

	var stdout, stderr bytes.Buffer
	_, err = stdcopy.StdCopy(&stdout, &stderr, attach.Reader)
	if err != nil {
		return Result{}, err
	}

	inspect, err := c.d.ContainerExecInspect(execCtx, execResp.ID)
	if err != nil {
		return Result{}, err
	}

	return Result{
		Stdout:   stdout.String(),
		Stderr:   stderr.String(),
		ExitCode: inspect.ExitCode,
	}, nil
}

func quoteCmd(cmd []string) []string {
	if len(cmd) == 0 {
		return cmd
	}
	// Build command string with PATH exported for sh
	cmdStr := "export PATH=/usr/local/sbin:/usr/local/bin:/usr/sbin:/usr/bin:/sbin:/bin:/usr/local/go/bin:/go/bin; "
	for i, part := range cmd {
		if i > 0 {
			cmdStr += " "
		}
		cmdStr += fmt.Sprintf("%q", part)
	}
	return []string{cmdStr}
}
