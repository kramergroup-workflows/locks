package main

import (
	"flag"
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kramergroup-workflows/lock-agent/api"
	agent "github.com/kramergroup-workflows/lock-agent/internal"
)

var endpointPtr = flag.String("api-endpoint", os.Getenv("API_ENDPOINT"), "Lock API endpoint URL")
var intervalPtr = flag.Int("interval", 60, "API polling interval in seconds")

func main() {

	flag.Parse()

	// Listen for system events to gracefully terminate
	sigs := make(chan os.Signal, 1)
	signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

	// Create Argo API
	argoAPI := agent.NewArgoAPI()

	// Start an API poller and handle released locks
	lockAPI := agent.NewLockAPI(*endpointPtr)
	poller := agent.NewPoller(lockAPI, time.Duration(*intervalPtr)*time.Second)
	poller.Start(func(lock lock.Lock) {
		err := argoAPI.ResumeWorkflow(lock.Workflow, lock.Namespace)
		if err == nil {
			log.Printf("Resuming workflow %s/%s", lock.Namespace, lock.Workflow)
			lockAPI.Delete(lock.ID)
		} else {
			log.Printf("ERROR resuming workflow: %s", err)
		}
	})

	// Wait for system signals and terminate gracefully
	<-sigs
	poller.Stop()
}
