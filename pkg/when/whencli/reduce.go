package whencli

import (
	"fmt"
	"io/ioutil"
	"os"
	"os/exec"
	"strings"
)

func Reduce(expression string, inputs []byte) (string, error) {
	path := "/tmp/input.json"

	fmt.Printf("Providing inputs: %s\n", string(inputs))

	err := ioutil.WriteFile(path, inputs, os.ModePerm)
	if err != nil {
		return "", err
	}

	bytes, err := exec.Command("when", "reduce", expression, "--input", path).Output()
	if err != nil {
		return "", err
	}

	return strings.TrimSpace(string(bytes)), nil
}
