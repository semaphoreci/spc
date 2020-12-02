package main

import (
	"fmt"
	"os"

	cli "github.com/semaphoreci/spc/pkg/cli"
)

func main() {
	err := cli.Execute()

	if err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}
