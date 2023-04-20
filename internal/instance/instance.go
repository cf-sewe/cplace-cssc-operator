package instance

// Instance is an interface that defines the methods for managing cplace instances.
type Instance interface {
	GetStatus() string  // get the status of the instance
	GetName() string    // get the name of the instance
	GetVersion() string // get the version of the instance

	run()            // start the instance
	stop()           // stop the instance
	getName() string // get the name of the instance
}

type InstanceInfo struct {
	// Name is the name of the instance.
	Name string
	// Id is the id of the instance.
	Id string
	// Version is the cplace version of the instance.
	Version string
	// Status is the runtime status of the instance (e.g. "running", "stopped", "degraded")
	Status string
}

// InstanceFilter is a filter for instances.
type InstanceFilter struct {
	// Name is the name of the instance.
	Name string
	// Id is the id of the instance.
	Id string
	// Type is the type of the instance.
	Type string
	// Status is the runtime status of the instance
	Status string
}
