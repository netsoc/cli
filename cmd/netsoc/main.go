package main

import (
	"os"

	"github.com/netsoc/cli/pkg/cmd"
	"github.com/netsoc/cli/pkg/util"
)

func main() {
	if err := cmd.NewCmdRoot().Execute(); err != nil {
		os.Exit(1)
	}

	os.Exit(util.ExitCode)
}
