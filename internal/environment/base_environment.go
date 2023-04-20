package environment

import (
	"github.com/cf-sewe/cplace-cssc-operator/internal/instance"
	"github.com/cf-sewe/cplace-cssc-operator/pkg/errors"
)

// import errors

// BaseEnvironment implements the common methods for all environments.
type BaseEnvironment struct {
	// some fields
}

// GetInstances returns a list of all instances in the environment.
func (e *BaseEnvironment) GetInstances() ([]instance.Instance, error) {
	// return error that the method must be implemented for the specific environment type
	return nil, errors.ErrNotImplemented
}

// GetInstancesWithFilter returns a list of all instances in the environment that match the given filter.
func (e *BaseEnvironment) GetInstancesWithFilter(filter instance.InstanceFilter) []instance.Instance {
	// Get all instances from the environment
	instances, _ := e.GetInstances()
	// Create an empty list to store the filtered instances
	filtered := []instance.Instance{}
	// Loop over the instances and check if they match the filter
	for _, instance := range instances {
		// Check if the instance status matches the filter status
		if filter.Status != "" && instance.GetStatus() != filter.Status {
			continue // Skip this instance
		}
		// Check if the instance name matches the filter name
		if filter.Name != "" && instance.GetName() != filter.Name {
			continue // Skip this instance
		}
		// Check if the instance version matches the filter version
		if filter.Version != "" && instance.GetVersion() != filter.Version {
			continue // Skip this instance
		}
		// If all checks pass, append the instance to the filtered list
		filtered = append(filtered, instance)
	}
	// Return the filtered list
	return filtered
}
