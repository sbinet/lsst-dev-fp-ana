package main

import (
	"fmt"
	"io"
	"os"
	"sync"

	fits "github.com/astrogo/fitsio"
)

type Result struct {
	out Data
	err error
}

type Worker interface {
	Start() error
	Run() error
	Stop() error
}

type sdssWorker struct {
	fname string // input file name to analyze
	r     io.ReadCloser
	f     *fits.File
	tbl   *fits.Table

	data chan Result
	wg   *sync.WaitGroup
}

func NewSdssWorker(fname string, ch chan Result, wg *sync.WaitGroup) (Worker, error) {
	wrk := &sdssWorker{
		fname: fname,
		data:  ch,
		wg:    wg,
	}
	return wrk, nil
}

func (wrk *sdssWorker) Start() error {
	var err error

	r, err := os.Open(wrk.fname)
	if err != nil {
		return err
	}
	wrk.r = r

	f, err := fits.Open(r)
	if err != nil {
		return err
	}
	wrk.f = f

	wrk.tbl = wrk.f.HDU(1).(*fits.Table)
	return err
}

func (wrk *sdssWorker) Run() error {
	var err error
	nrows := wrk.tbl.NumRows()
	fmt.Printf("=== file [%s]...\n", wrk.fname)
	fmt.Printf(">> nrows=%d\n", nrows)
	rows, err := wrk.tbl.Read(0, nrows)
	if err != nil {
		return err
	}
	defer rows.Close()
	defer wrk.wg.Done()

	for i := 0; rows.Next(); i++ {
		var data SdssData
		err = rows.Scan(&data)
		if err != nil {
			wrk.data <- Result{err: err}
			return err
		}

		// convert radians to degrees
		ra := data.Ra * rad2deg
		dec := data.Dec * rad2deg

		wrk.data <- Result{
			out: Data{
				ID:  data.ID,
				Ra:  ra,
				Dec: dec,
			},
		}

	}

	return err
}

func (wrk *sdssWorker) Stop() error {
	var err error
	defer wrk.r.Close()
	defer wrk.f.Close()

	err = wrk.tbl.Close()
	if err != nil {
		fmt.Printf("*** error: [%T] - table %v\n", wrk, err)
		return err
	}

	err = wrk.f.Close()
	if err != nil {
		fmt.Printf("*** error: [%T] - fits-file %v\n", wrk, err)
		return err
	}

	// err = wrk.r.Close()
	// if err != nil {
	// 	fmt.Printf("*** error: [%T] - os-file %v\n", wrk, err)
	// 	return err
	// }

	return err
}
