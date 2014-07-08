package lsst

import (
	"fmt"
	"math"
)

type RaDec struct {
	Ra  float64
	Dec float64
}

// RaDecLim represents a portion of the sky
type RaDecLim struct {
	Min RaDec
	Max RaDec

	NbRa  int
	NbDec int

	DeltaRa  float64
	DeltaDec float64
}

// RunFieldMinMax represents a SDSS run with a range of field numbers
type RunFieldMinMax struct {
	Run      int
	FieldMin int
	FieldMax int
}

// RunFieldCamCol represents a SDSS [run, field, camcol]
type RunFieldCamCol struct {
	Run    int
	Field  int
	CamCol int
}

var Filters = [...]byte{'u', 'g', 'r', 'i', 'z', 'y'}

// FilterID2Filter converts a filter index [1-6] to the SDSS filter byte
func FilterID2Filter(i int) byte {
	return Filters[i-1]
}

// CamColID2CamCol converts a camcol index [1-6] to the SDSS camcol byte
func CamColID2CamCol(i int) byte {
	return byte(i)
}

// FilterID returns the SDSS filter index from the according SDSS filter byte
func FilterID(b byte) int {
	for i, v := range Filters {
		if v == b {
			return i + 1
		}
	}
	panic(fmt.Errorf("invalid filter byte %q", b))
}

// CamColID returns the camcol index from the according SDSS camcol byte
func CamColID(b byte) int {
	ccid := int(b)
	if ccid < 1 || ccid > 6 {
		panic(fmt.Errorf("invalid camcol byte %q", b))
	}
	return ccid
}

// FluxRec holds individual (or sum) flux measurement(s)
type FluxRec struct {
	N          int
	SumMean    float64
	SqSumSigma float64
}

// FPMeasure holds multi-color informations extracted from forced-photometry FITS files
type FPMeasure struct {
	ID     int64
	OID    int64
	RaDec  RaDec
	Fluxes []FluxRec
}

// Add adds a new flux measure
func (m *FPMeasure) Add(idx int, flux float64) {
	v := &m.Fluxes[idx]
	v.N += 1
	v.SumMean += flux
	v.SqSumSigma += flux * flux
}

// ComputeMean computes the mean and standard deviation of this measurement
func (m *FPMeasure) ComputeMean() {
	for i, flx := range m.Fluxes {
		if flx.N < 1 {
			continue
		}
		n := float64(flx.N)
		flx.SumMean /= n
		flx.SqSumSigma = math.Sqrt(flx.SqSumSigma/n - flx.SumMean*flx.SumMean)
		m.Fluxes[i] = flx
	}
}

type FPMeasures map[int64]FPMeasure
