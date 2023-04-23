package config

import (
	"os"
)

const (
	// LogLevel defines the log level for the operator service
	LogLevel string = "CSSC_OPERATOR_LOG_LEVEL"

	// EnvName is the name of the environment (e.g. "test")
	EnvName string = "CSSC_OPERATOR_ENV_NAME"

	// EnvDomain is the domain of the environment (e.g. "test.cplace.cloud")
	EnvDomain string = "CSSC_OPERATOR_ENV_DOMAIN"

	// GitRepo is the GIT repository URL that contains the cplace instance definitions
	GitRepo string = "CSSC_OPERATOR_GIT_REPO"

	// GitBranch is the GIT branch name
	GitBranch string = "CSSC_OPERATOR_GIT_BRANCH"

	// GitUser is the GIT user name for authentication
	GitUser string = "CSSC_OPERATOR_GIT_USER"

	// GitPassword is the GIT password for authentication
	GitPassword string = "CSSC_OPERATOR_GIT_PASSWORD"
)

// Get returns the value of an OS environment variable or a default value if not set
func GetEnv(key, defaultValue string) string {
	value := os.Getenv(key)
	if value == "" {
		return defaultValue
	}
	return value
}
