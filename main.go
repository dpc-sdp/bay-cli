package main

import (
	"github.com/dpc-sdp/bay-cli/cmd/kms"
	cli "github.com/urfave/cli/v2"
	"log"
	"os"
)

func main() {
	app := &cli.App{
		//Name:  "bay",
		//Usage: "CLI tool to interact with the Bay container platform",
		//Action: func(*cli.Context) error {
		//	fmt.Println("boom! I say!")
		//	return nil
		//},
		Commands: []*cli.Command{
			{
				Name:  "kms",
				Usage: "Colllection of commands to interact with KMS encryption service",
				Subcommands: []*cli.Command{
					{
						Name:  "encrypt",
						Usage: "encrypt a file",
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
						Action: kms.Encrpyt,
					},
				},
			},
		},
	}

	if err := app.Run(os.Args); err != nil {
		log.Fatal(err)
	}
}
