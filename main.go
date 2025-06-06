package main

import (
	"fmt"
	"time"

	"github.com/DrakkarStorm/deadlinkr/cmd"
	"github.com/DrakkarStorm/deadlinkr/logger"
	"github.com/DrakkarStorm/deadlinkr/model"
)

// version is replaced by GoReleaser via ldflags at build time
var version = "dev"

func main() {
	model.TimeExecution = time.Now()

	logger.InitLogger(logger.DebugLevel)
	defer logger.CloseLogger()

	fmt.Printf("deadlinkr %s\n", version)
	cmd.Execute()

	logger.Durationf("Completed with version %s", version)
}
