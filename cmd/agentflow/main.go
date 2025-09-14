package main

import (
	"fmt"
	"os"

	"agentflow/internal/cli"
	"github.com/spf13/viper"
)

func main() {
	v := viper.New()
	cmd := cli.NewRootCmd(v)
	if err := cmd.Execute(); err != nil {
		// Print a friendly error and exit non-zero
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
