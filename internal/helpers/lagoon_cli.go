package helpers

import (
	"fmt"
	"io/ioutil"
	"os"
	"strings"

	lagoon_client "github.com/uselagoon/machinery/api/lagoon/client"
	"golang.org/x/crypto/ssh"
	yaml "gopkg.in/yaml.v3"
)

// Ripped from https://github.com/uselagoon/lagoon-cli/blob/main/internal/lagoon/config.go

// LagoonConfig is used for the lagoon configuration.
type LagoonConfig struct {
	Current string                   `yaml:"current"`
	Default string                   `yaml:"default"`
	Lagoons map[string]LagoonContext `yaml:"lagoons"`
}

// LagoonContext is used for each lagoon context in the config file.
type LagoonContext struct {
	GraphQL   string `yaml:"graphql"`
	HostName  string `yaml:"hostname"`
	UI        string `yaml:"ui,omitempty"`
	Kibana    string `yaml:"kibana,omitempty"`
	Port      string `yaml:"port"`
	Token     string `yaml:"token,omitempty"`
	Version   string `yaml:"version,omitempty"`
	SSHKey    string `yaml:"sshkey,omitempty"`
	SSHPortal bool   `yaml:"sshPortal,omitempty"`
}

func GetLagoonConfigContext(context *string) (*LagoonContext, error) {
	homeDir, err := os.UserHomeDir()
	if err != nil {
		return nil, err
	}

	yamlPath := fmt.Sprintf("%s/.lagoon.yml", homeDir)
	yamlData, err := ioutil.ReadFile(yamlPath)
	if err != nil {
		return nil, err
	}

	Config := LagoonConfig{}
	err = yaml.Unmarshal(yamlData, &Config)
	if err != nil {
		return nil, err
	}

	if context == nil {
		context = &Config.Current
	}

	lagoonContext := Config.Lagoons[*context]
	return &lagoonContext, nil
}

func ConvertPrivateKeyToPublic(privateKey string) (string, error) {
	signer, err := ssh.ParsePrivateKey([]byte(privateKey))
	if err != nil {
		return "", err
	}
	publicKey := signer.PublicKey()
	return strings.TrimSuffix(string(ssh.MarshalAuthorizedKey(publicKey)), "\n"), nil
}

func NewLagoonClient(context *string) (*lagoon_client.Client, error) {
	lagoonContext, err := GetLagoonConfigContext(context)
	if err != nil {
		return nil, err
	}

	client := lagoon_client.New(lagoonContext.GraphQL, "github.com/dpc-sdp/bay-cli", &lagoonContext.Token, false)
	return client, nil
}
