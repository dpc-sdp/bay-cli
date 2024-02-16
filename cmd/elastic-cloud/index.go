package elastic_cloud

import (
	"context"
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
	Index string
}

var setupLog = ctrl.Log.WithName("setup")

func GetIndex(c *cli.Context) error {
	var config EsConfig
	if err := envconfig.Process(context.Background(), &config); err != nil {
		setupLog.Error(err, "unable to parse environment variables")
		os.Exit(1)
	}

	// client, err := elasticsearch.NewClient(elasticsearch.Config{esConfig})
	client, err := elasticsearch.NewClient(elasticsearch.Config{APIKey: config.ApiKey, CloudID: config.CloudId})

	if err != nil {
		return err
	}

	// settings, _ := esapi.ClusterGetSettingsRequest{Human: true, FilterPath: []string{"*.settings.index.creation_date_string"}}.Do(context.TODO(), client)
	settings, _ := esapi.IndicesGetSettingsRequest{FilterPath: []string{"*.settings.index.creation_date"}}.Do(context.TODO(), client)

	// h := []string{"index"}
	// res, _ := esapi.CatIndicesRequest{Format: "JSON", H: []string{"index"}, S: []string{"index"}}.Do(context.TODO(), client)

	// input := esapi.CatIndicesRequest{Format: "JSON"}

	// indices, err := client.Cat.Indices(&input)

	// indicesBody, _ := io.ReadAll(indices.Body)

	fmt.Printf(settings.String())

	return nil
}
