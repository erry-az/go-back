package server

import (
	"context"

	"github.com/rs/zerolog/log"

	"github.com/erry-az/go-back/pkg/graceful"
	"github.com/gofiber/fiber/v2"
)

type RESTConfig struct {
	AppName    string
	Port       string
	OnShutdown map[string]func(context.Context) error
}

type RestRoute func(*fiber.App)

func StartREST(cfg *RESTConfig, routes ...RestRoute) {
	app := fiber.New(fiber.Config{
		AppName: cfg.AppName,
	})

	for _, route := range routes {
		route(app)
	}

	watcher := graceful.Init()

	watcher.RegisterProcess(func() error {
		return app.Listen(cfg.Port)
	})

	watcher.RegisterShutdownProcessWithTag(func(ctx context.Context) error {
		return app.Shutdown()
	}, "fiber")

	for tag, process := range cfg.OnShutdown {
		watcher.RegisterShutdownProcessWithTag(process, tag)
	}

	err := watcher.Wait()
	if err != nil {
		log.Error().Stack().Err(err).Send()
	}
}
