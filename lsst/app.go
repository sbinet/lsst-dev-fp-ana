package lsst

import (
	"time"

	"github.com/gonuts/logger"
)

var msg = logger.New("app")

// App is the main driver for the mini go-lsst processing framework.
//
// An App has a slice of P, each run in turn.
// The basic code-flow is as follow:
// - configure each processor
// - start each processor
// - run each processor
// - stop each processor
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

// Options is any value passed to processors to provide additional user-defined configuration data.
type Options interface{}

// Configurer models processors which can configure themselves.
type Configurer interface {
	Configure(cfg Options) error
}

// Configure configures each processor, if it implements the Configurer interface.
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

// Run runs the processors (Start/Process/Stop)
func (app *App) Run() error {
	var err error
	start := time.Now()
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

	delta := time.Since(start)
	msg.Infof("run... [done] (%v)\n", delta)

	return err
}
