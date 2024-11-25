package main

import (
	"log"
	"os"

	"github.com/dpc-sdp/bay-cli/cmd/kms"
	cli "github.com/urfave/cli/v2"

	deployment "github.com/dpc-sdp/bay-cli/cmd/deployment"
	project_map "github.com/dpc-sdp/bay-cli/cmd/project-map"
)

const (
	EnvLagoonProject         = "LAGOON_PROJECT"
	EnvLagoonEnvironmentType = "LAGOON_ENVIRONMENT_TYPE"
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
								EnvVars:  []string{EnvLagoonProject},
								Required: true,
							},
							&cli.StringFlag{
								Name:     "key",
								Usage:    "Name of key",
								EnvVars:  []string{EnvLagoonEnvironmentType},
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
				Name:  "deployment",
				Usage: "commands for deployment actions",
				Subcommands: []*cli.Command{
					{
						Name:      "metadata",
						Usage:     "generates a file with deployment metadata",
						UsageText: "bay deployment metadata",
						Action:    deployment.DeploymentMetadata,
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
