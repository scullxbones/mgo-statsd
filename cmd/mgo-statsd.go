package main

import (
	"log"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/kr/pretty"
	mgostatsd "github.com/scullxbones/mgo-statsd"
	"gopkg.in/mgo.v2"
)

func main() {
	config := mgostatsd.LoadConfig()

	quit := make(chan struct{})
	for i, server := range config.Mongo.Addresses {
		session, err := mgostatsd.GetSession(config.Mongo, server)
		if err != nil {
			log.Printf("Error connecting to mongo %s: %v\n", server, err)
			continue
		}
		defer session.Close()

		ticker := time.NewTicker(config.Interval)
		go func(session *mgo.Session, server string, num int) {
			for {
				select {
				case <-ticker.C:
					if config.Verbose {
						log.Printf("[%v] Starting stats for address %v \n", num, server)
					}

					status, err := mgostatsd.GetServerStatus(session)
					if err != nil {
						log.Printf("Error running 'serverStatus' command: %v\n", err)
						continue
					}
					if config.Verbose {
						log.Println(pretty.Sprintf("Mongo ServerStatus: \n%v\n", status))
					}

					err = mgostatsd.PushStats(config.Statsd, status, config.Verbose)
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
		}(session, server, i)
	}
	ch := make(chan os.Signal)
	signal.Notify(ch, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	sig := <-ch
	log.Printf("Received signal [%s]", sig.String())
	close(quit)
}
