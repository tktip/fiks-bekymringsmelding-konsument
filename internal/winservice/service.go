// +build windows

package winservice

import (
	"github.com/kardianos/service"
)

const (
	serviceName        = "Tip-Bekymringsmelding-Service"
	serviceDisplayName = "Tip Bekymringsmelding"
	serviceDescription = "Service that retrieves data from FIKSIO queue / server"
)

// New - creates a new windows service to run program
func New() (service.Service, error) {
	svcConfig := &service.Config{
		Name:        serviceName,
		DisplayName: serviceDisplayName,
		Description: serviceDescription,
	}

	return service.New(&program{}, svcConfig)
}
