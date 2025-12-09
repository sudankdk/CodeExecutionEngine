package docker

import (
	"archive/tar"
	"bytes"
	"context"
	"fmt"
	"io"
	"log"
	"path/filepath"
	"strings"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
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
	Cli         *client.Client
	Mu          sync.Mutex
	Contd       *sync.Cond
	Image       RunnerImage
	Containers  map[string]*ConatainerInfo
	Pool        chan string // to borrow and return
	PoolSize    int
	Workspace   string
	BorrowTTL   time.Duration
	HealthFreq  time.Duration
	Stopch      chan struct{}
	Tmpfs       bool
	ExecTimeout time.Duration
}

// factory method user gareko
// pool manager create garna ka lagi
func NewPoolManager(poolsize int, imagename RunnerImage, ctx context.Context) (*PoolManager, error) {
	//client attribute use garera aauta naya instance can be created suruma
	client, err := client.NewClientWithOpts(client.FromEnv,
		client.WithAPIVersionNegotiation())
	if err != nil {
		return nil, err
	}
	pool := PoolManager{
		Cli:         client,
		Containers:  make(map[string]*ConatainerInfo),
		Pool:        make(chan string),
		Image:       imagename,
		PoolSize:    poolsize,
		Workspace:   "/workspace",
		Tmpfs:       true, // temp file system
		HealthFreq:  15 * time.Second,
		Stopch:      make(chan struct{}),
		BorrowTTL:   10 * time.Second,
		ExecTimeout: 10 * time.Second,
	}
	pool.Contd = sync.NewCond(&pool.Mu)

	if err := pool.EnsureImage(ctx, string(imagename)); err != nil {
		return nil, err
	}
	return &pool, nil
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

func (pm *PoolManager) StartPool(ctx context.Context) error {
	for i := 0; i < pm.PoolSize; i++ {
		// start the container
		if err := pm.StartContainer(ctx); err != nil {
			return err
		}
	}
	// start health checker maybe with go routine
	return nil
}

func (pm *PoolManager) StartContainer(ctx context.Context) error {
	config := &container.Config{
		Tty:        false,
		Image:      string(pm.Image),
		WorkingDir: pm.Workspace,
		Cmd:        []string{"sleep", "infinity"},
		Env:        []string{"RUNNER-1"},
	}
	hostConfig := &container.HostConfig{
		Privileged: true,
		AutoRemove: false,
		Resources: container.Resources{
			Memory:   512 * 1024 * 1024,
			NanoCPUs: 500_000_000,
		},
	}
	if pm.Tmpfs {
		hostConfig.Tmpfs = map[string]string{pm.Workspace: "rw"} // provide read write access to the mounted dir or file system whatevere
	}

	resp, err := pm.Cli.ContainerCreate(ctx, config, hostConfig, nil, nil, "")
	if err != nil {
		return err
	}
	if err := pm.Cli.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return err
	}
	pm.Mu.Lock()
	pm.Containers[resp.ID] = &ConatainerInfo{
		ID:        resp.ID,
		Image:     pm.Image,
		Busy:      false,
		LastCheck: time.Now(),
	}
	pm.Pool <- resp.ID
	log.Printf("started container %s for image %s", resp.ID[:12], pm.Image)
	return nil
}

// check conatiner health in frequent duration and restart the bad ones and also logs the bad ones
func (pm *PoolManager) Healthchecker() {

	ticker := time.NewTicker(pm.HealthFreq)
	defer ticker.Stop()
	for {
		select {
		case <-pm.Stopch:
			return

		case <-ticker.C:
			pm.Mu.Lock()
			for id, v := range pm.Containers {

				go func(id string, info *ConatainerInfo) {
					ctx, cancel := context.WithTimeout(context.Background(), 3*time.Second)
					defer cancel()
					_, err := pm.Cli.ContainerInspect(ctx, id)
					if err != nil {
						log.Printf("container %s unhealthy, restarting: %v", id[:12], err)

						// restart the container
						pm.RestartContainer(ctx, id)
					} else {
						pm.Mu.Lock()
						info.LastCheck = time.Now()
						pm.Mu.Unlock()
					}
				}(id, v)
			}
			pm.Mu.Unlock()
		}
	}
}

func (pm *PoolManager) RestartContainer(ctx context.Context, id string) {
	pm.Mu.Lock()
	if c, ok := pm.Containers[id]; ok {
		c.Busy = true
	}
	pm.Mu.Unlock()

	// first stop ani delete and recreate
	_ = pm.Cli.ContainerStop(ctx, id, container.StopOptions{})
	_ = pm.Cli.ContainerRemove(ctx, id, container.RemoveOptions{})
	pm.Mu.Lock()
	delete(pm.Containers, id)
	pm.Mu.Unlock()
	if err := pm.StartContainer(ctx); err != nil {
		log.Printf("failed to recreate container: %v", err)

	}
}

func (pm *PoolManager) ExecRunner(ctx context.Context, id string, files map[string][]byte, command []string, timeout time.Duration) (stdout, stderr []byte, exitCode int, err error) {
	//copy files to container
	if err := pm.CopyFilesToContainer(ctx, id, files, pm.Workspace); err != nil {
		return nil, nil, -1, err
	}
	execCtx, cancel := context.WithTimeout(ctx, timeout)

	defer cancel()
	//create exec
	execResp, err := pm.Cli.ContainerExecCreate(execCtx, id, container.ExecOptions{
		Cmd:          command,
		AttachStdin:  true,
		AttachStdout: true,
		WorkingDir:   pm.Workspace,
		Env:          []string{},
	})
	if err != nil {
		return nil, nil, -1, err
	}
	respAttach, err := pm.Cli.ContainerExecAttach(execCtx, execResp.ID, container.ExecStartOptions{})

	if err != nil {
		return nil, nil, -1, err
	}

	defer respAttach.Close()
	var outBuf, errBuf bytes.Buffer

	_, err = io.Copy(&outBuf, respAttach.Reader)
	if err != nil && !strings.Contains(err.Error(), "context deadline exceeded") {
		log.Println("exec copy error:", err)
	}
	inspect, err := pm.Cli.ContainerExecInspect(ctx, execResp.ID)
	if err != nil {
		return outBuf.Bytes(), errBuf.Bytes(), -1, err
	}

	// Docker doesn't separate stderr easily without stdcopy; for simplicity return combined output as stdout
	return outBuf.Bytes(), nil, inspect.ExitCode, nil
}

func (pm *PoolManager) CopyFilesToContainer(ctx context.Context, containerId string, files map[string][]byte, destDir string) error {
	var buf bytes.Buffer
	tw := tar.NewWriter(&buf)
	for name, content := range files {
		hdr := &tar.Header{
			Name: filepath.Join(destDir, name),
			Mode: 0644,
			Size: int64(len(content)),
		}
		if err := tw.WriteHeader(hdr); err != nil {
			tw.Close()
			return err
		}
		if _, err := tw.Write(content); err != nil {
			tw.Close()
			return err
		}
	}
	if err := tw.Close(); err != nil {
		return err
	}

	return pm.Cli.CopyToContainer(ctx, containerId, "/", &buf, container.CopyToContainerOptions{AllowOverwriteDirWithFile: true})

}

func (pm *PoolManager) CleanWorkspace(ctx context.Context, containerID string) error {
	cmd := []string{"bash", "-lc", fmt.Sprintf("rm -rf %s/*", pm.Workspace)}
	_, _, _, err := pm.ExecRunner(ctx, containerID, nil, cmd, 5*time.Second)
	return err
}

func (pm *PoolManager) Shutdown(ctx context.Context) {
	close(pm.Stopch)
	pm.Mu.Lock()
	ids := make([]string, 0, len(pm.Containers))
	for id := range pm.Containers {
		ids = append(ids, id)
	}
	pm.Mu.Unlock()
	for _, id := range ids {
		timeout := 2
		_ = pm.Cli.ContainerStop(ctx, id, container.StopOptions{Timeout: &timeout})
		_ = pm.Cli.ContainerRemove(ctx, id, container.RemoveOptions{Force: true})
	}
	pm.Cli.Close()
}
