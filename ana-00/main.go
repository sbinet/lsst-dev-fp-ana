package main

import (
	"flag"
	"fmt"
	"math"
	"path/filepath"
	"sort"
)

const (
	rad2deg = 180.0 / math.Pi
)

var (
	// catalog = flag.String("type", "sdss", "type of the input catalog (lsst|sdss)")
	ofname = flag.String("o", "out.fits", "path to output FITS file")
)

func main() {
	flag.Parse()

	// ifname := "/sps/lsst/data/astrometry_net_data/sdss-dr9-raw/sdss-dr9-fink-v5b/astromSweeps-8162.fits"
	// ifname := "/afs/in2p3.fr/home/l/lsstprod/prod/DC_2014/test_cfht/output/icSrc/06AL01/D3/2006-06-02/r/ICSRC-850587-00.fits"
	// if flag.NArg() > 0 {
	// 	ifname = flag.Arg(0)
	// }

	app := New(*ofname)

	cfg := Config{
		Type: "lsst",
		Dir:  "/afs/in2p3.fr/home/l/lsstprod/prod/DC_2014/test_cfht/output/icSrc/06AL01/D3/2006-06-02/r",
	}
	fnames := make([]string, 0, 36)
	for i := 0; i < 36; i++ {
		ff, err := filepath.Glob(cfg.Dir + fmt.Sprintf("/*-%02d.fits", i))
		if err != nil {
			panic(err)
		}
		if len(ff) != 1 {
			panic(fmt.Errorf("invalid number of files. got=%d. want=1", len(ff)))
		}
		fnames = append(fnames, ff[0])
	}
	sort.Strings(fnames)

	for _, fname := range fnames {
		err := app.AddWorker(fname, cfg.Type)
		if err != nil {
			panic(err)
		}
	}

	err := app.Run()
	if err != nil {
		panic(err)
	}

}
