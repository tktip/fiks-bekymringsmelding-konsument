// +build linux

package main

//revive:disable:confusing-naming

import (
	"github.com/tktip/cfger"
	"github.com/tktip/fiks-bekymringsmelding-konsument/internal/fiks"
	"github.com/tktip/fiks-bekymringsmelding-konsument/internal/health"
)

func main() {

	go health.StartHandlerIfEnabled()
	// Create FIKS handler
	handler := fiks.Handler{}
	_, err := cfger.ReadStructuredCfgRecursive("env::CONFIG", &handler)
	if err != nil {
		panic(err)
	}

	err = handler.Init()
	if err != nil {
		panic(err)
	}

	// Run FIKS handler
	err = handler.Run()
	if err != nil {
		panic(err)
	}
}
