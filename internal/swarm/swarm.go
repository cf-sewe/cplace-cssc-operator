package swarm

import (
	"context"
	"fmt"

	"github.com/docker/docker/client"
	"github.com/rs/zerolog"
)

// Swarm contains the Docker client and context.
type Swarm struct {
	// The Docker client is used to communicate with the Docker daemon.
	cli *client.Client

	// The context is used to cancel the Docker client.
	ctx *context.Context

	// The logger is used to log messages.
	logger *zerolog.Logger
}

// New creates a new Env from the given context.
// It connects to the Docker daemon and checks if it is part of a Swarm cluster.
// It returns an error if the connection fails or the node is not a Swarm manager.
func NewSwarm(ctx *context.Context, log *zerolog.Logger) (*Swarm, error) {
	log.Info().Msg("Initializing Swarm environment")
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		return nil, fmt.Errorf("failed to create Docker client: %w", err)
	}

	info, err := cli.Info(*ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get Docker info: %w", err)
	}

	if !info.Swarm.ControlAvailable {
		return nil, fmt.Errorf("node is not a Swarm manager")
	}

	return &Swarm{
		cli:    cli,
		ctx:    ctx,
		logger: log,
	}, nil
}

func (s *Swarm) Bogus() {
	s.logger.Info().Msg("Bogus")
}
