package lsst

import (
	"github.com/gonuts/logger"
)

var msg = logger.New("app")

type App struct {
	Procs []P
}

// P is a processor interface
type P interface {
	// StartProcess is called before starting file scan
	StartProcess() error

	// Process processes a region of the sky
	Process() error

	// StopProcess is called at the end of the file scan
	StopProcess() error
}

type Options interface{}

type Configurer interface {
	Configure(cfg Options) error
}

func (app *App) Configure(opts Options) error {
	var err error
	msg.Infof("configure...\n")
	for _, proc := range app.Procs {
		cfg, ok := proc.(Configurer)
		if !ok {
			continue
		}
		err = cfg.Configure(opts)
		if err != nil {
			return err
		}
	}
	return err
}

func (app *App) Run() error {
	var err error
	msg.Infof("run...\n")
	msg.Infof("start...\n")
	for _, proc := range app.Procs {
		err = proc.StartProcess()
		if err != nil {
			return err
		}
	}
	msg.Infof("start... [done]\n")

	msg.Infof("process...\n")
	for _, proc := range app.Procs {
		err = proc.Process()
		if err != nil {
			return err
		}
	}
	msg.Infof("process... [done]\n")

	msg.Infof("stop...\n")
	for _, proc := range app.Procs {
		err = proc.StopProcess()
		if err != nil {
			return err
		}
	}
	msg.Infof("stop... [done]\n")

	msg.Infof("run... [done]\n")
	return err
}
