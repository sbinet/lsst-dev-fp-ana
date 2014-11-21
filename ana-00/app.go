package main

import (
	"fmt"
	"os"
	"sync"

	fits "github.com/astrogo/fitsio"
)

type App struct {
	workers []Worker
	data    chan Result
	fname   string
	wg      *sync.WaitGroup
	done    chan struct{}
}

func New(fname string) App {
	var wg sync.WaitGroup
	return App{
		workers: make([]Worker, 0, 2),
		data:    make(chan Result),
		fname:   fname,
		wg:      &wg,
		done:    make(chan struct{}),
	}
}

func (app *App) AddWorker(fname, wtype string) error {
	var err error
	switch wtype {
	case "lsst", "LSST":
		wrk, err := NewLSSTWorker(fname, app.data, app.wg)
		if err != nil {
			return err
		}
		app.wg.Add(1)
		app.workers = append(app.workers, wrk)
	case "sdss", "SDSS":
		wrk, err := NewSDSSWorker(fname, app.data, app.wg)
		if err != nil {
			return err
		}
		app.wg.Add(1)
		app.workers = append(app.workers, wrk)
	}
	return err
}

func (app *App) Run() error {
	var err error

	_ = os.Remove(app.fname)
	fmt.Printf("--- output: [%s]...\n", app.fname)
	w, err := os.Create(app.fname)
	if err != nil {
		return err
	}
	defer w.Close()

	f, err := fits.Create(w)
	if err != nil {
		return err
	}
	defer f.Close()

	phdu, err := fits.NewPrimaryHDU(nil)
	if err != nil {
		return err
	}

	err = f.Write(phdu)
	if err != nil {
		return err
	}

	otbl, err := fits.NewTableFrom("astro", Data{}, fits.BINARY_TBL)
	if err != nil {
		return err
	}

	go app.collect(otbl)

	fmt.Printf("--- start workers...\n")
	for _, wrk := range app.workers {
		err = wrk.Start()
		if err != nil {
			return err
		}
	}

	fmt.Printf("--- run workers...\n")
	for _, wrk := range app.workers {
		err = wrk.Run()
		if err != nil {
			return err
		}
	}

	app.wg.Wait()
	close(app.data)
	<-app.done

	fmt.Printf("--- stop workers...\n")
	for _, wrk := range app.workers {
		err = wrk.Stop()
		if err != nil {
			fmt.Printf("*** error stop worker: %v\n", err)
			return err
		}
	}

	fmt.Printf("--- write out results...\n")
	err = f.Write(otbl)
	if err != nil {
		fmt.Printf("*** error: %v\n", err)
		return err
	}

	err = otbl.Close()
	if err != nil {
		return err
	}

	err = f.Close()
	if err != nil {
		return err
	}

	// err = w.Close()
	// if err != nil {
	// 	return err
	// }

	return err
}

func (app *App) collect(table *fits.Table) {
	for res := range app.data {
		if res.err != nil {
			panic(res.err)
		}
		err := table.Write(&res.out)
		if err != nil {
			panic(err)
		}
	}

	fmt.Printf("--- stop collector\n")
	app.done <- struct{}{}
}
