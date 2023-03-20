package main

import (
	"context"
	"errors"

	"github.com/erry-az/go-back/internal/server"
	"github.com/gofiber/fiber/v2"
	"github.com/rs/zerolog/log"
)

func main() {
	server.StartREST(&server.RESTConfig{
		AppName: "go-back-rest",
		Port:    ":3000",
		OnShutdown: map[string]func(context.Context) error{
			"test": func(ctx context.Context) error {
				return errors.New("err1")
			},
			"test 2": func(ctx context.Context) error {
				log.Info().Msg("test 2 called")
				defer log.Info().Msg("sccess")
				return errors.New("test err")
			},
		},
	}, func(app *fiber.App) {
		app.Get("/", func(ctx *fiber.Ctx) error {
			return ctx.Send([]byte("test"))
		})
	})
}
