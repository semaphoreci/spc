package whencli

import (
	"fmt"
	"os/exec"

	gabs "github.com/Jeffail/gabs/v2"
)

func ListInputs(expression string) (*gabs.Container, error) {
	bytes, err := exec.Command("when", "list-inputs", expression).Output()
	if err != nil {
		return nil, fmt.Errorf("Unprecessable when expression %s", string(bytes))
	}

	result, err := gabs.ParseJSON(bytes)
	if err != nil {
		return nil, fmt.Errorf("Unprocessable input list for when expressions %s", err.Error())
	}

	return result, nil
}
