package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/google/subcommands"
)

type rootOptions struct {
	// Build-time properties
	name      string
	version   string
	sha       string
	buildDate string

	// Top-level options
	verbose       bool
	natsServerUrl string
	credentials   string
}

type baseCommand struct {
	name     string
	synopsis string
	usage    string
	setFlags func(*flag.FlagSet)
	execute  func(*rootOptions, *flag.FlagSet) error
}

func (bCmd *baseCommand) Name() string { return bCmd.name }

func (bCmd *baseCommand) Synopsis() string { return bCmd.synopsis }

func (bCmd *baseCommand) Usage() string { return bCmd.usage }

func (bCmd *baseCommand) SetFlags(f *flag.FlagSet) {
	if bCmd.setFlags != nil {
		bCmd.setFlags(f)
	}
}

func (bCmd *baseCommand) Execute(_ context.Context, f *flag.FlagSet, args ...interface{}) subcommands.ExitStatus {
	if bCmd.execute == nil {
		fmt.Fprintf(os.Stderr, "Not implemented")
		return subcommands.ExitFailure
	}

	var rootOpts *rootOptions = args[0].(*rootOptions)

	if rootOpts.verbose {
		fmt.Printf("%s args: %v\n", bCmd.name, f.Args())
	}

	err := bCmd.execute(rootOpts, f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Command %s failed: %v\n", bCmd.name, err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}

// Entry point
func Run(name, version, sha, buildDate string, args []string) int {
	var rootOps = rootOptions{
		name:      name,
		version:   version,
		sha:       sha,
		buildDate: buildDate,
	}

	rootFlagSet := flag.NewFlagSet("", flag.ExitOnError)
	rootFlagSet.BoolVar(&rootOps.verbose, "v", false, "verbose")
	rootFlagSet.StringVar(&rootOps.natsServerUrl, "server", "nats://localhost:4222", "NATS server URL")
	rootFlagSet.StringVar(&rootOps.credentials, "creds", "", "Path to credentials file")

	cmdr := subcommands.NewCommander(rootFlagSet, name)
	cmdr.ImportantFlag("server")
	cmdr.ImportantFlag("v")
	cmdr.ImportantFlag("creds")

	commandsMap := map[string][]subcommands.Command{
		"maintenance": {
			initCommand(),
			wipeCommand(),
		},
		"submit, monitor, find jobs": {
			submitCommand(),
			waitCommand(),
			listCommand(),
		},
		"reporting, analysis & graphs": {
			compareSpeedCommand(),
			downloadCommand(),
		},
		"worker": {
			workerCommand(),
		},
		"help": {
			versionCommand(),
			cmdr.HelpCommand(),
			cmdr.FlagsCommand(),
			cmdr.CommandsCommand(),
		},
	}

	for groupName, commands := range commandsMap {
		for _, command := range commands {
			cmdr.Register(command, groupName)
		}
	}

	err := rootFlagSet.Parse(args)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Failed to parse arguments: %v\n", err)
		return 1
	}
	return int(cmdr.Execute(context.Background(), &rootOps))
}
