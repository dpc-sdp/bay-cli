package project_map

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/alexeyco/simpletable"
	"github.com/urfave/cli/v3"
	"github.com/uselagoon/machinery/api/schema"

	"github.com/dpc-sdp/bay-cli/internal/helpers"
)

type ByBackendResponse struct {
	Items []ByBackendResponseItem `json:"items"`
}

type ByBackendResponseItem struct {
	Project   string   `json:"project"`
	FrontEnds []string `json:"frontends"`
}

func ByBackend(ctx context.Context, c *cli.Command) error {
	client, err := helpers.NewLagoonClient(nil)
	if err != nil {
		return err
	}

	output := ByBackendResponse{}

	all := c.Bool("all")
	args := make([]string, 0)
	if all {
		projects := make([]schema.ProjectMetadata, 0)
		err = client.ProjectsByMetadata(ctx, "type", "tide", &projects)
		if err != nil {
			return err
		}
		for _, p := range projects {
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
		err := client.ProjectByNameMetadata(ctx, v, project)
		if err != nil {
			return err
		}

		projects := make([]schema.ProjectMetadata, 0)
		err = client.ProjectsByMetadata(ctx, "backend-project", v, &projects)
		if err != nil {
			return err
		}

		item := ByBackendResponseItem{
			Project:   v,
			FrontEnds: make([]string, 0),
		}

		for _, p := range projects {
			item.FrontEnds = append(item.FrontEnds, p.Name)
		}

		output.Items = append(output.Items, item)
	}

	if c.String("output") == "json" {
		a, _ := json.Marshal(output)
		io.WriteString(c.Writer, string(a))
	} else {
		table := simpletable.New()

		table.Header = &simpletable.Header{
			Cells: []*simpletable.Cell{
				{Align: simpletable.AlignLeft, Text: "Backend"},
				{Align: simpletable.AlignLeft, Text: "Frontends"},
			},
		}

		for _, item := range output.Items {
			r := []*simpletable.Cell{
				{Text: item.Project},
				{Text: strings.Join(item.FrontEnds, "\n")},
			}
			table.Body.Cells = append(table.Body.Cells, r)
		}
		table.SetStyle(simpletable.StyleCompactLite)
		io.WriteString(c.Writer, table.String())
	}

	return nil
}
