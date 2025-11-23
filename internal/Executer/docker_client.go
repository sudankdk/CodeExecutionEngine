package executer

import (
	"context"

	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"
)


type Client struct{
	cli *client.Client
}

type Result struct {
    Stdout   string
    Stderr   string
    ExitCode int
}

func New() (*Client,error){
	cli,err:=client.NewClientWithOpts(client.FromEnv)
	if err != nil { return  nil,err}
	return &Client{cli: cli},nil
}

func (c *Client) createContainer(ctx context.Context, image string, sb Config) (string,error){
	resp,err:=c.cli.ContainerCreate(ctx,
	&container.Config{
		Image: image,
		Cmd: sb.Cmd,
		Tty: false,
	},
	&container.HostConfig{
		Resources: container.Resources{
			Memory: sb.Memory,
			NanoCPUs: sb.Cpu,
		},
		ReadonlyRootfs: true,
		Binds: sb.Binds,
	},
	nil,nil,"",
	)
	if err != nil {
        return "", err
    }
    return resp.ID, nil
}

func (c *Client) runContainer(ctx context.Context,id string) (*Result,error){
	if err:=c.cli.ContainerStart(ctx,id,nil);err != nil {
		return nil,err
	}
	
}