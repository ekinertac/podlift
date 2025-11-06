package main

import (
	"os"

	"github.com/ekinertac/podlift/cmd/podlift/commands"
)

func main() {
	if err := commands.Execute(); err != nil {
		os.Exit(1)
	}
}

