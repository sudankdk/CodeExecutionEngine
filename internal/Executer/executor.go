package executer

import (
	"context"
	"errors"
	"path/filepath"

	"github.com/sudankdk/ceev2/internal/docker"
	"github.com/sudankdk/ceev2/internal/languages"
	"github.com/sudankdk/ceev2/internal/sandbox"
	"github.com/sudankdk/ceev2/internal/utils"
)

type Request struct {
	Language string `json:"language"`
	Code     string `json:"code"`
	Stdin    string `json:"stdin"`
	Timeout  int    `json:"timeout"` // seconds
}

type Response struct {
	Stdout   string `json:"stdout"`
	Stderr   string `json:"stderr"`
	ExitCode int    `json:"exit_code"`
}

type Executor struct {
	docker  *docker.Client
	pooling *docker.PoolManager
	langs   languages.LanguageMap
}

func NewExecutor(d *docker.Client, lang languages.LanguageMap, p *docker.PoolManager) *Executor {
	return &Executor{docker: d, langs: lang, pooling: p}
}

func (e *Executor) Run(ctx context.Context, req Request) (*Response, error) {
	langCfg, ok := e.langs[req.Language]
	if !ok {
		return nil, errors.New("unsupported language")
	}
	pc := e.pooling.Acquire(langCfg.Image)
	defer e.pooling.Release(pc)

	files, err := utils.Save(req.Code, req.Stdin, langCfg.Ext)
	if err != nil {
		return nil, err
	}
	defer utils.CleanupFiles(files.Dir)

	//enuser image presents
	codeFileName := "/run/code/" + filepath.Base(files.CodePath)
	stdInFileName := "/run/code/" + filepath.Base(files.StdinPath)
	// codeFileName := "/run/code/main.py"
	// stdInFileName := "/run/code/stdin.txt"

	sb := sandbox.NewConfig(files.Dir, codeFileName, stdInFileName, req.Language)

	// Copy files to the container
	if err := e.docker.CopyFilesToContainer(ctx, pc.ID, files.Dir); err != nil {
		return nil, err
	}

	// Run container using your Docker client
	// res, err := e.docker.Run(ctx, langCfg.Image, sb)
	// if err != nil {
	// 	return nil, err
	// }

	res, err := e.docker.ExecInExistingContainer(ctx, pc.ID, sb)
	if err != nil {
		return nil, err
	}
	return &Response{
		Stdout:   res.Stdout,
		Stderr:   res.Stderr,
		ExitCode: res.ExitCode,
	}, nil
}
