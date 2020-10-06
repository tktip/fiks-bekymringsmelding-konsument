package health

import (
	"net/http"
	"os"

	log "github.com/sirupsen/logrus"
)
func health(w http.ResponseWriter, _ *http.Request) {
	w.Write([]byte("OK"))
}

// StartHandlerIfEnabled starts a health-check handler if it is not disabled
func StartHandlerIfEnabled() {
	if !isHealthDisabled() {
		startHandler()
	}
}

func startHandler() {
	http.HandleFunc("/health", health)
	log.Info("Starting health check on :8090/health")
	if err := http.ListenAndServe(":8090", nil); err != nil {
		panic(err)
	}
}

// isHealthDisabled returns true if either env var DISABLE_HEALTH is "true" or if the config
// variable disableHealth is set to true
func isHealthDisabled() bool {
	if os.Getenv("DISABLE_HEALTH") == "true" {
		log.Warn("Healthcheck was disabled because the env var DISABLE_HEALTH was set to \"true\"")
		return true
	}

	return false
}
