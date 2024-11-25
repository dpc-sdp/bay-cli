package project_map

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5"
	errors "github.com/pkg/errors"
	"github.com/urfave/cli/v2"
)

type Metadata struct {
	Deployment Item `json:"deployment"`
}

type Item struct {
	Sha        string `json:"sha"`
	AuthorName string `json:"authorName"`
	When       string `json:"when"`
	Tag        string `json:"tag"`
	Msg        string `json:"msg"`
}

func DeploymentMetadata(c *cli.Context) error {
	// logger := log.New(c.App.ErrWriter, "", log.LstdFlags)

	repo, err := git.PlainOpenWithOptions(".", &git.PlainOpenOptions{DetectDotGit: true})
	if err != nil {
		return errors.Wrap(err, "unable to open git repository")
	}

	ref, err := repo.Head()
	if err != nil {
		return errors.Wrap(err, "unable to get HEAD reference")
	}

	msg, err := repo.CommitObject(ref.Hash())
	if err != nil {
		return errors.Wrap(err, "unable to get commit object")
	}

	tag, _ := repo.Tag("HEAD")

	msgFirstLn := strings.TrimLeft(strings.Split(msg.String(), "\n")[4], " ")

	item := Item{
		Sha:        fmt.Sprintf("%s", ref.Hash()),
		AuthorName: fmt.Sprintf("%s", msg.Author.Name),
		When:       fmt.Sprintf("%s", msg.Author.When),
		Msg:        fmt.Sprintf("%s", msgFirstLn),
		Tag:        fmt.Sprintf("%s", tag),
	}

	md := Metadata{
		Deployment: item,
	}

	json, _ := json.Marshal(md)

	// Write string to stdout
	fmt.Fprintf(c.App.Writer, "%s", json)

	return nil
}
