package git

import (
	"fmt"
	"os/exec"
	"strings"
)

func Fetch(name string) ([]byte, error) {
	fmt.Printf("Fetching branch from remote: '%s'\n", name)

	flags := []string{"fetch", "origin", fmt.Sprintf("+refs/heads/%s:refs/heads/%s", name, name)}
	fmt.Printf("Running git %s\n", strings.Join(flags, " "))

	return exec.Command("git", flags...).CombinedOutput()
}
