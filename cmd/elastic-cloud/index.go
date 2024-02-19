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

	"github.com/dpc-sdp/bay-cli/internal/helpers"
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
	args := make([]string, 0)
	args = helpers.GetAllArgs(c)

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
		// filter list for indices by name `*--elasticsearch_index_*`
		if dryRun {
			fmt.Println("this is a dry-run")
			fmt.Printf("%+v", args)
		}
		for k, i := range list {
			if strings.Contains(k, "elasticsearch_index") {
				now := time.Now().UnixMilli()
				created, err := strconv.ParseInt(i.IndexItem.IndexDetail.CreationDate, 10, 64)
				if err != nil {
					panic(err)
				}

				// for _, a := range args {
				// }

				diffInDays := (now - created) / (1000 * 60 * 60 * 24)
				if diffInDays > 30 {
					fmt.Printf("%+v - %+v diff is %v\n", k, i.IndexItem.IndexDetail.CreationDate, diffInDays)
				}
			}
		}
	}

	return nil
}
