// +build windows

package main

import (
	"github.com/tktip/fiks-bekymringsmelding-konsument/internal/log"
	"github.com/tktip/fiks-bekymringsmelding-konsument/internal/winservice"
)

func main() {

	// Creates a new windows service and starts it
	s, err := winservice.New()
	if err != nil {
		log.Logger.Fatal(err.Error())
	}

	//s.Install()

	//Start async service and quit. Fatal if error.
	err = s.Run()
	if err != nil {
		log.Logger.Error(err)
	}

}
