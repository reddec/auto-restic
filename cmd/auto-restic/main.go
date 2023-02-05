package main

import (
	"os"

	"github.com/jessevdk/go-flags"
	"github.com/reddec/auto-restic/cmd/auto-restic/commands"
)

type Config struct {
	Run   commands.Run   `command:"run" description:"Run service"`
	Ready commands.Ready `command:"ready" description:"Check is service ready"`
}

func main() {
	var config Config
	parser := flags.NewParser(&config, flags.Default)
	parser.ShortDescription = "Backup and restore"

	if _, err := parser.Parse(); err != nil {
		os.Exit(1)
	}

}
