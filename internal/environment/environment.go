package environment

import (
	"log"
	"os"
	"time"

	"github.com/cf-sewe/cplace-cssc-operator/internal/instance"
)

// Environment is an interface that defines the methods for managing cplace instances in a specific environment type.
type Environment interface {
	// GetEnvironmentInfo returns the information about the environment as per configuration.
	GetEnvironmentInfo() (EnvironmentInfo, error)
	// GetEnvironmentStatus returns the status of the environment.
	GetEnvironmentStatus() (EnvironmentStatus, error)
	// GetEnvironmentMetrics returns the metrics for the environment.
	GetEnvironmentMetrics() (EnvironmentMetrics, error)

	// GetInstances returns a list of all instances in the environment.
	GetInstances() ([]instance.Instance, error)
	// GetInstancesWithFilter returns a list of all instances in the environment that match the given filter.
	GetInstancesWithFilter(filter instance.InstanceFilter) ([]instance.Instance, error)

	// GetInstanceByName returns the instance with the given name.
	GetInstanceByName(name string) (*instance.Instance, error)
	// GetInstanceById returns the instance with the given id.
	GetInstanceById(id string) (*instance.Instance, error)

	// ApplyInstance applies the given instance configuration to the cluster.
	ApplyInstance(instance *instance.Instance) error
	// DeleteInstance deletes the given instance from the cluster.
	DeleteInstance(instance *instance.Instance) error
}

// EnvironmentInfo represents the information about the environment as per configuration.
type EnvironmentInfo struct {
	// Type is the type of the environment (e.g. "swarm", "nomad")
	Type string
	// Name is the environment specifier, e.g. "test", "stag", "prod"
	Name string
	// Domain is the base domain of the environment where the instances are reachable
	Domain string
}

// EnvironmentStatus represents the status of an environment.
type EnvironmentStatus struct {
	// Status is the status of the environment in the cluster.
	// Possible values are: "running", "degraded", "maintenance"
	Status string
	// Message is a human readable message about the status.
	Message string
	// LastChangedAt is the time when the status last changed.
	LastChangedAt time.Time
}

// EnvironmentMetrics contains the metrics for the environment.
type EnvironmentMetrics struct {
	Uptime        time.Duration
	InstanceCount int
	NodeCount     int
	Capacity      struct {
		Memory struct {
			Total int
			Used  int
		}
		Disk struct {
			Total int
			Used  int
		}
		Cpu struct {
			Total int
			Used  int
		}
	}
}

func New() Environment {
	var env Environment
	envType := os.Getenv("CSSC_OPERATOR_ENVIRONMENT_TYPE")
	switch envType {
	case "swarm":
		env = NewSwarm()
	default:
		log.Fatalf("Unsupported environment type: %s", envType)
	}
	return env
}
