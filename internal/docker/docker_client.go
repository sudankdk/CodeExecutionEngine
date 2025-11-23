// package executer

// import (
// 	"context"
// 	"io"

// 	"github.com/docker/docker/api/types/container"
// 	"github.com/docker/docker/client"
// )

// type Client struct{
// 	cli *client.Client
// }

// type Result struct {
//     Stdout   string
//     Stderr   string
//     ExitCode int
// }

// func New() (*Client,error){
// 	cli,err:=client.NewClientWithOpts(client.FromEnv)
// 	if err != nil { return  nil,err}
// 	return &Client{cli: cli},nil
// }

// func (c *Client) createContainer(ctx context.Context, image string, sb Config) (string,error){
// 	resp,err:=c.cli.ContainerCreate(ctx,
// 	&container.Config{
// 		Image: image,
// 		Cmd: sb.Cmd,
// 		Tty: false,
// 	},
// 	&container.HostConfig{
// 		Resources: container.Resources{
// 			Memory: sb.Memory,
// 			NanoCPUs: sb.Cpu,
// 		},
// 		ReadonlyRootfs: true,
// 		Binds: sb.Binds,
// 	},
// 	nil,nil,"",
// 	)
// 	if err != nil {
//         return "", err
//     }
//     return resp.ID, nil
// }

// func (c *Client) runContainer(ctx context.Context,id string) (*Result,error){
// 	if err:=c.cli.ContainerStart(ctx,id,container.StartOptions{});err != nil {
// 		return nil,err
// 	}

// 	statusCh,errCh:=c.cli.ContainerWait(ctx,id,container.WaitConditionNotRunning)
// 	if err := <-errCh; err != nil {
// 		return nil, err
// 	}
// 	waitResp := <-statusCh

// 	logs,err:=c.cli.ContainerLogs(ctx,id,container.LogsOptions{
// 		ShowStdout: false,
// 		ShowStderr: true,
// 	})
// 	if err != nil {
// 		return nil,err
// 	}
// 	out,_:=io.ReadAll(logs)
// 	return &Result{
// 		Stdout: string(out),
// 		Stderr: "",
// 		ExitCode: int(waitResp.StatusCode),
// 	},nil

// }

// func (c *Client) cleanup(id string){
// 	_=c.cli.ContainerRemove(context.Background(),id,container.RemoveOptions{Force: true})
// }

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
	cli, _ := client.NewClientWithOpts(client.FromEnv,client.WithAPIVersionNegotiation())
	return &Client{d: cli}
}

type Result struct {
	Stdout   string
	Stderr   string
	ExitCode int
}

//check garnu paro image xa ki xaina 
//if image xa vane return nil
//yedi image xaina vane pull garnu paro
//ani nil return garnu parne xa
// func (c *Client) EnsureImage(ctx context.Context, image string) error {
//     res, err := c.d.ImageInspect(ctx, image)
//     if err == nil {
//         return nil // present xa kai garnu pardaina
//     }

//     // pull the image if not present
//     rc, err := c.d.ImagePull(ctx, image, dockerTypes.ImagePullOptions{})
//     if err != nil {
//         return err
//     }
//     defer rc.Close()

//     // consume the pull output to complete the pull
//     _, _ = io.Copy(io.Discard, rc)

//     _ = res
//     return nil
// }



// func (c *Client) Run(ctx context.Context,image string, sb sandbox.Config) (Result, error) {

//     resp, err := c.d.ContainerCreate(ctx,
//         &container.Config{
//             Image: image,
//             Cmd:   sb.Cmd, 
//         },
//         &container.HostConfig{
//             AutoRemove: true,
// 			Binds: sb.Binds,
//             Resources:     container.Resources{Memory: sb.Memory, NanoCPUs: sb.CPU},
// 			ReadonlyRootfs: sb.ReadonlyRootfs,
// 			NetworkMode:   "none",
//         },
//         nil, nil, "",
//     )
//     if err != nil {
//         return Result{}, err
//     }

//     if err := c.d.ContainerStart(ctx, resp.ID, container.StartOptions{}); err != nil {
//         return Result{}, err
//     }

//     statusCh, errCh := c.d.ContainerWait(ctx, resp.ID, container.WaitConditionNotRunning)

//     var waitResp container.WaitResponse

//     select {
//     case err := <-errCh:
//         if err != nil {
//             return Result{}, fmt.Errorf("container wait error: %w", err)
//         }
// 	timeout := time.Duration(30) * time.Second
// 	if timeout > 0 {
// 		timeout = time.Duration(timeout) * time.Second
// 	}
	
//     case waitResp = <-statusCh:
//         // ok
//     }

//     logReader, err := c.d.ContainerLogs(ctx, resp.ID, container.LogsOptions{
//         ShowStdout: true,
//         ShowStderr: true,
//     })
//     if err != nil {
//         return Result{}, err
//     }
//     defer logReader.Close()

//     var stdout, stderr bytes.Buffer
//     io.Copy(&stdout, logReader)

//     if waitResp.StatusCode != 0 {
//         stderr.WriteString(fmt.Sprintf("Container exited with code %d", waitResp.StatusCode))
//     }


//     return Result{
//         Stdout:   stdout.String(),
//         Stderr:   stderr.String(),
//         ExitCode: int(waitResp.StatusCode),
//     }, nil
// }


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
