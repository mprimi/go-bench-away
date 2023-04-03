package cmd

import (
	"context"
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/mprimi/go-bench-away/pkg/client"

	"github.com/google/subcommands"
)

type downloadCmd struct {
	baseCommand
	outputDirPath string
}

func downloadCommand() subcommands.Command {
	return &downloadCmd{
		baseCommand: baseCommand{
			name:     "download",
			synopsis: "Download job records and artifacts",
			usage:    "download <jobId> [jobId [...]]\n",
		},
	}
}

func (cmd *downloadCmd) SetFlags(f *flag.FlagSet) {
	f.StringVar(&cmd.outputDirPath, "output", ".", "Output directory")
}

func (cmd *downloadCmd) Execute(_ context.Context, f *flag.FlagSet, _ ...interface{}) subcommands.ExitStatus {
	if rootOptions.verbose {
		fmt.Printf("%s args: %v\n", cmd.name, f.Args())
	}

	if len(f.Args()) < 1 {
		fmt.Fprintf(os.Stderr, "Missing job ID argument(s)\n")
		return subcommands.ExitUsageError
	}

	outputDir, err := os.Open(cmd.outputDirPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return subcommands.ExitUsageError
	}
	defer outputDir.Close()

	fileInfo, err := outputDir.Stat()
	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return subcommands.ExitFailure
	}

	if !fileInfo.IsDir() {
		fmt.Fprintf(os.Stderr, "Not a directory: %s\n", cmd.outputDirPath)
		return subcommands.ExitUsageError
	}
	outputDir.Close()

	c, err := client.NewClient(
		rootOptions.natsServerUrl,
		rootOptions.credentials,
		rootOptions.namespace,
		client.InitJobsRepository(),
		client.InitArtifactsStore(),
		client.Verbose(rootOptions.verbose),
	)

	if err != nil {
		fmt.Fprintf(os.Stderr, "%v\n", err)
		return subcommands.ExitFailure
	}
	defer c.Close()

	for _, jobId := range f.Args() {
		job, _, err := c.LoadJob(jobId)
		if err != nil {
			fmt.Fprintf(os.Stderr, "%v\n", err)
			return subcommands.ExitFailure
		}

		if job.Log == "" {
			fmt.Printf("No log artifact for job %s\n", job.Id)
		} else {
			fileName := fmt.Sprintf("%s_log.txt", jobId)
			filePath := filepath.Join(cmd.outputDirPath, fileName)
			err := c.DownloadLogArtifact(job, filePath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Download failed: %v\n", err)
				return subcommands.ExitFailure
			}
			fmt.Printf("Downloaded %s\n", filePath)
		}

		if job.Results == "" {
			fmt.Printf("No results artifact for job %s\n", job.Id)
		} else {
			fileName := fmt.Sprintf("%s_results.txt", jobId)
			filePath := filepath.Join(cmd.outputDirPath, fileName)
			err := c.DownloadResultsArtifact(job, filePath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Download failed: %v\n", err)
				return subcommands.ExitFailure
			}
			fmt.Printf("Downloaded %s\n", filePath)
		}

		if job.Script == "" {
			fmt.Printf("No script artifact for job %s\n", job.Id)
		} else {
			fileName := fmt.Sprintf("%s_run.sh", jobId)
			filePath := filepath.Join(cmd.outputDirPath, fileName)
			err := c.DownloadScriptArtifact(job, filePath)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Download failed: %v\n", err)
				return subcommands.ExitFailure
			}
			fmt.Printf("Downloaded %s\n", filePath)
		}
	}

	return subcommands.ExitSuccess
}
