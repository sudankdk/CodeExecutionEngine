package api

import (
	"github.com/gofiber/fiber/v2"
	executer "github.com/sudankdk/ceev2/internal/Executer"
)

type Server struct {
	exec   *executer.Executor
}

func NewServer(exec *executer.Executor) *Server {
	return &Server{exec: exec}
}

func(s *Server) StartServer() {
	app := fiber.New()

	s.setupRoutes(app)

	app.Listen(":3000")
}
