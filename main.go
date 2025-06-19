package main

import (
	"fmt"

	"github.com/DrakkarStorm/deadlinkr/cmd"
)

// version is replaced by GoReleaser via ldflags at build time
var version = "dev"

func main() {
	fmt.Printf("Starting with version %s\n", version)
	cmd.Execute()
	fmt.Printf("Completed with version %s\n", version)
}
