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
	IndividualFacts      map[string]string `json:"individual_facts,omitempty"`
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

// parseFactTypes parses a comma-separated string of fact types and returns a slice of fact types
func parseFactTypes(factTypesStr string) []string {
	if factTypesStr == "" {
		return []string{}
	}

	// Split by comma and trim whitespace
	factTypes := strings.Split(factTypesStr, ",")
	for i, t := range factTypes {
		factTypes[i] = strings.TrimSpace(t)
	}

	return factTypes
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
	factTypesStr := c.String("fact-types")
	args := make([]string, 0)

	// Parse the type parameter to handle comma-separated values
	types := parseTypes(metadataType)

	// Parse the fact-types parameter to handle comma-separated values
	factTypes := parseFactTypes(factTypesStr)
	useIndividualFacts := len(factTypes) > 0

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
			// Check if this is the throttling error we're looking for
			errStr := err.Error()
			if strings.Contains(errStr, "invalid character '<' looking for beginning of value") {
				return fmt.Errorf("API throttling detected - server returned HTML instead of JSON. Error: %v", err)
			} else if strings.Contains(errStr, "decoding response") {
				return fmt.Errorf("API response decoding error - possible throttling or server issue. Error: %v", err)
			}
			return err
		}

		// Get extended project info to access productionEnvironment field
		extendedProject := &schema.Project{}
		err = client.ProjectByNameExtended(ctx, v, extendedProject)
		if err != nil {
			// Check if this is the throttling error we're looking for
			errStr := err.Error()
			if strings.Contains(errStr, "invalid character '<' looking for beginning of value") {
				return fmt.Errorf("API throttling detected during extended project fetch - server returned HTML instead of JSON. Error: %v", err)
			} else if strings.Contains(errStr, "decoding response") {
				return fmt.Errorf("API response decoding error during extended project fetch - possible throttling or server issue. Error: %v", err)
			}
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

		// Fetch facts if requested or if individual fact types are specified
		if includeFacts || useIndividualFacts {
			facts := make(map[string]string)
			individualFacts := make(map[string]string)

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
				// Capture full error details for debugging
				errStr := err.Error()
				facts["error_type"] = "ProcessRaw_Error"
				facts["full_error"] = errStr

				// Check for specific error patterns
				if strings.Contains(errStr, "invalid character '<'") || strings.Contains(errStr, "decoding response") {
					facts["status"] = "API returned HTML error page - possible throttling or server error"
					// Try to extract more details from the error
					if strings.Contains(errStr, "invalid character '<' looking for beginning of value") {
						facts["likely_cause"] = "HTML_response_instead_of_JSON"
					}
				} else if strings.Contains(strings.ToLower(errStr), "throttl") || strings.Contains(strings.ToLower(errStr), "rate limit") {
					facts["status"] = "API throttling detected"
					facts["likely_cause"] = "Rate_limiting"
				} else if strings.Contains(strings.ToLower(errStr), "timeout") {
					facts["status"] = "Request timeout"
					facts["likely_cause"] = "Timeout"
				} else {
					facts["status"] = fmt.Sprintf("Unable to fetch facts: %v", err)
				}
			} else {
				// Parse the response
				responseBytes, err := json.Marshal(response)
				if err != nil {
					facts["status"] = fmt.Sprintf("Unable to marshal response: %v", err)
				} else {
					// Check if the response looks like HTML (error page)
					responseStr := string(responseBytes)
					if strings.HasPrefix(responseStr, "\"<") || strings.Contains(responseStr, "<html") {
						facts["status"] = "GraphQL API returned HTML error page instead of JSON"
					} else {
						var projectResponse ProjectByNameResponse
						err = json.Unmarshal(responseBytes, &projectResponse)
						if err != nil {
							facts["status"] = fmt.Sprintf("Unable to parse response: %v", err)
						} else {
							// Extract facts from production environments
							if len(projectResponse.ProjectByName.Environments) > 0 {
								for _, env := range projectResponse.ProjectByName.Environments {
									if env.Name == "production" || env.Name == "master" {
										for _, fact := range env.Facts {
											if fact.Name != "" && fact.Value != "" {
												// Always store in facts for backward compatibility
												if includeFacts {
													facts[fact.Name] = fact.Value
												}

												// Store individual facts if specific types are requested
												if useIndividualFacts {
													for _, requestedType := range factTypes {
														if fact.Name == requestedType {
															individualFacts[fact.Name] = fact.Value
															break
														}
													}
												}
											}
										}
									} else {
										continue
									}
								}
								if includeFacts && len(facts) == 0 {
									facts["status"] = "No facts available for production environment"
								}
							} else {
								if includeFacts {
									facts["status"] = "No production environment found"
								}
							}
						}
					}
				}
			}

			if includeFacts {
				item.Facts = facts
			}
			if useIndividualFacts {
				item.IndividualFacts = individualFacts
			}
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
		if useIndividualFacts {
			for _, factType := range factTypes {
				header = append(header, factType)
			}
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
			if useIndividualFacts {
				for _, factType := range factTypes {
					value := ""
					if item.IndividualFacts != nil {
						if val, exists := item.IndividualFacts[factType]; exists {
							value = val
						}
					}
					record = append(record, value)
				}
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
		if useIndividualFacts {
			for _, factType := range factTypes {
				headerCells = append(headerCells, &simpletable.Cell{Align: simpletable.AlignLeft, Text: factType})
			}
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
			if useIndividualFacts {
				for _, factType := range factTypes {
					value := ""
					if item.IndividualFacts != nil {
						if val, exists := item.IndividualFacts[factType]; exists {
							value = val
						}
					}
					r = append(r, &simpletable.Cell{Text: value})
				}
			}
			table.Body.Cells = append(table.Body.Cells, r)
		}
		table.SetStyle(simpletable.StyleCompactLite)
		io.WriteString(c.Writer, table.String())
	}

	return nil
}
