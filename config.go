package mgostatsd

import (
	"flag"
	"fmt"
	"time"

	"github.com/vharitonsky/iniflags"
)

type strings []string

/* Mongo portion of configuration */
type Mongo struct {
	Addresses []string
	User      string
	Pass      string
	AuthDb    string
}

/* Statsd portion of configuration */
type Statsd struct {
	Host    string
	Port    int
	Env     string
	Cluster string
}

/* Config contains full configuration for utility */
type Config struct {
	Verbose  bool
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

var mongoAddresses strings

/* LoadConfig loads the configuration from command-line options */
func LoadConfig() Config {
	var (
		verbose       = flag.Bool("verbose", false, "Verbose logging")
		mongoUser     = flag.String("mongo_user", "", "MongoDB User")
		mongoPass     = flag.String("mongo_pass", "", "MongoDB Password")
		mongoAuthDb   = flag.String("mongo_auth_db", "admin", "MongoDB Authentication DB")
		statsdHost    = flag.String("statsd_host", "localhost", "StatsD Host")
		statsdPort    = flag.Int("statsd_port", 8125, "StatsD Port")
		statsdEnv     = flag.String("statsd_env", "dev", "StatsD metric environment prefix")
		statsdCluster = flag.String("statsd_cluster", "unknown", "StatsD metric cluster prefix")
		interval      = flag.Duration("interval", 5*time.Second, "Polling interval")
	)

	flag.Var(&mongoAddresses, "mongo_address", "List of mongo addresses in host:port format")
	iniflags.Parse()
	if len(mongoAddresses) == 0 {
		mongoAddresses = append(mongoAddresses, "localhost:27017")
	}
	cfg := Config{
		Verbose:  *verbose,
		Interval: *interval,
		Mongo: Mongo{
			Addresses: mongoAddresses,
			User:      *mongoUser,
			Pass:      *mongoPass,
			AuthDb:    *mongoAuthDb,
		},
		Statsd: Statsd{
			Host:    *statsdHost,
			Port:    *statsdPort,
			Env:     *statsdEnv,
			Cluster: *statsdCluster,
		},
	}

	return cfg
}
