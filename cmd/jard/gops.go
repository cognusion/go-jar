package main

import (
	"log"

	"github.com/google/gops/agent"
)

func startAgent(stopChan chan struct{}) {

	go func() {
		if err := agent.Listen(agent.Options{}); err != nil {
			log.Fatal(err)
		}
		<-stopChan
	}()
}
