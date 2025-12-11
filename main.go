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
func main(){
	fmt.Println("on some bullshit")
	langs, err := languages.Load("internal/languages/languages.json")
	if err != nil {
		log.Fatalf("failed to load languages: %v", err)
	}
	dc := docker.New()
	exec := executer.NewExecutor(dc, langs)
	ctx,cancel := context.WithCancel(context.Background())
	go func(){
		if err:=dc.DeleteZombieContainer(ctx); err != nil {
			log.Fatal(err)
		}
	}()
	server:=api.NewServer(exec)
	if err = server.StartServer(); err !=nil {
		cancel()
		log.Println("Error in server starting")
	}

}