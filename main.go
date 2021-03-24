package main

import (
	"os"

	"github.com/ronaudinho/dbbeat/cmd"

	_ "github.com/ronaudinho/dbbeat/include"
)

func main() {
	if err := cmd.RootCmd.Execute(); err != nil {
		os.Exit(1)
	}
}
