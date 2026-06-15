package main

import (
	"os"

	"github.com/JdgaleTorre/onering/cmd"
)

func main() {
	if err := cmd.Execute(); err != nil {
		os.Exit(1)
	}
}
