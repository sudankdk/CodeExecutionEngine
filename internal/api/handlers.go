package api

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
	executer "github.com/sudankdk/ceev2/internal/Executer"
	"github.com/sudankdk/ceev2/internal/languages"
	"github.com/sudankdk/ceev2/internal/utils"
)

type ExecuteRequest struct {
	Language string `json:"language"`
	Code     string `json:"code"`
	Stdin    string `json:"stdin"`
	Timeout  int    `json:"timeout"`
}

func (s *Server) setupRoutes(app *fiber.App) {
	app.Post("/execute", s.executeHandler)
	app.Get("/", func(c *fiber.Ctx) error { return c.SendString("CEE Running") })
}

func (s *Server) executeHandler(c *fiber.Ctx) error {
	var req ExecuteRequest
	if err := c.BodyParser(&req); err != nil {
		return fiber.NewError(400, "Invalid request body")
	}

	lang, err := languages.Get(req.Language)
	if err != nil {
		return fiber.NewError(400, "Language not supported")
	}

	files, err := utils.Save(req.Code, req.Stdin, lang.Ext)
	if err != nil {
		return fiber.NewError(500, "Failed to save code")
	}
	defer utils.CleanupFiles(files.Dir)
	const maxTimeout = 20
	timeout := req.Timeout
	if timeout <= 0 || timeout > maxTimeout {
		timeout = maxTimeout
	}
	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(timeout)*time.Second)
	defer cancel()
	er := executer.Request{
		Language: req.Language,
		Code:     req.Code,
		Stdin:    req.Stdin,
		Timeout:  timeout,
	}
	result, err := s.exec.Run(ctx, er)
	if err != nil {
		return fiber.NewError(500, err.Error())
	}

	return c.JSON(result)
}
