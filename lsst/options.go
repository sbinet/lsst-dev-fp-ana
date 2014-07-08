package lsst

type FileOptions struct {
	BaseDir string
	OutDir  string

	RaDec RaDecLim

	RunFMMs []RunFieldMinMax
	RunFCCs []RunFieldCamCol
	Filters []string

	Flux [2]float64
}
