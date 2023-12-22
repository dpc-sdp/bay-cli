package main

import (
	"log"
	"os"

	cli "github.com/urfave/cli/v2"

	"github.com/dpc-sdp/bay-cli/cmd/dr"
	"github.com/dpc-sdp/bay-cli/cmd/kms"
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
			//
			{
				Name:  "dr",
				Usage: "interact with Section.io QuantCDN DR system",
				Subcommands: []*cli.Command{
					{
						Name:      "enable",
						Usage:     "enable QuantCDN DR system",
						UsageText: "bay dr enable --application www.vic.gov.au --environment develop --hosts=www.schools.vic.gov.au",
						Flags: []cli.Flag{
							&cli.StringFlag{
								Name:     "application",
								Usage:    "Name of section.io application",
								Required: true,
							},
							&cli.StringFlag{
								Name:        "environment",
								Usage:       "Name of section.io environment - defaults to develop",
								DefaultText: dr.DrFlagDefaultValueEnvironment,
							},
							&cli.StringFlag{
								Name:        "hosts",
								Usage:       "Hostnames that should have DR enabled - defaults to all hostnames",
								DefaultText: dr.DrFlagDefaultValueHosts,
							},
							&cli.StringFlag{
								Name:    "section_username",
								Usage:   "Username for section.io",
								EnvVars: []string{"SECTION_USERNAME"},
							},
							&cli.StringFlag{
								Name:    "section_password",
								Usage:   "Password for section.io",
								EnvVars: []string{"SECTION_PASSWORD"},
							},
						},
						Action: dr.Enable,
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
