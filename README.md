# clibase

Clibase provides the basic features and functionality common to all of my Cobra Command based Golang projects.

# Features
* Generate and return a standard root command
* Add standard flags and subcommands to an existing root command
* Functions for setting Cobra Command flag values based on environment variables
* A `version` subcommand that will print out all the imported packages, including overrides specified in `go.mod` (at the time of compilation)
* Logrus configuration and top level logging related flags (like `log-format` and `log-level`)

# Installation
```bash
go get -u github.com/SkyMack/clibase
```

# Usage
```go
package main

import (
	"github.com/SkyMack/clibase"
	log "github.com/sirupsen/logrus"
)

const (
	appName        = "thumbnailer"
	appDescription = "Generates sequentially numbered thumbnail images based on the specified image and text settings."
)

func main() {
	// Create a new, standard root command
	rootCmd := clibase.New(appName, appDescription)

	// Add application specific subcommands
	AddCmdGeneratePng(rootCmd)

	// Execute command
	if err := rootCmd.Execute(); err != nil {
		log.WithFields(
			log.Fields{
				"app.name": appName,
				"error":    err.Error(),
			},
		).Fatal("application exited with an error")
	}
}
```