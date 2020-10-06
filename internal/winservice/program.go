// +build windows

package winservice

// revive:disable:cyclomatic

import (
	"github.com/tktip/cfger"
	"github.com/tktip/fiks-bekymringsmelding-konsument/internal/fiks"
	"github.com/tktip/fiks-bekymringsmelding-konsument/internal/health"
	"os"

	"github.com/kardianos/service"
	"github.com/tktip/fiks-bekymringsmelding-konsument/internal/log"
)

type program struct {
	UseEventLog bool   `yaml:"useEventLog"`
	LogFile     string `yaml:"logFile"`

	fiksHandler fiks.Handler

	//service running this program
	service service.Service
}

// Start - starts the windows service
// should quit immediately to avoid service start timeout.
func (p *program) Start(service service.Service) error {

	p.service = service

	//Install service name for event logging purposes
	go service.Install()

	// Start should not block. Do the actual work async.
	go p.run()
	return nil
}

// addLogHooks - initializes event logger and file logger.
// If useEventLog is set to true, eventlogger is not used.
// If logFile is blank, no file logger is used.
func (p *program) addLogHooks() error {

	if p.LogFile != "" {
		err := log.AddFileHook(p.LogFile)
		if err != nil {
			return err
		}
		log.Logger.Infof("Logging to file '%s'", p.LogFile)
	} else {
		log.Logger.Info("File log not enabled.")
	}

	if p.UseEventLog {
		err := log.AddEventLogHook(p.service)
		if err != nil {
			return err
		}
		log.Logger.Info("Event log enabled, adding event logger hook")
	} else {
		log.Logger.Info("Event log not enabled.")
	}

	return nil
}

// Init - initialize the program and sub-structs.
func (p *program) Init(configFile string) (err error) {

	//read fiks config
	_, err = cfger.ReadStructuredCfgRecursive("file::"+configFile, &p.fiksHandler)
	if err != nil {
		return err
	}

	//read windows specific log config
	_, err = cfger.ReadStructuredCfgRecursive("file::"+configFile, &p)
	if err != nil {
		return err
	}

	err = p.fiksHandler.Init()
	if err != nil {
		return err
	}

	// initialize file and event loggers
	err = p.addLogHooks()

	return err
}

func (p *program) run() {
	if len(os.Args) < 2 {
		log.Logger.Fatalf("Expected config location as first parameter, found none.")
	}

	err := p.Init(os.Args[1])
	if err != nil {
		log.Logger.Fatalf("Failed during initalization: %s." + err.Error())
	}

	log.Logger.Debug("Config read.")

	// start /health at port 8090
	go health.StartHandlerIfEnabled()

	// run until failed, log on error
	log.Logger.Fatalf("Fiks handler failed: %s.", p.fiksHandler.Run())
}

// revive:disable:unused-receiver

// Stop - stops the service and ends the program
func (p *program) Stop(_ service.Service) error {
	log.Logger.Warn("Shutting down...")
	// Stop should not block. Return with a few seconds.
	return nil
}
