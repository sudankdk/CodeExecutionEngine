package api

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	executer "github.com/sudankdk/ceev2/internal/Executer"
)

type Server struct {
	exec *executer.Executor
}

func NewServer(exec *executer.Executor) *Server {
	return &Server{exec: exec}
}

func (s *Server) StartServer() error {
	app := fiber.New()
	app.Use(cors.New(cors.Config{
		AllowCredentials: true,
		AllowOrigins:     "http://localhost:5173",
		AllowMethods:     "GET,POST,PUT,DELETE,OPTIONS",
		AllowHeaders:     "Origin, Content-Type, Accept, Authorization",
	}))

	s.setupRoutes(app)

	return app.Listen(":3000")
}
