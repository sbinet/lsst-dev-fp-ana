package main

import (
	"flag"
	"math"
)

const (
	rad2deg = 180.0 / math.Pi
)

var (
	catalog = flag.String("type", "sdss", "type of the input catalog (lsst|sdss)")
	ofname  = flag.String("o", "out.fits", "path to output FITS file")
)

func main() {
	flag.Parse()

	ifname := "/sps/lsst/data/astrometry_net_data/sdss-dr9-raw/sdss-dr9-fink-v5b/astromSweeps-8162.fits"
	if flag.NArg() > 0 {
		ifname = flag.Arg(0)
	}

	app := New(*ofname)
	err := app.AddWorker(ifname, *catalog)
	if err != nil {
		panic(err)
	}

	err = app.Run()
	if err != nil {
		panic(err)
	}

}
