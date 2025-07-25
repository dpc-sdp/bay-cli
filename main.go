package main

import (
	"context"
	"log"
	"os"

	"github.com/dpc-sdp/bay-cli/cmd/kms"
	cli "github.com/urfave/cli/v3"

	deployment "github.com/dpc-sdp/bay-cli/cmd/deployment"
	elastic_cloud "github.com/dpc-sdp/bay-cli/cmd/elastic-cloud"
	project_map "github.com/dpc-sdp/bay-cli/cmd/project-map"
)

const (
	EnvLagoonProject         = "LAGOON_PROJECT"
	EnvLagoonEnvironmentType = "LAGOON_ENVIRONMENT_TYPE"
)

func main() {
	app := &cli.Command{
		Name:  "bay",
		Usage: "CLI tool to interact with the Bay container platform",
		Commands: []*cli.Command{
			{
				Name:  "kms",
				Usage: "interact with KMS encryption service",
				Commands: []*cli.Command{
					{
						Name:      "encrypt",
						Usage:     "encrypt data",
						UsageText: "cat file.pem | bay kms encrypt > file.pem.asc",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "project",
								Usage:    "Name of lagoon project",
								Sources:  cli.EnvVars(EnvLagoonProject),
								Required: true,
							},
							&cli.StringFlag{
								Name:     "key",
								Usage:    "Name of key",
								Sources:  cli.EnvVars(EnvLagoonEnvironmentType),
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
				Commands: []*cli.Command{
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
				Name:  "deployment",
				Usage: "commands for deployment actions",
				Commands: []*cli.Command{
					{
						Name:      "metadata",
						Usage:     "generates a json object with deployment metadata",
						UsageText: "bay deployment metadata",
						Action:    deployment.DeploymentMetadata,
					},
				},
			},
			{
				Name:   "elastic-cloud",
				Usage:  "commands to interact with Elastic Cloud deployments",
				Hidden: true,
				Flags: []cli.Flag{
					&cli.StringFlag{
						Name:     "deployment-id",
						Usage:    "cloud deployment ID as listed on the Elastic Cloud 'manage' page",
						Required: true,
						Sources:  cli.EnvVars("EC_DEPLOYMENT_CLOUD_ID"),
					},
					&cli.StringFlag{
						Name:     "deployment-api-key",
						Required: true,
						Hidden:   true,
						Sources:  cli.EnvVars("EC_DEPLOYMENT_API_KEY"),
					},
				},
				Commands: []*cli.Command{
					{
						Name:      "unassigned-shards",
						Usage:     "Prints unassigned shards in JSON format",
						UsageText: "bay elastic-cloud unassigned-shards",
						Action:    elastic_cloud.ListUnassignedShards,
						Flags:     []cli.Flag{},
					},
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
							&cli.BoolFlag{
								Name:  "output-delete-list",
								Usage: "outputs a list of indices that would be deleted",
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

	if err := app.Run(context.Background(), os.Args); err != nil {
		log.Fatal(err)
	}
}
