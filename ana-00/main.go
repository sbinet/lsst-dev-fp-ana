package main

import (
	"fmt"
	"math"
	"os"

	fits "github.com/astrogo/fitsio"
)

const (
	rad2deg = 180.0 / math.Pi
)

type SdssData struct {
	ID  int64   `fits:"ID"`
	OID int64   `fits:"thing_id"`
	Ra  float64 `fits:"RA"`
	Dec float64 `fits:"DEC"`
}

type Data struct {
	ID       int64      `fits:"id"`
	OID      int64      `fits:"objectId"`
	ParentID int64      `fits:"parent"`
	Coord    [2]float64 `fits:"coord"`

	Flux    float64 `fits:"flux_psf"`
	RefFlux float64 `fits:"refFlux"`
}

type OutData struct {
	ID      int64      `fits:"id"`
	OID     int64      `fits:"thing_id"`
	RaDec   [2]float64 `fits:"radec"`
	AssocID int64      `fits:"assoc_id"`
}

type RefType struct {
	ID  int64
	OID int64
}

type Coord [2]float64

var (
	coords = make(map[Coord][]RefType)
)

func main() {
	ifname := "/sps/lsst/data/astrometry_net_data/sdss-dr9-raw/sdss-dr9-fink-v5b/astromSweeps-8162.fits"
	if len(os.Args) > 1 {
		ifname = os.Args[1]
	}

	r, err := os.Open(ifname)
	if err != nil {
		panic(err)
	}
	defer r.Close()

	fin, err := fits.Open(r)
	if err != nil {
		panic(err)
	}
	defer fin.Close()

	ofname := "out.fits"
	if len(os.Args) > 2 {
		ifname = os.Args[2]
	}

	w, err := os.Create(ofname)
	if err != nil {
		panic(err)
	}
	defer w.Close()

	fout, err := fits.Create(w)
	if err != nil {
		panic(err)
	}

	phdu, err := fits.NewPrimaryHDU(nil)
	if err != nil {
		panic(err)
	}

	err = fout.Write(phdu)
	if err != nil {
		panic(err)
	}

	otbl, err := fits.NewTableFrom("astro", OutData{}, fits.BINARY_TBL)
	if err != nil {
		panic(err)
	}

	table := fin.HDU(1).(*fits.Table)
	defer table.Close()

	nrows := table.NumRows()
	fmt.Printf(">>> nrows=%d\n", nrows)

	rows, err := table.Read(0, nrows)
	if err != nil {
		panic(err)
	}
	defer rows.Close()

	for i := 0; rows.Next(); i++ {
		var data SdssData

		err = rows.Scan(&data)
		if err != nil {
			fmt.Printf(">>> row=%d: %v\n", i, err)
			panic(err)
		}

		if i%10 == 0 && false {
			fmt.Printf(">>> %#v\n", data)
		}

		// convert radians to degrees
		ra := data.Ra * rad2deg
		dec := data.Dec * rad2deg

		odata := OutData{
			ID:    data.ID,
			OID:   data.OID,
			RaDec: [2]float64{ra, dec},
		}

		radec := Coord{ra, dec}
		coords[radec] = append(coords[radec], RefType{data.ID, data.OID})

		err = otbl.Write(&odata)
		if err != nil {
			panic(err)
		}
	}

	err = otbl.Close()
	if err != nil {
		panic(err)
	}

	err = fout.Close()
	if err != nil {
		panic(err)
	}

	fmt.Printf("processed %d rows\n", nrows)
	fmt.Printf("stored %d coords\n", len(coords))
}
