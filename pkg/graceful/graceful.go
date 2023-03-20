package graceful

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/rs/zerolog/log"
	"golang.org/x/sync/errgroup"
)

const (
	defaultMaxShutdownTime    = 10 * time.Second
	defaultMaxShutdownProcess = 5
)

type Graceful struct {
	groupCtx, signalCtx context.Context
	signalCancel        context.CancelFunc
	group               *errgroup.Group
	shutdownProcess     []func(ctx context.Context) error
	shutdownTags        []string
	maxShutdownTime     time.Duration
	maxShutdownProcess  int
}

func Init() *Graceful {
	return InitContext(context.Background())
}

func InitContext(ctx context.Context) *Graceful {
	var (
		signalCtx, signalCancel = signal.NotifyContext(ctx, os.Interrupt,
			syscall.SIGINT, syscall.SIGTERM, syscall.SIGHUP)
		group, groupCtx = errgroup.WithContext(signalCtx)
	)

	return &Graceful{
		groupCtx:           groupCtx,
		signalCtx:          signalCtx,
		signalCancel:       signalCancel,
		group:              group,
		shutdownProcess:    make([]func(ctx context.Context) error, 0),
		shutdownTags:       make([]string, 0),
		maxShutdownTime:    defaultMaxShutdownTime,
		maxShutdownProcess: defaultMaxShutdownProcess,
	}
}

func (g *Graceful) SetMaxShutdownTime(duration time.Duration) {
	if duration < 1 {
		g.maxShutdownTime = defaultMaxShutdownTime

		return
	}

	g.maxShutdownTime = duration
}

func (g *Graceful) SetMaxShutdownProcess(max int) {
	if max < 1 {
		g.maxShutdownProcess = defaultMaxShutdownProcess

		return
	}

	g.maxShutdownProcess = max
}

func (g *Graceful) RegisterProcess(process func() error) {
	if process == nil {
		return
	}

	g.group.Go(process)
}

func (g *Graceful) RegisterShutdownProcess(process func(context.Context) error) {
	g.RegisterShutdownProcessWithTag(process, "")
}

func (g *Graceful) RegisterShutdownProcessWithTag(process func(context.Context) error, tag string) {
	if process == nil {
		return
	}

	g.shutdownTags = append(g.shutdownTags, tag)
	g.shutdownProcess = append(g.shutdownProcess, process)
}

func (g *Graceful) createTag(i int) string {
	tag := g.shutdownTags[i]
	if tag == "" {
		tag = fmt.Sprintf("process %d", i)
	}

	return tag
}

func (g *Graceful) shutdown() error {
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), g.maxShutdownTime)
	defer shutdownCancel()

	shutdownGroup, shutdownGroupCtx := errgroup.WithContext(shutdownCtx)
	shutdownGroup.SetLimit(g.maxShutdownProcess)

	for i, process := range g.shutdownProcess {
		var (
			iCopy       = i
			processCopy = process
		)

		shutdownGroup.Go(func() error {
			tag := g.createTag(iCopy)

			err := processCopy(shutdownGroupCtx)
			if err != nil {
				log.Error().Str("tag", tag).Err(err).Send()
			} else {
				log.Info().Str("tag", tag).Msg("success")
			}

			return nil
		})
	}

	return shutdownGroup.Wait()
}

func (g *Graceful) Wait() error {
	defer g.signalCancel()

	g.group.Go(func() error {
		<-g.groupCtx.Done()

		if len(g.shutdownProcess) > 0 {
			return g.shutdown()
		}

		return nil
	})

	return g.group.Wait()
}
