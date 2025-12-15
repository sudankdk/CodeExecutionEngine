package main

import (
	"context"
	"fmt"
	"log"

	executer "github.com/sudankdk/ceev2/internal/Executer"
	"github.com/sudankdk/ceev2/internal/api"
	"github.com/sudankdk/ceev2/internal/docker"
	"github.com/sudankdk/ceev2/internal/languages"
)

func main() {
	fmt.Println("on some bullshit")
	langs, err := languages.Load("internal/languages/languages.json")
	if err != nil {
		log.Fatalf("failed to load languages: %v", err)
	}
	dc := docker.New()
	pooling := docker.NewPoolManager(dc)
	exec := executer.NewExecutor(dc, langs, pooling)
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()
	pooling.PreWarm(ctx, langs)
	// Removed zombie cleanup for now
	server := api.NewServer(exec)
	if err = server.StartServer(); err != nil {
		cancel()
		log.Println("Error in server starting")
	}

}
