package graceful

import (
	"context"
	"errors"
	"log"
	"syscall"
	"testing"
	"time"

	"github.com/gofiber/fiber/v2"
	"github.com/stretchr/testify/assert"
)

func TestGraceful_Run(t *testing.T) {
	graceful := Init()

	graceful.SetMaxShutdownProcess(0)
	graceful.SetMaxShutdownTime(0)
	graceful.RegisterProcess(nil)
	graceful.RegisterShutdownProcess(nil)

	graceful.SetMaxShutdownTime(1 * time.Second)
	graceful.SetMaxShutdownProcess(1)

	graceful.RegisterProcess(func() error {
		return nil
	})

	app := fiber.New()

	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("Hello, World 👋!")
	})

	graceful.RegisterProcess(func() error {
		return app.Listen(":3000")
	})

	log.Println("test")

	graceful.RegisterShutdownProcessWithTag(func(ctx context.Context) error {
		log.Println("test 6")
		time.Sleep(1 * time.Second)
		return app.Shutdown()
	}, "test app fiber")

	log.Println("test 2")
	graceful.RegisterShutdownProcess(func(ctx context.Context) error {
		log.Println("test 7")
		time.Sleep(1 * time.Second)
		return errors.New("err")
	})

	log.Println("test 3")
	go func() {
		log.Println("test 4", syscall.Getpid())
		time.Sleep(1 * time.Second)
		_ = syscall.Kill(syscall.Getpid(), syscall.SIGINT)
		log.Println("test 5")

	}()

	err := graceful.Wait()

	assert.Nil(t, err)
}

func TestGraceful_EmptyShutdown(t *testing.T) {
	graceful := Init()

	graceful.RegisterProcess(func() error {
		return nil
	})

	go func() {
		time.Sleep(1 * time.Second)
		_ = syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	}()

	err := graceful.Wait()

	assert.Nil(t, err)
}
