package docker

import (
	"context"
	"testing"

	"github.com/docker/docker/client"
	"github.com/stretchr/testify/assert"
)

// TestEnsureImageSuccess tests EnsureImage when the image already exists
func TestEnsureImageSuccess(t *testing.T) {
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.FromEnv, client.WithAPIVersionNegotiation())
	if err != nil {
		t.Skipf("Docker not available: %v", err)
	}
	defer cli.Close()

	pm := &PoolManager{
		Cli:   cli,
		Image: GoImage,
	}

	// Test with an image that should exist
	err = pm.EnsureImage(ctx, "alpine")
	// This will pass if docker is running and alpine exists or can be pulled
	assert.NoError(t, err, "EnsureImage should not error for existing image")
}
