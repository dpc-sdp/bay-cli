package elastic_cloud

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/dpc-sdp/bay-cli/internal/helpers"
	elasticsearch "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/manifoldco/promptui"
	errors "github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	lagoon_client "github.com/uselagoon/machinery/api/lagoon/client"
	"github.com/uselagoon/machinery/api/schema"
)

type IndexSettings struct {
	IndexItem struct {
		IndexDetail struct {
			CreationDate string `json:"creation_date"`
		} `json:"index"`
	} `json:"settings"`
}

type Indices map[string]IndexSettings

type IndicesByProject map[string][]Indices

type AliasAttr struct {
	IsHidden bool `json:"is_hidden"`
}

type Aliases struct {
	Aliases map[string]AliasAttr `json:"aliases"`
}

func DeleteStaleIndices(c *cli.Context) error {
	force := c.Bool("force")
	apiKey := c.String("deployment-api-key")
	cloudId := c.String("deployment-id")
	age := c.Int64("age")
	deleteList := make([]string, 0)

	hashes, err := NewHashMap()
	if err != nil {
		return err
	}

	client, err := elasticsearch.NewClient(elasticsearch.Config{APIKey: apiKey, CloudID: cloudId})
	if err != nil {
		return err
	}

	settings, err := esapi.IndicesGetSettingsRequest{FilterPath: []string{"*.settings.index.creation_date"}}.Do(context.TODO(), client)

	if err != nil {
		return err
	}

	indicesList := Indices{}
	IndicesByProject := IndicesByProject{}

	if err := json.NewDecoder(settings.Body).Decode(&indicesList); err != nil {
		return errors.Wrap(err, "Error parsing the response body")
	} else {
		for k, i := range indicesList {
			if strings.Contains(k, "elasticsearch_index") {
				// TODO: Add handling of non hash based Drupal indices.
				hash := strings.Split(k, "--")[0]
				project, err := hashes.LookupProjectFromHash(hash)
				if err != nil {
					fmt.Printf("Error looking up project for hash %+v\n", hash)
				}

				IndicesByProject[project] = append(IndicesByProject[project], map[string]IndexSettings{k: i})
			}
		}
		for p, i := range IndicesByProject {
			fmt.Printf("Scanning %s indices: \n", p)
			for _, j := range i {
				for k, i := range j {
					// TODO: Add a parameter to exclude a project by it's name.
					a, err := esapi.IndicesGetAliasRequest{Index: []string{k}}.Do(context.TODO(), client)
					if err != nil {
						return err
					}
					aliasList := map[string]Aliases{}
					if err := json.NewDecoder(a.Body).Decode(&aliasList); err != nil {
						return errors.Wrap(err, "Error parsing the response body")
					}
					now := time.Now().UnixMilli()
					created, err := strconv.ParseInt(i.IndexItem.IndexDetail.CreationDate, 10, 64)
					if err != nil {
						return err
					}

					diffInDays := (now - created) / (1000 * 60 * 60 * 24)

					if diffInDays > age {
						// Add helper function to compute hash from k.

						if len(aliasList[k].Aliases) > 0 {
							for aliasName := range aliasList[k].Aliases {
								fmt.Fprintf(c.App.Writer, "The index %s is %d days old but will not be deleted because it has an associated alias %s\n", k, diffInDays, aliasName)
							}
						} else {
							fmt.Fprintf(c.App.Writer, "The index %s is %d days old and will be marked for deletion\n", k, diffInDays)
							deleteList = append(deleteList, k)
						}
					}

				}
			}
		}
		if i := len(deleteList); i > 0 {
			if force {
				fmt.Fprint(c.App.Writer, "Deleting indices marked for deletion.")
				statusCode, err := deleteIndices(client, deleteList, i)
				if err != nil {
					return errors.Wrap(err, "error deleting indices")
				} else {
					if statusCode == 200 {
						fmt.Fprintf(c.App.Writer, "Deletion request failed. Status code %d", statusCode)
					} else {
						fmt.Fprintf(c.App.Writer, "%+v indices successfully deleted.", i)
					}
				}
			} else {
				prompt := promptui.Prompt{
					Label:     "Delete indices",
					IsConfirm: true,
				}

				prompt_result, _ := prompt.Run()

				if prompt_result == "y" {
					_, err := deleteIndices(client, deleteList, i)
					if err != nil {
						return err
					}
				} else {
					fmt.Printf("Operation cancelled.\nThere are %d indices marked for deletion.\n", i)
				}
			}
		} else {
			fmt.Printf("No indices meet the criteria for deletion.")
		}
	}
	return nil
}

func deleteIndices(client *elasticsearch.Client, deleteList []string, c int) (int, error) {
	res, err := esapi.IndicesDeleteRequest{Index: deleteList}.Do(context.TODO(), client)
	if err != nil {
		return res.StatusCode, err
	} else {
		if res.StatusCode != 200 {
			fmt.Printf("Deletion request failed. Status code %+v", res.StatusCode)
			return res.StatusCode, errors.New("non 200 status code")
		} else {
			fmt.Printf("%+v indices successfully deleted.", c)
			return res.StatusCode, nil
		}
	}
}

type HashMap struct {
	Hashes map[string]string `json:"hashes"`
}

func (h *HashMap) LookupProjectFromHash(hash string) (string, error) {
	if val, ok := h.Hashes[hash]; ok {
		return val, nil
	}
	return "", errors.New("hash not found")

}

// Compute the lookup table of hash to project name.
func NewHashMap() (HashMap, error) {
	hashMap := HashMap{
		Hashes: make(map[string]string),
	}
	client, err := helpers.NewLagoonClient(nil)
	if err != nil {
		return hashMap, err
	}
	projects, _ := getLagoonProjects(context.TODO(), client)

	for _, project := range projects {
		searchHash, _ := getLagoonProjectVar(context.TODO(), client, project.Name, "SEARCH_HASH")
		hashMap.Hashes[searchHash] = project.Name
	}
	fmt.Printf("Hashmap: %+v\n", hashMap)
	return hashMap, nil
}

// Lookup Lagoon projects
func getLagoonProjects(ctx context.Context, client *lagoon_client.Client) ([]schema.ProjectMetadata, error) {
	projects := make([]schema.ProjectMetadata, 0)

	err := client.ProjectsByMetadata(ctx, "type", "tide", &projects)
	return projects, err
}

// Lookup Lagoon projects
func getLagoonProjectVar(ctx context.Context, client *lagoon_client.Client, projectName string, varName string) (string, error) {
	vars := []schema.EnvKeyValue{}
	err := client.GetEnvVariablesByProjectEnvironmentName(ctx, &schema.EnvVariableByProjectEnvironmentNameInput{Project: projectName}, &vars)
	for _, v := range vars {
		if v.Name == varName {
			return strings.ToLower(v.Value), nil
		}
	}
	return "", err
}
