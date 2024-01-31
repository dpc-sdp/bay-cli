package template

import (
	"fmt"
	"regexp"
)

func ConvertGithubUriToPage(input string) string {
	re := regexp.MustCompile(`^git@github\.com:([\w-]+)/([\w-]+)\.git$`)
	matches := re.FindStringSubmatch(input)
	if len(matches) > 2 {
		org := matches[1]
		repo := matches[2]
		return fmt.Sprintf("https://github.com/%s/%s", org, repo)
	} else {
		return input
	}
}
