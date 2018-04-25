package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	mstatsd "github.com/scullxbones/mgo-statsd"
)

func main() {
	config := mstatsd.LoadConfig()

	quit := make(chan struct{})
	for i, server := range config.Mongo.Addresses {
		mgocnf := mstatsd.Mongo{
			Addresses: []string{server},
			User:      config.Mongo.User,
			Pass:      config.Mongo.Pass,
			AuthDb:    config.Mongo.AuthDb,
		}
		ticker := time.NewTicker(config.Interval)
		go func(cnf mstatsd.Mongo, num int) {
			for {
				select {
				case <-ticker.C:
					if config.Verbose {
						log.Printf("[%v] Starting stats for address %v \n", num, cnf.Addresses)
					}
					err := mstatsd.PushStats(config.Statsd, mstatsd.GetServerStatus(cnf), config.Verbose)
					if err != nil {
						log.Printf("[%v] ERROR: %v\n", num, err)
					}
					if config.Verbose {
						log.Printf("[%v] Done pushing stats for address %v\n", num, cnf.Addresses)
					}
				case <-quit:
					ticker.Stop()
					return
				}
			}
		}(mgocnf, i)
	}
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	sig := <-ch
	log.Printf("Received signal [%s]", sig.String())
	close(quit)
}
