package main

import (
	"os"

	"github.com/rassu/dbbeat/cmd"

	_ "github.com/rassu/dbbeat/include"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
