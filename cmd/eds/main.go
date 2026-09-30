package main

import "github.com/cloud-ru/evolution-devservices-cli/cmd"

// version is set via -ldflags "-X main.version=..." at build time.
var version = "dev"

func main() {
	cmd.SetVersion(version)
	cmd.Execute()
}
