package lsst

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/gonuts/logger"
)

// Stats holds statistics gathered during a job.
type Stats struct {
	Files        int
	MissingFiles int
	BadFiles     int
	FilesSize    int64
}

// File represents a FITS input file (from the LSST stack) to be processed/analyzed.
type File struct {
	Name   string
	Filter byte
	CamCol byte
	Field  int
	Run    int
}

// Processor is the base value holding the context to process data.
// Processor implements the P processor interface.
// A user-defined processor provides a .Proc function, which will be called
// on each FITS file being processed.
type Processor struct {
	Config func(opts Options) error
	Start  func() error
	Proc   func(f File) error
	Stop   func() error

	name string
	msg  *logger.Logger

	BaseDir   string
	OutputDir string

	RunFMMs []RunFieldMinMax
	RunFCCs []RunFieldCamCol

	RunFMMDb map[int]RunFieldMinMax

	Files []File

	RaDec RaDecLim
	Flux  [2]float64

	Stats Stats
}

// NewProcessor creates a new Processor named name.
// Users may provide  .Conf, .Start and .Stop function fields.
// They need to provide a .Proc function field.
func NewProcessor(name string) *Processor {
	const (
		ramin  = 0
		ramax  = 360
		decmin = -90
		decmax = +90
		nra    = 36
		ndec   = 18
	)

	return &Processor{
		name:     name,
		msg:      logger.New(name),
		RunFMMDb: make(map[int]RunFieldMinMax),
		RaDec: RaDecLim{
			Min: RaDec{
				Ra:  ramin,
				Dec: decmin,
			},
			Max: RaDec{
				Ra:  ramax,
				Dec: decmax,
			},
			NbRa:     nra,
			NbDec:    ndec,
			DeltaRa:  (ramax - ramin) / float64(nra),
			DeltaDec: (decmax - decmin) / float64(ndec),
		},
		Flux: [2]float64{0, 5.0e5},
	}
}

func (proc *Processor) Configure(opts Options) error {
	err := proc.configure(opts)
	if err != nil {
		return err
	}

	if proc.Config == nil {
		return nil
	}

	return proc.Config(opts)
}

func (proc *Processor) StartProcess() error {
	err := proc.start()
	if err != nil {
		return err
	}

	if proc.Start == nil {
		return nil
	}

	return proc.Start()
}

func (proc *Processor) StopProcess() error {
	err := proc.stop()
	if err != nil {
		return err
	}

	if proc.Stop == nil {
		return err
	}

	err = proc.Stop()
	if err != nil {
		return err
	}

	return err
}

func (proc *Processor) Process() error {
	if proc.Proc == nil {
		return fmt.Errorf("lsst: process [%s] has no Process function", proc.name)
	}

	var err error
	for _, f := range proc.Files {
		proc.Stats.Files += 1
		if fi, estat := os.Stat(f.Name); estat != nil {
			proc.Stats.MissingFiles += 1
			continue
		} else {
			proc.Stats.FilesSize += fi.Size()
		}
		err = proc.Proc(f)
		if err != nil {
			proc.Stats.BadFiles += 1
			return err
		}
	}
	return err
}

func (proc *Processor) Debugf(format string, args ...interface{}) (int, error) {
	return proc.msg.Debugf(format, args...)
}
func (proc *Processor) Infof(format string, args ...interface{}) (int, error) {
	return proc.msg.Infof(format, args...)
}
func (proc *Processor) Warnf(format string, args ...interface{}) (int, error) {
	return proc.msg.Warnf(format, args...)
}
func (proc *Processor) Errorf(format string, args ...interface{}) (int, error) {
	return proc.msg.Errorf(format, args...)
}

func (proc *Processor) configure(opts Options) error {
	var err error
	cfg, ok := opts.(FileOptions)
	if !ok {
		return err
	}

	proc.Infof(">>> options: %#v\n", cfg)

	proc.BaseDir = cfg.BaseDir
	proc.OutputDir = cfg.OutDir

	switch {
	case cfg.RunFMMs != nil:
		proc.Infof(">>> RunFieldMinMax: len=%d\n", len(cfg.RunFMMs))

		for _, r := range cfg.RunFMMs {
			//proc.Infof("==> %#v\n", r)

			camcol := byte(1)
			filter := byte('i')

			ccfd := fmt.Sprintf("%d/%s", camcol, string(filter))
			ccfn := fmt.Sprintf("%s%d", string(filter), camcol)

			for field := r.FieldMin; field <= r.FieldMax; field++ {
				fname := fmt.Sprintf(
					"forcedsources-%06d-%s-%04d.fits",
					r.Run,
					ccfn,
					field,
				)
				proc.Files = append(proc.Files,
					File{
						Name:   filepath.Join(cfg.BaseDir, fmt.Sprintf("%d", r.Run), ccfd, fname),
						Filter: filter,
						CamCol: camcol,
						Field:  field,
						Run:    r.Run,
					},
				)
			}
		}

	case cfg.RunFCCs != nil:
		proc.Infof(">>> RunFieldCamCol: len=%d\n", len(cfg.RunFCCs))
		for _, r := range cfg.RunFCCs {
			camcol := byte(r.CamCol)
			for _, sfilter := range cfg.Filters {
				filter := byte(sfilter[0])
				ccfd := fmt.Sprintf("%d/%s", camcol, sfilter)
				ccfn := fmt.Sprintf("%s%d", sfilter, camcol)

				fname := fmt.Sprintf(
					"forcedsources-%06d-%s-%04d.fits",
					r.Run,
					ccfn,
					r.Field,
				)
				proc.Files = append(proc.Files,
					File{
						Name:   filepath.Join(cfg.BaseDir, fmt.Sprintf("%d", r.Run), ccfd, fname),
						Filter: filter,
						CamCol: camcol,
						Field:  r.Field,
						Run:    r.Run,
					},
				)

			}
		}

	default:
		proc.Errorf("one needs either a RunFMM or RunFCC list\n")
		return fmt.Errorf("invalid configuration")
	}

	return err
}

func (proc *Processor) start() error {
	var err error

	if _, err = os.Stat(proc.OutputDir); err != nil && proc.OutputDir != "" {
		err = os.MkdirAll(proc.OutputDir, 0755)
		if err != nil {
			return err
		}
	}

	return err
}

func (proc *Processor) stop() error {
	var err error
	proc.Infof("----- stats -----\n")
	proc.Infof(" #files:     %d\n", proc.Stats.Files)
	proc.Infof(" #missing:   %d\n", proc.Stats.MissingFiles)
	proc.Infof(" #bad:       %d\n", proc.Stats.BadFiles)
	proc.Infof(" total size: %d kb\n", proc.Stats.FilesSize/1024)
	proc.Infof("-----------------\n")
	return err
}

// EOF
