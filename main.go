package main

import (
	"github.com/dpc-sdp/bay-cli/cmd/kms"
	cli "github.com/urfave/cli/v2"
	"log"
	"os"
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
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
