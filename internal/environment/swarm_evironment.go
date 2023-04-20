package environment

import (
	"log"

	"github.com/cplace/cssc-operator/internal/instance"
	"github.com/docker/docker/client"
)

// Swarm is a struct that implements the Environment interface for Docker Swarm.
type SwarmEnvironment struct {
	*BaseEnvironment
	client *client.Client
}

// NewSwarm creates a new Swarm instance with a Docker client.
func NewSwarm() *SwarmEnvironment {
	cli, err := client.NewClientWithOpts(client.FromEnv)
	if err != nil {
		log.Fatalf("Failed to create Docker client: %v", err)
	}
	return &SwarmEnvironment{client: cli}
}

// ApplyInstance applies the given instance configuration to the swarm cluster.
func (s *SwarmEnvironment) ApplyInstance(instance *instance.Instance) error {
	// TODO: implement logic for applying instance configuration to swarm cluster
	return nil
}

// DeleteInstance deletes the given instance from the swarm cluster.
func (s *SwarmEnvironment) DeleteInstance(instance *instance.Instance) error {
	// TODO: implement logic for deleting instance from swarm cluster
	return nil
}

// GetInstanceStatus returns the status of the given instance in the swarm cluster.
func (s *SwarmEnvironment) GetInstanceStatus(instance *instance.Instance) (string, error) {
	// TODO: implement logic for getting instance status from swarm cluster
	return "", nil
}

// GetEnvironmentStatus returns the status of the swarm cluster.
func (s *SwarmEnvironment) GetEnvironmentStatus() (string, error) {
	// TODO: implement logic for getting environment status from swarm cluster
	return "", nil
}
