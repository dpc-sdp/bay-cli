package project_map

import (
	"encoding/json"
	"fmt"
	"github.com/pkg/errors"
	"path"
	"strings"
	"text/template"

	"github.com/alexeyco/simpletable"
	"github.com/urfave/cli/v2"
	"github.com/uselagoon/machinery/api/schema"

	"github.com/dpc-sdp/bay-cli/internal/helpers"
	template_helpers "github.com/dpc-sdp/bay-cli/internal/template"
)

type ByBackendResponse struct {
	Items []ByBackendResponseItem `json:"items"`
}

type ByBackendResponseItem struct {
	Project   string   `json:"project"`
	FrontEnds []string `json:"frontends"`
}

type ByBackendTemplateVars struct {
	Inventory map[string]schema.Project
	Items     []ByBackendResponseItem
}

func ByBackend(c *cli.Context) error {
	client, err := helpers.NewLagoonClient(nil)
	if err != nil {
		return err
	}

	output := ByBackendResponse{}
	projectInventory := make(map[string]schema.Project)

	all := c.Bool("all")
	args := make([]string, 0)
	if all {
		projects := make([]schema.ProjectMetadata, 0)
		err = client.ProjectsByMetadata(c.Context, "type", "tide", &projects)
		if err != nil {
			return err
		}
		for _, p := range projects {
			args = append(args, p.Name)

			projectInventory[p.Name] = p.Project
		}
	} else {
		args = helpers.GetAllArgs(c)
	}

	for _, v := range args {
		project := &schema.Project{}
		err := client.ProjectByNameExtended(c.Context, v, project)
		if err != nil {
			return err
		}
		projectInventory[project.Name] = *project

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
			projectInventory[p.Name] = p.Project
		}

		output.Items = append(output.Items, item)
	}

	// JSON output
	outputFormat := c.String("output")
	if outputFormat == "json" {
		a, _ := json.Marshal(output)
		fmt.Fprintf(c.App.Writer, string(a))
	} else if outputFormat == "go-template-file" {
		pathOnDisk := c.String("go-template-file")
		basePath := path.Base(pathOnDisk)
		if pathOnDisk == "" {
			return fmt.Errorf("--go-template-file flag was empty, but is a required field when using the --output=go-template-file option")
		}
		templateInputs := ByBackendTemplateVars{
			Inventory: projectInventory,
			Items:     output.Items,
		}

		tmpl, err := template.New(basePath).Funcs(template.FuncMap{
			"convert_github_url_to_page": template_helpers.ConvertGithubUriToPage,
		}).ParseFiles(pathOnDisk)
		if err != nil {
			return errors.Wrap(err, "could not load template file")
		}
		err = tmpl.Execute(c.App.Writer, templateInputs)
		if err != nil {
			return errors.Wrap(err, "could not render template")
		}
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
		table.SetStyle(simpletable.StyleMarkdown)
		fmt.Fprintf(c.App.Writer, table.String())
	}

	return nil
}
