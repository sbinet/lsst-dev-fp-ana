// fp-scan scans and analyzes forced-photometry output files from the LSST stack
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"runtime/pprof"

	"github.com/gonuts/toml"
	"github.com/lsst-france/fp-ana/lsst"
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
			NewFileScanner("fscanner"),
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

	f, err := os.Create("prof.cpu")
	if err != nil {
		fmt.Printf("**error: %v\n", err)
		return 1
	}
	defer f.Close()

	pprof.StartCPUProfile(f)
	defer pprof.StopCPUProfile()

	err = app.Run()
	if err != nil {
		fmt.Printf("**error: %v\n", err)
		return 1
	}

	return 0
}
