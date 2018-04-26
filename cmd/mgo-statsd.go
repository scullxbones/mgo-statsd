package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	mstatsd "github.com/scullxbones/mgo-statsd"
	"gopkg.in/mgo.v2"
)

func main() {
	config := mstatsd.LoadConfig()

	quit := make(chan struct{})
	for i, server := range config.Mongo.Addresses {
		dialInfo := mgo.DialInfo{
			Addrs:   []string{server},
			Direct:  true,
			Timeout: time.Second * 5,
		}

		if len(config.Mongo.User) > 0 {
			dialInfo.Username = config.Mongo.User
			dialInfo.Password = config.Mongo.Pass
			dialInfo.Source = config.Mongo.AuthDb
		}

		session, err := mgo.DialWithInfo(&dialInfo)
		if err != nil {
			log.Printf("Error connecting to mongo %v: %v\n", dialInfo, err)
			return
		}
		defer session.Close()

		ticker := time.NewTicker(config.Interval)
		go func(server string, num int) {
			for {
				select {
				case <-ticker.C:
					if config.Verbose {
						log.Printf("[%v] Starting stats for address %v \n", num, server)
					}
					err := mstatsd.PushStats(config.Statsd, mstatsd.GetServerStatus(session), config.Verbose)
					if err != nil {
						log.Printf("[%v] ERROR: %v\n", num, err)
					}
					if config.Verbose {
						log.Printf("[%v] Done pushing stats for address %v\n", num, server)
					}
				case <-quit:
					ticker.Stop()
					return
				}
			}
		}(server, i)
	}
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	sig := <-ch
	log.Printf("Received signal [%s]", sig.String())
	close(quit)
}
