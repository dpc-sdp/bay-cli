package elastic_cloud

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	elasticsearch "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/manifoldco/promptui"
	errors "github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

type IndexSettings struct {
	IndexItem struct {
		IndexDetail struct {
			CreationDate string `json:"creation_date"`
		} `json:"index"`
	} `json:"settings"`
}

type Indices map[string]IndexSettings

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

	client, err := elasticsearch.NewClient(elasticsearch.Config{APIKey: apiKey, CloudID: cloudId})
	if err != nil {
		return err
	}

	settings, err := esapi.IndicesGetSettingsRequest{FilterPath: []string{"*.settings.index.creation_date"}}.Do(context.TODO(), client)

	if err != nil {
		return err
	}

	indicesList := Indices{}

	if err := json.NewDecoder(settings.Body).Decode(&indicesList); err != nil {
		return errors.Wrap(err, "Error parsing the response body")
	} else {
		for k, i := range indicesList {
			if strings.Contains(k, "elasticsearch_index") {
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
					if len(aliasList[k].Aliases) > 0 {
						for aliasName, _ := range aliasList[k].Aliases {
							fmt.Fprintf(c.App.Writer, "The index %+v is %v days old but will not be deleted because it has an associated alias %+v\n", k, diffInDays, aliasName)
						}
					} else {
						fmt.Fprintf(c.App.Writer, "The index %+v is %v days old and will be marked for deletion\n", k, diffInDays)
						deleteList = append(deleteList, k)

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
						fmt.Fprintf(c.App.Writer, "Deletion request failed. Status code %+v", statusCode)
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
					fmt.Printf("Operation cancelled.\nThere are %+v indices marked for deletion.\n", i)
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
