package main

import (
	"os"

	"github.com/josegale/lazycode/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
