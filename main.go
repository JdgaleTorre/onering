package main

import (
	"os"

	"github.com/josegale/onering/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
