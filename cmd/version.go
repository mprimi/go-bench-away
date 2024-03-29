package cmd

import (
	"flag"
	"fmt"

	"github.com/mprimi/go-bench-away/v1/core"

	"github.com/google/subcommands"
)

type versionCmd struct {
	baseCommand
}

func versionCommand() subcommands.Command {
	return &versionCmd{
		baseCommand: baseCommand{
			name:     "version",
			synopsis: "prints version information",
			usage:    "version\n",
			setFlags: func(_ *flag.FlagSet) {},
			execute: func(f *flag.FlagSet) error {
				fmt.Printf(
					"%s version:%s (%s) (built: %s)\n",
					core.Name,
					core.Version,
					core.SHA,
					core.BuildDate,
				)
				return nil
			},
		},
	}
}
