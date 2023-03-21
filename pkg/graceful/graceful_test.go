package graceful

import (
	"context"
	"errors"
	"syscall"
	"testing"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/stretchr/testify/assert"
)

func TestGraceful_Run(t *testing.T) {
	graceful := New()

	graceful.SetMaxShutdownProcess(0)
	graceful.SetMaxShutdownTime(0)
	graceful.RegisterProcess(nil)
	graceful.RegisterShutdownProcess(nil)

	graceful.SetMaxShutdownTime(1 * time.Second)
	graceful.SetMaxShutdownProcess(1)
	graceful.SetCancelOnError(false)

	graceful.RegisterProcess(func() error {
		return nil
	})

	graceful.RegisterProcess(func() error {
		log.Info().Msg("start new process")
		return nil
	})

	graceful.RegisterShutdownProcessWithTag(func(ctx context.Context) error {
		log.Info().Msg("stop new process")
		time.Sleep(1 * time.Second)
		return nil
	}, "test app fiber")

	graceful.RegisterShutdownProcess(func(ctx context.Context) error {
		time.Sleep(1 * time.Second)
		return errors.New("err")
	})

	go func() {
		time.Sleep(1 * time.Second)
		_ = syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	}()

	err := graceful.Wait()

	assert.Nil(t, err)
}

func TestGraceful_EmptyShutdown(t *testing.T) {
	graceful := New()

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

func TestGraceful_CancelOnError(t *testing.T) {
	graceful := New()
	graceful.SetCancelOnError(true)

	graceful.RegisterProcess(func() error {
		return nil
	})

	graceful.RegisterShutdownProcess(func(ctx context.Context) error {
		time.Sleep(1 * time.Second)
		return errors.New("err")
	})

	go func() {
		time.Sleep(1 * time.Second)
		_ = syscall.Kill(syscall.Getpid(), syscall.SIGINT)
	}()

	err := graceful.Wait()

	assert.NotNil(t, err)
}
