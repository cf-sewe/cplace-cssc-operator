package main

import (
	"log"

	"github.com/cf-sewe/cplace-cssc-operator/internal/environment"
	"github.com/cf-sewe/cplace-cssc-operator/internal/routes"
	"github.com/joho/godotenv"
)

// load environment configuration from .env file
func loadEnv() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}
}

// main initializes the environment and starts the workers and routes
func main() {
	loadEnv()
	var env environment.Environment
	env = environment.New()
	workers.Init(env)
	routes.Init(env)
}
