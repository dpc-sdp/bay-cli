package elastic_cloud

import (
	"context"
	"encoding/json"
	"fmt"

	elasticsearch "github.com/elastic/go-elasticsearch/v8"
	"github.com/elastic/go-elasticsearch/v8/esapi"
	errors "github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

type Shard struct {
	Index     string `json:"index"`
	ShardName string `json:"shard"`
	State     string `json:"state"`
}

type Shards []Shard

func ListUnassignedShards(c *cli.Context) error {
	apiKey := c.String("deployment-api-key")
	cloudId := c.String("deployment-id")
	client, err := elasticsearch.NewClient(elasticsearch.Config{APIKey: apiKey, CloudID: cloudId})
	if err != nil {
		return err
	}
	shards, err := esapi.CatShardsRequest{Format: "json", FilterPath: []string{"index", "shard", "state"}}.Do(context.Background(), client)

	shardsList := Shards{}
	unassignedShards := Shards{}

	if err := json.NewDecoder(shards.Body).Decode(&shardsList); err != nil {
		return errors.Wrap(err, "Error parsing the response body")
	} else {
		for _, s := range shardsList {
			if s.State == "UNASSIGNED" {
				unassignedShards = append(unassignedShards, s)
			}
		}

		if err != nil {
			return err
		}

		json, _ := json.Marshal(unassignedShards)
		fmt.Printf("%s", json)

		return nil
	}
}
