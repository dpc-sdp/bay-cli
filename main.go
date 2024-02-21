package main

import (
	"log"
	"os"

	"github.com/dpc-sdp/bay-cli/cmd/kms"
	cli "github.com/urfave/cli/v2"

	elastic_cloud "github.com/dpc-sdp/bay-cli/cmd/elastic-cloud"
	project_map "github.com/dpc-sdp/bay-cli/cmd/project-map"
)

func main() {
	app := &cli.App{
		Name:  "bay",
		Usage: "CLI tool to interact with the Bay container platform",
		Commands: []*cli.Command{
			{
				Name:  "kms",
				Usage: "interact with KMS encryption service",
				Subcommands: []*cli.Command{
					{
						Name:      "encrypt",
						Usage:     "encrypt data",
						UsageText: "cat file.pem | bay kms encrypt > file.pem.asc",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "project",
								Usage:    "Name of lagoon project",
								Required: true,
							},
							&cli.StringFlag{
								Name:     "key",
								Usage:    "Name of key",
								Required: true,
							},
						},
						Action: kms.Encrypt,
					},
					{
						Name:      "decrypt",
						Usage:     "decrypt a file",
						UsageText: "cat file.pem.asc | bay kms decrypt > file.pem",
						Action:    kms.Decrypt,
					},
				},
			},
			{
				Name:  "project-map",
				Usage: "commands to show relationships between projects",
				Subcommands: []*cli.Command{
					{
						Name:      "by-backend",
						Usage:     "shows all frontends that connect to a specific backend",
						UsageText: "bay project-map by-backend --all",
						Flags: []cli.Flag{
							&cli.BoolFlag{
								Name:  "all",
								Usage: "List all projects",
							},
							&cli.StringFlag{
								Name:        "output",
								Usage:       "Output format - supports json, table",
								DefaultText: "table",
							},
						},
						Action: project_map.ByBackend,
					},
					{
						Name:      "by-frontend",
						Usage:     "shows the backend that a list of frontends connect to",
						UsageText: "bay project-map by-frontend --all",
						Flags: []cli.Flag{
							&cli.BoolFlag{
								Name:  "all",
								Usage: "List all projects",
							},
							&cli.StringFlag{
								Name:        "output",
								Usage:       "Output format - supports json, table",
								DefaultText: "table",
							},
						},
						Action: project_map.ByFrontend,
					},
				},
			},
			{
				Name:  "elastic-cloud",
				Usage: "commands to interact with Elastic Cloud deployments",
				Subcommands: []*cli.Command{
					{
						Name:      "delete-stale",
						Usage:     "deletes stale indices (> 30 days old)",
						UsageText: "bay elastic-cloud delete-stale",
						Action:    elastic_cloud.DeleteStaleIndices,
						Flags: []cli.Flag{
							&cli.BoolFlag{
								Name:  "force",
								Usage: "skips all confirmation prompts and immediately executes mutations",
							},
							&cli.StringFlag{
								Name:     "EC_DEPLOYMENT_CLOUD_ID",
								Required: true,
							},
							&cli.Int64Flag{
								Name:  "age",
								Value: int64(30),
								Usage: "sets the minimum age of indices to be marked for deletion",
							},
						},
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
