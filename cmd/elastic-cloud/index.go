package elastic_cloud

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	"github.com/urfave/cli/v2"

	envconfig "github.com/sethvargo/go-envconfig"
	ctrl "sigs.k8s.io/controller-runtime"
)

// Config for Elasticsearch client
type EsConfig struct {
	ApiKey  string `env:"EC_API_KEY"`
	CloudId string `env:"EC_CLOUD_ID"`
}

type IndexSettings struct {
	IndexItem struct {
		IndexDetail struct {
			CreationDate string `json:"creation_date"`
		} `json:"index"`
	} `json:"settings"`
}

type Indices map[string]IndexSettings

var setupLog = ctrl.Log.WithName("setup")

func GetIndex(c *cli.Context) error {
	dryRun := c.Bool("dry-run")
	deleteList := make([]string, 0)

	var config EsConfig
	if err := envconfig.Process(context.Background(), &config); err != nil {
		setupLog.Error(err, "unable to parse environment variables")
		os.Exit(1)
	}

	client, err := elasticsearch.NewClient(elasticsearch.Config{APIKey: config.ApiKey, CloudID: config.CloudId})

	if err != nil {
		return err
	}

	settings, _ := esapi.IndicesGetSettingsRequest{FilterPath: []string{"*.settings.index.creation_date"}}.Do(context.TODO(), client)

	var list Indices

	if err := json.NewDecoder(settings.Body).Decode(&list); err != nil {
		log.Fatalf("Error parsing the response body: %s", err)
	} else {
		for k, i := range list {
			if strings.Contains(k, "elasticsearch_index") {
				now := time.Now().UnixMilli()
				created, err := strconv.ParseInt(i.IndexItem.IndexDetail.CreationDate, 10, 64)
				if err != nil {
					return err
				}

				diffInDays := (now - created) / (1000 * 60 * 60 * 24)
				if diffInDays > 30 {
					fmt.Printf("The index %+v is %v days old and will be marked for deletion\n", k, diffInDays)
					deleteList = append(deleteList, k)
				}
			}
		}
		if c := len(deleteList); c > 0 {
			if !dryRun {
				fmt.Println("Deleting indices marked for deletion.")
				_, err := esapi.IndicesDeleteRequest{Index: deleteList}.Do(context.TODO(), client)
				if err != nil {
					return err
				} else {
					fmt.Printf("%+v indices successfully deleted.", c)
				}
			} else {
				fmt.Printf("The 'dry-run' flag is set - no further action taken. There are %+v indices marked for deletion", c)
			}
		}
	}
	return nil
}
