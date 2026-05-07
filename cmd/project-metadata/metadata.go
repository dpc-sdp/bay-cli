package project_metadata

import (
	"context"
	"encoding/csv"
	"encoding/json"
	"fmt"
	"io"
	"strings"

	"github.com/alexeyco/simpletable"
	"github.com/urfave/cli/v3"
	lagoon_client "github.com/uselagoon/machinery/api/lagoon/client"
	"github.com/uselagoon/machinery/api/schema"

	"github.com/dpc-sdp/bay-cli/internal/helpers"
)

type Response struct {
	Items []ProjectMetadata `json:"items"`
}

type ProjectMetadata struct {
	ProjectName          string            `json:"project"`
	Type                 string            `json:"type"`
	Maintainer           string            `json:"maintainer"`
	SectionIoApplication string            `json:"section-io-application"`
	ApexDomain           string            `json:"apex-domain"`
	BackendProject       string            `json:"backend-project"`
	ProductionDomain     string            `json:"production-domain"`
	Facts                map[string]string `json:"facts,omitempty"`
}

// Fact represents a single fact from the Lagoon API
type Fact struct {
	Name        string `json:"name"`
	Value       string `json:"value"`
	Description string `json:"description"`
}

// Environment represents an environment with facts
type Environment struct {
	Name  string `json:"name"`
	Facts []Fact `json:"facts"`
}

// ProjectByNameResponse represents the response from the ProjectByName query
type ProjectByNameResponse struct {
	ProjectByName struct {
		Name         string        `json:"name"`
		Environments []Environment `json:"environments"`
	} `json:"projectByName"`
}

// parseTypes parses a comma-separated string of types and returns a slice of types
func parseTypes(typeStr string) []string {
	if typeStr == "" || typeStr == "all" {
		return []string{"all"}
	}

	// Split by comma and trim whitespace
	types := strings.Split(typeStr, ",")
	for i, t := range types {
		types[i] = strings.TrimSpace(t)
	}

	return types
}

// getProjectsByTypes fetches projects for multiple types
func getProjectsByTypes(ctx context.Context, client *lagoon_client.Client, types []string) ([]string, error) {
	var allArgs []string

	for _, metadataType := range types {
		if metadataType == "all" {
			// Get all projects by fetching with empty type filter
			projects := make([]schema.ProjectMetadata, 0)
			err := client.ProjectsByMetadata(ctx, "type", "", &projects)
			if err != nil {
				return nil, err
			}
			for _, p := range projects {
				allArgs = append(allArgs, p.Name)
			}
		} else {
			// Get projects of specific type
			projects := make([]schema.ProjectMetadata, 0)
			err := client.ProjectsByMetadata(ctx, "type", metadataType, &projects)
			if err != nil {
				// Continue with other types if one fails
				continue
			}
			for _, p := range projects {
				allArgs = append(allArgs, p.Name)
			}
		}
	}

	return allArgs, nil
}

func Metadata(ctx context.Context, c *cli.Command) error {
	client, err := helpers.NewLagoonClient(nil)
	if err != nil {
		return err
	}

	output := Response{}

	all := c.Bool("all")
	metadataType := c.String("type")
	includeFacts := c.Bool("include-facts")
	args := make([]string, 0)

	// Parse the type parameter to handle comma-separated values
	types := parseTypes(metadataType)

	if all {
		// Get projects for all specified types
		projectArgs, err := getProjectsByTypes(ctx, client, types)
		if err != nil {
			return err
		}
		args = projectArgs
	} else {
		args = helpers.GetAllArgs(c)
	}

	// If no arguments are provided and --all flag is not set, default to returning all projects
	if len(args) == 0 && !all {
		// Get projects for all specified types
		projectArgs, err := getProjectsByTypes(ctx, client, types)
		if err != nil {
			return err
		}
		args = projectArgs
	}

	if len(args) == 0 {
		return fmt.Errorf("no projects found")
	}

	for _, v := range args {
		project := &schema.ProjectMetadata{}
		err := client.ProjectByNameMetadata(ctx, v, project)
		if err != nil {
			return err
		}

		// Get extended project info to access productionEnvironment field
		extendedProject := &schema.Project{}
		err = client.ProjectByNameExtended(ctx, v, extendedProject)
		if err != nil {
			return err
		}

		item := ProjectMetadata{
			ProjectName:          project.Name,
			Type:                 project.Metadata["type"],
			Maintainer:           project.Metadata["maintainer"],
			SectionIoApplication: project.Metadata["section-io-application"],
			ApexDomain:           project.Metadata["apex-domain"],
			BackendProject:       project.Metadata["backend-project"],
			ProductionDomain:     project.Metadata["production-domain"],
		}

		// Fetch facts if requested
		if includeFacts {
			facts := make(map[string]string)

			// Use the specific GraphQL query to fetch facts from production environment
			query := `
				query ProjectByName(
					$name: String!
					$type: EnvType
					$keyFacts: Boolean
					$summary: Boolean
				) {
					projectByName(name: $name) {
						name
						environments(type: $type) {
							name
							facts(keyFacts: $keyFacts, summary: $summary) {
								name
								value
								description
							}
						}
					}
				}
			`

			variables := map[string]interface{}{
				"name":     project.Name,
				"type":     "PRODUCTION",
				"keyFacts": true,
				"summary":  false,
			}

			response, err := client.ProcessRaw(ctx, query, variables)
			if err != nil {
				facts["status"] = fmt.Sprintf("Unable to fetch facts: %v", err)
			} else {
				// Parse the response
				responseBytes, err := json.Marshal(response)
				if err != nil {
					facts["status"] = fmt.Sprintf("Unable to marshal response: %v", err)
				} else {
					var projectResponse ProjectByNameResponse
					err = json.Unmarshal(responseBytes, &projectResponse)
					if err != nil {
						facts["status"] = fmt.Sprintf("Unable to parse response: %v", err)
					} else {
						// Extract facts from production environments
						if len(projectResponse.ProjectByName.Environments) > 0 {
							for _, env := range projectResponse.ProjectByName.Environments {
								for _, fact := range env.Facts {
									if fact.Name != "" && fact.Value != "" {
										facts[fact.Name] = fact.Value
									}
								}
							}
							if len(facts) == 0 {
								facts["status"] = "No facts available for production environment"
							}
						} else {
							facts["status"] = "No production environment found"
						}
					}
				}
			}

			item.Facts = facts
		}

		output.Items = append(output.Items, item)
	}

	if c.String("output") == "json" {
		a, _ := json.Marshal(output)
		io.WriteString(c.Writer, string(a))
	} else if c.String("output") == "csv" {
		writer := csv.NewWriter(c.Writer)
		defer writer.Flush()

		// Write CSV header
		header := []string{
			"Project",
			"Type",
			"Maintainer",
			"SectionIO App",
			"Apex Domain",
			"Backend Project",
			"Production Domain",
		}
		if includeFacts {
			header = append(header, "Facts")
		}
		writer.Write(header)

		// Write CSV data rows
		for _, item := range output.Items {
			record := []string{
				item.ProjectName,
				item.Type,
				item.Maintainer,
				item.SectionIoApplication,
				item.ApexDomain,
				item.BackendProject,
				item.ProductionDomain,
			}
			if includeFacts {
				factsStr := ""
				if item.Facts != nil {
					for key, value := range item.Facts {
						if factsStr != "" {
							factsStr += "; "
						}
						factsStr += fmt.Sprintf("%s: %s", key, value)
					}
				}
				record = append(record, factsStr)
			}
			writer.Write(record)
		}
	} else {
		table := simpletable.New()

		headerCells := []*simpletable.Cell{
			{Align: simpletable.AlignLeft, Text: "Project"},
			{Align: simpletable.AlignLeft, Text: "Type"},
			{Align: simpletable.AlignLeft, Text: "Maintainer"},
			{Align: simpletable.AlignLeft, Text: "SectionIO App"},
			{Align: simpletable.AlignLeft, Text: "Apex Domain"},
			{Align: simpletable.AlignLeft, Text: "Backend Project"},
			{Align: simpletable.AlignLeft, Text: "Production Domain"},
		}
		if includeFacts {
			headerCells = append(headerCells, &simpletable.Cell{Align: simpletable.AlignLeft, Text: "Facts"})
		}
		table.Header = &simpletable.Header{
			Cells: headerCells,
		}

		for _, item := range output.Items {
			r := []*simpletable.Cell{
				{Text: item.ProjectName},
				{Text: item.Type},
				{Text: item.Maintainer},
				{Text: item.SectionIoApplication},
				{Text: item.ApexDomain},
				{Text: item.BackendProject},
				{Text: item.ProductionDomain},
			}
			if includeFacts {
				factsStr := ""
				if item.Facts != nil {
					for key, value := range item.Facts {
						if factsStr != "" {
							factsStr += "\n"
						}
						factsStr += fmt.Sprintf("%s: %s", key, value)
					}
				}
				r = append(r, &simpletable.Cell{Text: factsStr})
			}
			table.Body.Cells = append(table.Body.Cells, r)
		}
		table.SetStyle(simpletable.StyleCompactLite)
		io.WriteString(c.Writer, table.String())
	}

	return nil
}
