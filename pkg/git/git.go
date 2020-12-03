package git

import (
	"fmt"
	"os/exec"
	"strings"
)

func Fetch(name string) ([]byte, error) {
	flags := []string{"fetch", "origin", fmt.Sprintf("+refs/heads/%s:refs/heads/%s", name, name)}
	fmt.Printf("Running git %s\n", strings.Join(flags, " "))

	return exec.Command("git", flags...).CombinedOutput()
}

func Diff(commitRange string) ([]string, string, error) {
	flags := []string{"diff", "--name-only", commitRange}
	fmt.Printf("Running git %s\n", strings.Join(flags, " "))

	bytes, err := exec.Command("git", flags...).CombinedOutput()
	if err != nil {
		return []string{}, string(bytes), err
	}

	return strings.Split(strings.TrimSpace(string(bytes)), "\n"), "", nil
}
