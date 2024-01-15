package helpers

import "github.com/urfave/cli/v2"

func GetAllArgs(c *cli.Context) []string {
	args := make([]string, 0)
	i := 0
	l := c.Args().Len()
	for i < l {
		args = append(args, c.Args().Get(i))
		i++
	}
	return args
}
