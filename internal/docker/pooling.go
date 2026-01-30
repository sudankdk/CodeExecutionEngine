package docker

import (
	"context"
	"sync"
	"time"

	"github.com/docker/docker/api/types/container"
	"github.com/sudankdk/ceev2/internal/languages"
)

type PoolManager struct {
	client *Client
	pool   map[string]*PoolContainer
	mu     sync.Mutex
}

type PoolContainer struct {
	ID       string
	Status   string
	LastUsed time.Time
	Image    string
}

func NewPoolManager(c *Client) *PoolManager {
	return &PoolManager{
		client: c,
		pool:   make(map[string]*PoolContainer),
	}
}

func (pm *PoolManager) createPoolContainer(ctx context.Context, image string) (string, error) {
	resp, err := pm.client.d.ContainerCreate(
		ctx,
		&container.Config{
			Image: image,
			Cmd:   []string{"tail", "-f", "/dev/null"},
			Labels: map[string]string{
				"pool":       "true",
				"pool.image": image,
			},
			Entrypoint: []string{},
		},
		&container.HostConfig{
			AutoRemove:  false,
			NetworkMode: "none",
		},
		nil, nil, "",
	)
	if err != nil {
		return "", err
	}

	if err := pm.client.d.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
		return "", err
	}

	return resp.ID, nil
}

func (pm *PoolManager) PreWarm(ctx context.Context, langs languages.LanguageMap) error {

	for _, lang := range langs {

		for i := 0; i < 1; i++ {
			id, err := pm.createPoolContainer(ctx, lang.Image)
			if err != nil {
				return err
			}

			pm.mu.Lock()
			pm.pool[id] = &PoolContainer{
				ID:       id,
				Status:   "idle",
				LastUsed: time.Now(),
				Image:    lang.Image,
			}
			pm.mu.Unlock()
		}
	}
	return nil
}

func (pm *PoolManager) Acquire(image string) *PoolContainer {
	pm.mu.Lock()
	defer pm.mu.Unlock()

	for _, cont := range pm.pool {
		if cont.Status == "idle" && cont.Image == image {
			cont.Status = "busy"
			return cont
		}
	}
	return nil
}

func (pm *PoolManager) Release(cont *PoolContainer) {
	pm.mu.Lock()
	cont.Status = "idle"
	cont.LastUsed = time.Now()
	pm.mu.Unlock()
}
