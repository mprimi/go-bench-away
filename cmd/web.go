package cmd

import (
	"context"
	"flag"
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/mprimi/go-bench-away/internal/web"
	"github.com/mprimi/go-bench-away/v1/client"

	"github.com/google/subcommands"
)

type webCmd struct {
	baseCommand
	port int
}

func webCommand() subcommands.Command {
	return &webCmd{
		baseCommand: baseCommand{
			name:     "web",
			synopsis: "start the web interface",
			usage:    "list [options]\n",
		},
	}
}

func (cmd *webCmd) SetFlags(f *flag.FlagSet) {
	f.IntVar(&cmd.port, "port", 8888, "Port number")
}

func (cmd *webCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if rootOptions.verbose {
		fmt.Printf("%s args: %v\n", cmd.name, f.Args())
	}

	c, err := client.NewClient(
		rootOptions.natsServerUrl,
		rootOptions.credentials,
		rootOptions.namespace,
		client.InitJobsQueue(),
		client.InitJobsRepository(),
		client.InitArtifactsStore(),
		client.Verbose(rootOptions.verbose),
	)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return subcommands.ExitFailure
	}
	defer c.Close()

	handler := web.NewHandler(c)

	s := &http.Server{
		Addr:         fmt.Sprintf(":%d", cmd.port),
		Handler:      handler,
		ReadTimeout:  10 * time.Second,
		WriteTimeout: 10 * time.Second,
	}

	fmt.Printf("Listening on: %s\n", s.Addr)
	err = s.ListenAndServe()

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return subcommands.ExitFailure
	}

	return subcommands.ExitSuccess
}
