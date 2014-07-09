package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"

	fits "github.com/astrogo/cfitsio"
	"github.com/lsst-france/fp-ana/lsst"
)

const (
	rad2deg = 180.0 / math.Pi
)

type listbuilder struct {
	*lsst.Processor

	Measures []lsst.FPMeasures
	FilterDb map[int]int
	Filters  []int

	NbObjects     int
	NbMeasures    int
	NbBadMeasures int
	NbMeasuresIn  int

	//NbErrOID   int
	NbErrRaDec int
}

func NewListBuilder(name string) lsst.P {
	ctx := &listbuilder{
		Processor: lsst.NewProcessor(name),
		FilterDb:  make(map[int]int),
		Filters:   []int{},
	}

	ctx.Config = ctx.config
	ctx.Start = ctx.start
	ctx.Proc = ctx.proc
	ctx.Stop = ctx.stop

	return ctx
}

func (proc *listbuilder) config(opts lsst.Options) error {
	var err error
	cfg, ok := opts.(lsst.FileOptions)
	if !ok {
		return err
	}

	for _, sfilter := range cfg.Filters {
		filter := byte(sfilter[0])
		proc.Filters = append(proc.Filters, lsst.FilterID(filter))
	}

	return err
}

func (proc *listbuilder) start() error {
	var err error

	for i, filter := range proc.Filters {
		proc.Infof(">>> filter=%d\n", filter)
		if filter < 1 {
			continue
		}
		proc.FilterDb[filter] = i
	}

	for i := 0; i < proc.RaDec.NbRa*proc.RaDec.NbDec; i++ {
		proc.Measures = append(proc.Measures, make(lsst.FPMeasures))
	}

	proc.Infof("filter-db: %v\n", proc.FilterDb)
	proc.Infof("measures:  %d\n", len(proc.Measures))
	proc.Infof("nfilters:  %d\n", len(proc.Filters))
	proc.Infof("radec:     %#v\n", proc.RaDec)

	return err
}

func (proc *listbuilder) proc(f lsst.File) error {
	var err error
	//proc.Infof(">>> file=%#v\n", f)

	fid, ok := proc.FilterDb[lsst.FilterID(f.Filter)]
	if !ok {
		proc.Errorf("filter-id for [%s] not found in filter-db\n", string(f.Filter))
		return err
	}
	proc.Infof("filter-id: %d (%s)\n", fid, string(f.Filter))

	ff, err := fits.Open(f.Name, fits.ReadOnly)
	if err != nil {
		return err
	}
	defer ff.Close()

	table := ff.HDU(1).(*fits.Table)
	nrows := table.NumRows()

	if nrows < 1 {
		proc.Errorf("file run=%d field=%d camcol=%d filter=%s == nrows=0\n",
			f.Run, f.Field, f.CamCol, f.Filter,
		)
		return fmt.Errorf("no data")
	}

	rows, err := table.Read(0, nrows)
	if err != nil {
		panic(err)
	}
	defer rows.Close()
	for rows.Next() {

		data := struct {
			ID      int64      `fits:"id"`
			OID     int64      `fits:"objectId"`
			Flux    float64    `fits:"flux_psf"`
			RefFlux float64    `fits:"refFlux"`
			Coord   [2]float64 `fits:"coord"`
		}{}

		err = rows.Scan(&data)
		if err != nil {
			return err
		}

		//fmt.Printf(">>> %v\n", data)

		id := data.ID
		oid := data.OID
		flx := data.Flux
		refflx := data.RefFlux
		// convert radians to degrees
		ra := data.Coord[0] * rad2deg
		dec := data.Coord[1] * rad2deg

		err = proc.updatelst(fid, id, oid, ra, dec, flx, refflx)
		if err != nil {
			return err
		}
	}

	return err
}

func (proc *listbuilder) updatelst(fid int, id, oid int64, ra, dec, flx, refflx float64) error {
	var err error
	proc.NbMeasures += 1

	if math.IsInf(flx, 0) || math.IsNaN(flx) {
		proc.NbBadMeasures += 1
		//proc.Debugf("invalid flux value = %v\n", flx)
		// err = fmt.Errorf("invalid flux value")
		return err
	}

	kdec := int((dec - proc.RaDec.Min.Dec) / float64(proc.RaDec.DeltaDec))
	kra := int((ra - proc.RaDec.Min.Ra) / float64(proc.RaDec.DeltaRa))

	// check whether we are indeed in the alpha/delta selected zone
	if kra < 0 || kdec < 0 || kra >= proc.RaDec.NbRa || kdec >= proc.RaDec.NbDec {
		return err
	}

	proc.NbMeasuresIn += 1

	rdidx := kdec*proc.RaDec.NbRa + kra
	measures := &proc.Measures[rdidx]
	measure, ok := (*measures)[oid]
	if !ok {
		// adding a new source / object
		measure = lsst.FPMeasure{
			ID:     id,
			OID:    oid,
			RaDec:  lsst.RaDec{Ra: ra, Dec: dec},
			Fluxes: make([]lsst.FluxRec, len(proc.FilterDb)),
		}
		measure.Add(fid, flx)
		proc.NbObjects += 1
	} else {
		const delta = 2. / 3600.
		if math.Abs(measure.RaDec.Ra-ra) > delta ||
			math.Abs(measure.RaDec.Dec-dec) > delta {
			proc.NbErrRaDec += 1
		}
		measure.Add(fid, flx)
	}
	(*measures)[oid] = measure

	return err
}

func (proc *listbuilder) stop() error {
	var err error
	proc.Infof("--- list-builder stats ---\n")
	proc.Infof(" #measures:  %d\n", proc.NbMeasures)
	proc.Infof(" #bad-meas:  %d\n", proc.NbBadMeasures)
	proc.Infof(" #meas-in:   %d\n", proc.NbMeasuresIn)
	proc.Infof(" #objects:   %d\n", proc.NbObjects)
	proc.Infof(" #err-radec: %d\n", proc.NbErrRaDec)

	fname := filepath.Join(proc.OutputDir, "srclist.txt")
	proc.Infof("saving object/source list to [%s]\n", fname)
	fout, err := os.Create(fname)
	if err != nil {
		return err
	}
	defer fout.Close()

	nsrc := 0
	fmt.Fprintf(fout, "## id oid ra dec flx-mean-1 flx-sigma-1 nmes-1 flx-mean-2 ...\n")
	// loop over cells in alpha/delta
	for i, measures := range proc.Measures {
		//measures := &proc.Measures[idx]
		proc.Debugf(" ra-dec-cell[%03d] ra,dec=(%+8.3f, %+8.3f) => #srcs=%d\n",
			i,
			proc.RaDec.Min.Ra+proc.RaDec.DeltaRa*(float64(i%proc.RaDec.NbRa)+0.5),
			proc.RaDec.Min.Dec+proc.RaDec.DeltaDec*(float64(i%proc.RaDec.NbDec)+0.5),
			len(measures),
		)
		// loop over sources of each cell
		for _, m := range measures {
			m.ComputeMean()
			mean := m.Fluxes[0].SumMean
			if mean < proc.Flux[0] || mean > proc.Flux[1] {
				continue
			}
			fmt.Fprintf(fout, "%#v\n", m)
			nsrc += 1
		}
	}
	proc.Infof(" #src written: %d/%d\n", nsrc, proc.NbObjects)

	return err
}
