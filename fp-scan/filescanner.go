package main

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"sort"

	fits "github.com/astrogo/cfitsio"
	"github.com/lsst-france/fp-ana/lsst"
)

const (
	rad2deg = 180.0 / math.Pi
)

type fscanner struct {
	*lsst.Processor

	fout fits.File
	tbl  *fits.Table
}

type ForcedPhotData struct {
	Run          int32      `fits:"run"`
	Field        int32      `fits:"field"`
	CamColFilter int32      `fits:"camcol_filter"`
	NbSrc        int32      `fits:"nsrc"`
	RaMinMax     [2]float64 `fits:"ra_mnx"`
	DecMinMax    [2]float64 `fits:"dec_mnx"`
	IDMinMax     [2]int64   `fits:"id_mnx"`
	OIDMinMax    [2]int64   `fits:"oid_mnx"`
	FluxMinMax   [2]float64 `fits:"flux_mnx"`
	NbFluxOk     int32      `fits:"nfluxok"`
	FluxMean     float64    `fits:"fluxmean"`
}

func imin(i, j int64) int64 {
	if i > j {
		return j
	}
	return i
}

func imax(i, j int64) int64 {
	if i > j {
		return i
	}
	return j
}

func NewFileScanner(name string) lsst.P {

	proc := &fscanner{
		Processor: lsst.NewProcessor(name),
	}

	proc.Config = proc.config
	proc.Start = proc.start
	proc.Proc = proc.proc
	proc.Stop = proc.stop
	return proc
}

func (proc *fscanner) config(opts lsst.Options) error {
	var err error
	return err
}

func (proc *fscanner) start() error {
	var err error
	proc.Infof("output dir: [%s]\n", proc.OutputDir)
	fname := filepath.Join(proc.OutputDir, "fpfsum.fits")
	_ = os.RemoveAll(fname)

	proc.fout, err = fits.Create(fname)
	if err != nil {
		return err
	}

	phdu, err := fits.NewPrimaryHDU(&proc.fout, fits.NewDefaultHeader())
	if err != nil {
		return fmt.Errorf("error creating PHDU: %v", err)
	}

	err = phdu.Close()
	if err != nil {
		return err
	}

	proc.tbl, err = fits.NewTableFrom(&proc.fout, "fpfsum", ForcedPhotData{}, fits.BINARY_TBL)
	if err != nil {
		return err
	}
	return err
}

func (proc *fscanner) proc(f lsst.File) error {
	var err error
	proc.Infof("processing [%s] filter-id=%s camcol=%v...\n",
		f.Name, string(f.Filter), f.CamCol,
	)
	ff, err := fits.Open(f.Name, fits.ReadOnly)
	if err != nil {
		return err
	}
	defer ff.Close()

	// update run list with fields min/max values
	if rf, ok := proc.RunFMMDb[f.Run]; !ok {
		proc.RunFMMDb[f.Run] = lsst.RunFieldMinMax{
			Run:      f.Run,
			FieldMin: f.Field,
			FieldMax: f.Field,
		}
	} else {
		if rf.FieldMin > f.Field {
			rf.FieldMin = f.Field
		}
		if rf.FieldMax < f.Field {
			rf.FieldMax = f.Field
		}
		proc.RunFMMDb[f.Run] = rf
	}

	camcolfilter := int32(10*lsst.CamColID(f.CamCol) + lsst.FilterID(f.Filter))

	table := ff.HDU(1).(*fits.Table)
	defer table.Close()

	nrows := table.NumRows()
	//proc.Infof(">>> nrows=%d\n", nrows)

	fpdata := ForcedPhotData{
		Run:          int32(f.Run),
		Field:        int32(f.Field),
		CamColFilter: camcolfilter,
		NbSrc:        int32(nrows),
		RaMinMax:     [2]float64{math.MaxFloat64, -math.MaxFloat64},
		DecMinMax:    [2]float64{math.MaxFloat64, -math.MaxFloat64},
		IDMinMax:     [2]int64{math.MaxInt64, -math.MaxInt64},
		OIDMinMax:    [2]int64{math.MaxInt64, -math.MaxInt64},
		FluxMinMax:   [2]float64{math.MaxFloat64, -math.MaxFloat64},
		NbFluxOk:     int32(0),
		FluxMean:     float64(0),
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

		// convert radians to degrees
		ra := data.Coord[0] * rad2deg
		dec := data.Coord[1] * rad2deg

		fpdata.IDMinMax[0] = imin(fpdata.IDMinMax[0], data.ID)
		fpdata.IDMinMax[1] = imax(fpdata.IDMinMax[1], data.ID)

		fpdata.OIDMinMax[0] = imin(fpdata.OIDMinMax[0], data.OID)
		fpdata.OIDMinMax[1] = imax(fpdata.OIDMinMax[1], data.OID)

		fpdata.RaMinMax[0] = math.Min(fpdata.RaMinMax[0], ra)
		fpdata.RaMinMax[1] = math.Max(fpdata.RaMinMax[1], ra)

		fpdata.DecMinMax[0] = math.Min(fpdata.DecMinMax[0], dec)
		fpdata.DecMinMax[1] = math.Max(fpdata.DecMinMax[1], dec)

		fpdata.FluxMinMax[0] = math.Min(fpdata.FluxMinMax[0], data.Flux)
		fpdata.FluxMinMax[1] = math.Max(fpdata.FluxMinMax[1], data.Flux)
		if data.Flux > proc.Flux[0] && data.Flux < proc.Flux[1] {
			fpdata.NbFluxOk += 1
			fpdata.FluxMean += data.Flux
		}
	}

	if fpdata.NbFluxOk > 0 {
		fpdata.FluxMean /= float64(fpdata.NbFluxOk)
	}

	err = rows.Err()
	if err != nil {
		return err
	}

	err = proc.tbl.Write(&fpdata)
	if err != nil {
		return err
	}

	//proc.Infof("processing [%s] filter-id=%v camcol=%v... [done]\n", f.Name, f.Filter, f.CamCol)
	return err
}

func (proc *fscanner) stop() error {
	var err error
	stats, err := os.Create(filepath.Join(proc.OutputDir, "stats.txt"))
	if err != nil {
		return err
	}
	defer stats.Close()

	fmt.Fprintf(stats, "## stats: %#v\n", proc.Stats)

	err = proc.tbl.Close()
	if err != nil {
		return err
	}

	err = proc.fout.Close()
	if err != nil {
		return err
	}

	runs := make([]int, 0, len(proc.RunFMMDb))
	for run := range proc.RunFMMDb {
		runs = append(runs, int(run))
	}
	sort.Ints(runs)

	fmt.Fprintf(stats, "## run field-min field-max\n")
	proc.Infof("## run field-min field-max\n")
	for _, run := range runs {
		rfmm := proc.RunFMMDb[run]
		proc.Infof("%06d\t%04d\t  %04d\n", rfmm.Run, rfmm.FieldMin, rfmm.FieldMax)
		fmt.Fprintf(stats, "%06d %04d %04d\n", rfmm.Run, rfmm.FieldMin, rfmm.FieldMax)
	}
	return err
}
