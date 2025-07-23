package deployment

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/go-git/go-git/v5"
	errors "github.com/pkg/errors"
	"github.com/urfave/cli/v3"
)

type Metadata struct {
	DeploymentMetadata Deployment `json:"deployment"`
}

type Deployment struct {
	Sha        string `json:"sha"`
	AuthorName string `json:"authorName"`
	When       string `json:"when"`
	Tag        string `json:"tag"`
	Msg        string `json:"msg"`
}

func DeploymentMetadata(ctx context.Context, c *cli.Command) error {
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

	tagString := ""

	tag, err := repo.Tag("HEAD")

	if err != nil {
		tagString = "No tag found"
	} else {
		tagString = tag.Name().Short()
	}

	msgFirstLn := strings.TrimLeft(strings.Split(msg.String(), "\n")[4], " ")

	item := Deployment{
		Sha:        ref.Hash().String(),
		AuthorName: msg.Author.Name,
		When:       msg.Author.When.String(),
		Msg:        msgFirstLn,
		Tag:        tagString,
	}

	md := Metadata{
		DeploymentMetadata: item,
	}

	json, _ := json.Marshal(md)

	// Write string to stdout
	fmt.Fprintf(c.Writer, "%s", json)

	return nil
}
