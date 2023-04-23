package operator

import (
	"context"
	"fmt"

	"github.com/cf-sewe/cplace-cssc-operator/internal/swarm"
	"github.com/rs/zerolog"
)

type Operator struct {
	Log   *zerolog.Logger
	Swarm *swarm.Swarm
}

func NewOperator(ctx *context.Context, log *zerolog.Logger) (*Operator, error) {
	// Initialize the Swarm environment with the given context
	se, err := swarm.NewSwarm(ctx, log)
	if err != nil {
		return nil, fmt.Errorf("failed to initialize Swarm environment: %w", err)
	}

	// initialize the GIT environment

	return &Operator{
		Log:   log,
		Swarm: se,
	}, nil
}
