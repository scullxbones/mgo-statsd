package main

import (
	"flag"
	"fmt"
	"github.com/vharitonsky/iniflags"
	"time"
)

type strings []string

type Mongo struct {
	Addresses []string
	User      string
	Pass      string
}

type Statsd struct {
	Host    string
	Port    int
	Env     string
	Cluster string
}

type Config struct {
	Interval time.Duration
	Mongo    Mongo
	Statsd   Statsd
}

func (s *strings) String() string {
	return fmt.Sprintf("%s", *s)
}

func (s *strings) Set(value string) error {
	*s = append(*s, value)
	return nil
}

var mongo_addresses strings

func LoadConfig() Config {
	var (
		mongo_user     = flag.String("mongo_user", "", "MongoDB User")
		mongo_pass     = flag.String("mongo_pass", "", "MongoDB Password")
		statsd_host    = flag.String("statsd_host", "localhost", "StatsD Host")
		statsd_port    = flag.Int("statsd_port", 8125, "StatsD Port")
		statsd_env     = flag.String("statsd_env", "dev", "StatsD metric environment prefix")
		statsd_cluster = flag.String("statsd_cluster", "0", "StatsD metric cluster prefix")
		interval       = flag.Duration("interval", 5*time.Second, "Polling interval")
	)

	flag.Var(&mongo_addresses, "mongo_address", "List of mongo addresses in host:port format")
	iniflags.Parse()
	if len(mongo_addresses) == 0 {
		mongo_addresses = append(mongo_addresses, "localhost:27017")
	}
	cfg := Config{
		Interval: *interval,
		Mongo: Mongo{
			Addresses: mongo_addresses,
			User:      *mongo_user,
			Pass:      *mongo_pass,
		},
		Statsd: Statsd{
			Host:    *statsd_host,
			Port:    *statsd_port,
			Env:     *statsd_env,
			Cluster: *statsd_cluster,
		},
	}

	return cfg
}
