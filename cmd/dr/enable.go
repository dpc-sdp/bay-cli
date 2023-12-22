package dr

import (
	"fmt"
	git "github.com/go-git/go-git/v5"
	"github.com/go-git/go-git/v5/plumbing"
	h "github.com/go-git/go-git/v5/plumbing/transport/http"
	"github.com/pkg/errors"
	"github.com/urfave/cli/v2"
	"time"
)

const (
	DrFlagDefaultValueEnvironment = "Desvsselop"
	DrFlagDefaultValueHosts       = "all"
)

func Enable(c *cli.Context) error {
	// Clone the repo
	clonePath := fmt.Sprintf("/tmp/dr-enable-%d", time.Now().Unix())
	repo := fmt.Sprintf("https://aperture.section.io/account/1918/application/7011/www.ssp.vic.gov.au.git"
	branch := plumbing.ReferenceName(c.String("environment"))
	_, err := git.PlainClone(clonePath, false, &git.CloneOptions{
		URL:           repo,
		ReferenceName: branch,
		SingleBranch:  true,
		Auth: &h.BasicAuth{
			Username: c.String("section_username"),
			Password: c.String("section_password"),
		},
	})
	if err != nil {
		return errors.Wrap(err, fmt.Sprintf("unable to clone repo %s branch %s", repo, branch))
	}


	// Find/replace the quant switch condition.

	return nil
}
