package project_map

import (
	"encoding/json"
	"fmt"
	"path"
	"text/template"

	"github.com/alexeyco/simpletable"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"github.com/uselagoon/machinery/api/schema"

	"github.com/dpc-sdp/bay-cli/internal/helpers"
	template_helpers "github.com/dpc-sdp/bay-cli/internal/template"
)

type ByFrontendResponse struct {
	Items map[string]string `json:"items"`
}

type ByFrontendTemplateVars struct {
	Inventory map[string]schema.Project
	Items     map[string]string
}

func ByFrontend(c *cli.Context) error {
	client, err := helpers.NewLagoonClient(nil)
	if err != nil {
		return err
	}

	output := ByFrontendResponse{
		Items: make(map[string]string, 0),
	}
	projectInventory := make(map[string]schema.Project)

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
			projectInventory[p.Name] = p.Project
		}

		ripple2Projects := make([]schema.ProjectMetadata, 0)
		err = client.ProjectsByMetadata(c.Context, "type", "ripple2", &ripple2Projects)
		if err != nil {
			return err
		}
		for _, p := range ripple2Projects {
			args = append(args, p.Name)
			projectInventory[p.Name] = p.Project
		}
	} else {
		args = helpers.GetAllArgs(c)
	}

	for _, v := range args {
		if _, ok := projectInventory[v]; !ok {
			project := &schema.ProjectMetadata{}
			err := client.ProjectByNameMetadata(c.Context, v, project)
			if err != nil {
				return err
			}
			projectInventory[v] = project.Project
		}

		if err != nil {
			return err
		}

		output.Items[v] = project.Metadata["backend-project"]
	}

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

		templateInputs := ByFrontendTemplateVars{
			Items: output.Items,
			Inventory:
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
