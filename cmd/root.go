package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"

	"github.com/google/subcommands"
	"github.com/mprimi/go-bench-away/internal/core"
)

var rootOptions struct {
	// Top-level options
	verbose       bool
	natsServerUrl string
	credentials   string
	namespace     string
}

type baseCommand struct {
	name     string
	synopsis string
	usage    string
	setFlags func(*flag.FlagSet)
	execute  func(*flag.FlagSet) error
}

func (bCmd *baseCommand) Name() string { return bCmd.name }

func (bCmd *baseCommand) Synopsis() string { return bCmd.synopsis }

func (bCmd *baseCommand) Usage() string { return bCmd.usage }

func (bCmd *baseCommand) SetFlags(f *flag.FlagSet) {
	if bCmd.setFlags != nil {
		bCmd.setFlags(f)
	}
}

func (bCmd *baseCommand) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if bCmd.execute == nil {
		fmt.Fprintf(os.Stderr, "Not implemented\n")
		return subcommands.ExitFailure
	}

	if rootOptions.verbose {
		fmt.Printf("%s args: %v\n", bCmd.name, f.Args())
	}

	err := bCmd.execute(f)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Command %s failed: %v\n", bCmd.name, err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}

// Entry point
func Run(args []string) int {

	rootFlagSet := flag.NewFlagSet("", flag.ExitOnError)
	rootFlagSet.BoolVar(&rootOptions.verbose, "v", false, "verbose")
	rootFlagSet.StringVar(&rootOptions.natsServerUrl, "server", "nats://localhost:4222", "NATS server URL")
	rootFlagSet.StringVar(&rootOptions.credentials, "creds", "", "Path to credentials file")
	rootFlagSet.StringVar(&rootOptions.namespace, "namespace", "default", "Namespace (allows isolated sets of jobs to share a NATS server)")

	cmdr := subcommands.NewCommander(rootFlagSet, core.Name)
	cmdr.ImportantFlag("server")
	cmdr.ImportantFlag("v")
	cmdr.ImportantFlag("creds")

	commandsMap := map[string][]subcommands.Command{
		"maintenance": {
			initCommand(),
			wipeCommand(),
		},
		"submit, monitor": {
			submitCommand(),
			waitCommand(),
		},
		"job debugging": {
			downloadCommand(),
			logCommand(),
		},
		"reporting, analysis & graphs": {
			trendReportCommand(),
			basicReportCommand(),
			comparativeReportCommand(),
			customReportCommand(),
		},
		"worker": {
			workerCommand(),
		},
		"explore job status": {
			listCommand(),
			webCommand(),
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
	return int(cmdr.Execute(context.Background()))
}
