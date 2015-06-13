package main

import (
	"github.com/lextoumbourou/idle"
	"log"
	"time"
)

func main() {
	var err error
	var idleTime time.Duration

	oneSecond, _ := time.ParseDuration("1s")

	for err == nil {
		idleTime, err = idle.Get()

		if idleTime.Seconds() > 1.0 {
			log.Printf("Idle for %d seconds.", int(idleTime.Seconds()))
		}

		time.Sleep(oneSecond)
	}

	if err != nil {
		log.Fatal(err)
	}
}
