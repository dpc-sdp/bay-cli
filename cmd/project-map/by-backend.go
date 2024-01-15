package project_map

import (
	"encoding/json"
	"fmt"
	"github.com/urfave/cli/v2"
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

func ByBackend(c *cli.Context) error {
	client, err := helpers.NewLagoonClient(nil)
	if err != nil {
		return err
	}

	output := ByBackendResponse{}

	args := helpers.GetAllArgs(c)
	for _, v := range args {
		project := &schema.ProjectMetadata{}
		err := client.ProjectByNameMetadata(c.Context, v, project)
		if err != nil {
			return err
		}

		projects := make([]schema.ProjectMetadata, 0)
		err = client.ProjectsByMetadata(c.Context, "backend-project", v, &projects)
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

	a, _ := json.Marshal(output)

	fmt.Fprintf(c.App.Writer, string(a))

	return nil
}
