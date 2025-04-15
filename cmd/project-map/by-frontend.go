package project_map

import (
	"encoding/json"
	"fmt"
	"github.com/alexeyco/simpletable"
	"github.com/urfave/cli/v2"
	"github.com/uselagoon/machinery/api/schema"

	"github.com/dpc-sdp/bay-cli/internal/helpers"
)

type ByFrontendResponse struct {
	Items map[string]string `json:"items"`
}

func ByFrontend(c *cli.Context) error {
	client, err := helpers.NewLagoonClient(nil)
	if err != nil {
		return err
	}

	output := ByFrontendResponse{
		Items: make(map[string]string, 0),
	}

	all := c.Bool("all")
	args := make([]string, 0)
	if all {
		// @todo once all frontends are on ripple 2, remove the obsolete check.
		rippleProjects := make([]schema.ProjectMetadata, 0)
		err = client.ProjectsByMetadata(c.Context, "type", "ripple", &rippleProjects)
		if err != nil {
			return err
		}
		for _, p := range rippleProjects {
			args = append(args, p.Name)
		}

		ripple2Projects := make([]schema.ProjectMetadata, 0)
		err = client.ProjectsByMetadata(c.Context, "type", "ripple2", &ripple2Projects)
		if err != nil {
			return err
		}
		for _, p := range ripple2Projects {
			args = append(args, p.Name)
		}
	} else {
		args = helpers.GetAllArgs(c)
	}

	if len(args) == 0 {
		return fmt.Errorf("no project specified, did you mean to add the --all flag?")
	}

	for _, v := range args {
		project := &schema.ProjectMetadata{}

		err := client.ProjectByNameMetadata(c.Context, v, project)
		if err != nil {
			return err
		}
		if err != nil {
			return err
		}

		output.Items[v] = project.Metadata["backend-project"]
	}

	if c.String("output") == "json" {
		a, _ := json.Marshal(output)
		fmt.Fprintf(c.App.Writer, string(a))
	} else {
		table := simpletable.New()

		table.Header = &simpletable.Header{
			Cells: []*simpletable.Cell{
				{Align: simpletable.AlignLeft, Text: "Frontend"},
				{Align: simpletable.AlignLeft, Text: "Backend"},
			},
		}

		for frontend, backend := range output.Items {
			r := []*simpletable.Cell{
				{Text: frontend},
				{Text: backend},
			}
			table.Body.Cells = append(table.Body.Cells, r)
		}
		table.SetStyle(simpletable.StyleCompactLite)
		fmt.Fprintf(c.App.Writer, table.String())
	}

	return nil
}
