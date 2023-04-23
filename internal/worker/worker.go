package worker

// The background worker is performing the following tasks:
// - checkout the GIT repository with the instance definitions
// - reconcile the instances with the current state of the cluster

import (
	"github.com/cf-sewe/cplace-cssc-operator/internal/environment"
	"github.com/cf-sewe/cplace-cssc-operator/internal/instance"
)

type Worker struct {
	workchan WorkerChannel
}

type WorkerChannel chan []instance.Instance

// Init initializes the worker
func Init(env environment.Environment) {
	// Create a channel to queue tasks
	Worker.workchan = make(WorkerChannel, 10)
	// Start a goroutine that listens for tasks on the channel and runs them
	go Worker.run()
}

// run is a goroutine that listens for tasks on the channel and runs them
func (w *Worker) run() {
	for instances := range Worker.workchan {
		// Run the reconcileInstances function with the instances parameter
		err := reconcileInstances(instances)
		if err != nil {
			// Handle error
		}
	}
}

// AddWork adds work to the worker
func AddWork(instances []instance.Instance) error {
	// Check to make sure that the worker is initialized
	if Worker.workchan == nil {
		// Return error
	}

	// Add work to the channel
	Worker.workchan <- instances

	return nil
}

// Shutdown shuts down the worker
func Shutdown() {
	// Close the channel when done
	close(Worker.workchan)
}
