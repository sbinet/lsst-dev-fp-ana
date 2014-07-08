package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/gonuts/toml"
	"github.com/lsst-france/fps-ana/lsst"
)

var (
	g_config = flag.String("jobo", "jobo.toml", "job configuration file")
)

func main() {

	flag.Parse()

	fmt.Printf("=== %s ===\n", filepath.Base(os.Args[0]))
	rc := run()

	os.Exit(rc)
}

func run() int {
	var err error
	app := lsst.App{
		Procs: []lsst.P{
			NewListBuilder("listbuilder"),
		},
	}

	var jobo lsst.FileOptions
	if *g_config != "" {
		_, err = toml.DecodeFile(*g_config, &jobo)
		if err != nil {
			fmt.Printf("**error: %v\n", err)
			return 1
		}
	}

	err = app.Configure(jobo)
	if err != nil {
		fmt.Printf("**error: %v\n", err)
		return 1
	}

	err = app.Run()
	if err != nil {
		fmt.Printf("**error: %v\n", err)
		return 1
	}

	return 0
}
