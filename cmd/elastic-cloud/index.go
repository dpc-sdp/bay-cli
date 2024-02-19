package elastic_cloud

import (
	"context"
	"encoding/json"
	"fmt"
	"os"

	"github.com/urfave/cli/v2"

	"github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	envconfig "github.com/sethvargo/go-envconfig"
	ctrl "sigs.k8s.io/controller-runtime"
)

// Config for Elasticsearch client
type EsConfig struct {
	ApiKey  string `env:"EC_API_KEY"`
	CloudId string `env:"EC_CLOUD_ID"`
}

type Indices struct {
	Index map[string]interface{} `json:"Index"`
}

var setupLog = ctrl.Log.WithName("setup")

func GetIndex(c *cli.Context) error {
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

	fmt.Printf(settings.String())

	var list Indices

	jsonErr := json.Unmarshal([]byte(settings.String()), &list)

	if jsonErr != nil {
		fmt.Println("Error:", jsonErr)
		return jsonErr
	}

	fmt.Printf("%+v\n", list)

	// use struct to get JSON content

	// filter indices by name `*--elasticsearch_index_*`

	return nil
}
