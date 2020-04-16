package main

import (
	"time"

	log "github.com/pavedroad-io/go-core/logger"
)

func main() {
	log.Debugf("Logging using env config: %s", "Debugf (should not appear)")
	log.Infof("Logging using env config: %s", "Infof")
	log.Warnf("Logging using env config: %s", "Warnf")
	log.Errorf("Logging using env config: %s", "Errorf")
	log.Printf("Logging using env config: %s", "Printf")
	log.Print("Logging using env config:", "Print")
	log.Println("Logging using env config:", "Println")
	time.Sleep(time.Second)
}
