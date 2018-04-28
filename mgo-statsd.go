package mstatsd

import (
	"fmt"
	"log"
	"regexp"
	str "strings"
	"time"

	"github.com/cactus/go-statsd-client/statsd"
	"github.com/kr/pretty"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
)

type Connections struct {
	Current      int64 `metric:"current"`
	Available    int64 `metric:"available"`
	TotalCreated int64 `metric:"totalCreated"`
}

type Mem struct {
	Resident          int64 `metric:"resident"`
	Virtual           int64 `metric:"virtual"`
	Mapped            int64 `metric:"mapped"`
	MappedWithJournal int64 `metric:"mappedWithJournal"`
}

type RWT struct {
	Readers int64 `metric:"readers"`
	Writers int64 `metric:"writers"`
	Total   int64 `metric:"total"`
}

type GlobalLock struct {
	TotalTime     int64 `metric:"totalTime"`
	LockTime      int64 `metric:"lockTime"`
	CurrentQueue  RWT   `metric:"currentQueue"`
	ActiveClients RWT   `metric:"activeClients"`
}

type Opcounters struct {
	Insert  int64 `metric:"insert"`
	Query   int64 `metric:"query"`
	Update  int64 `metric:"update"`
	Delete  int64 `metric:"delete"`
	GetMore int64 `metric:"getmore"`
	Command int64 `metric:"command"`
}

type ExtraInfo struct {
	PageFaults       int64 `metric:"page_faults"`
	HeapUsageInBytes int64 `metric:"heap_usage_bytes"`
}

type ReplicaInfo struct {
	IsMaster  bool `metric:"ismaster"`
	Secondary bool `metric:"secondary"`
}

type CommandCounter struct {
	Failed int64 `metric:"failed"`
	Total  int64 `metric:"total"`
}

type CursorMetrics struct {
	TimedOut int64            `metric:"timedOut"`
	Open     map[string]int64 `metric:"open"`
}

type ServerMetrics struct {
	Commands      map[string]CommandCounter `metric:"commands"`
	Cursor        CursorMetrics             `metric:"cursor"`
	Document      map[string]int64          `metric:"document"`
	Operation     map[string]int64          `metric:"operation"`
	QueryExecutor map[string]int64          `metric:"queryExecutor"`
}

type ConcurrentTransactionsInfo struct {
	Write map[string]int64 `metric:"write"`
	Read  map[string]int64 `metric:"read"`
}

type WiredTigerInfo struct {
	Cache                  map[string]int64           `metric:"cache"`
	Connection             map[string]int64           `metric:"connection"`
	ConcurrentTransactions ConcurrentTransactionsInfo `metric:"concurrentTransactions"`
}

type ServerStatus struct {
	Host                 string              `metric:"host"`
	Version              string              `metric:"version"`
	Process              string              `metric:"process"`
	Pid                  int64               `metric:"pid"`
	Uptime               int64               `metric:"uptime"`
	UptimeInMillis       int64               `metric:"uptimeMillis"`
	UptimeEstimate       int64               `metric:"uptimeEstimate"`
	LocalTime            bson.MongoTimestamp `metric:"localTime"`
	Connections          Connections         `metric:"connections"`
	ExtraInfo            ExtraInfo           `metric:"extra_info"`
	Mem                  Mem                 `metric:"mem"`
	GlobalLocks          GlobalLock          `metric:"globalLock"`
	Opcounters           Opcounters          `metric:"opcounters"`
	OpcountersReplicaSet Opcounters          `metric:"opcountersRepl"`
	ReplicaSet           ReplicaInfo         `metric:"repl"`
	Metrics              ServerMetrics       `metric:"metrics"`
	WiredTiger           *WiredTigerInfo     `metric:"wiredTiger"`
}

// GetSession creates and configures a new mgo.Session
func GetSession(mongoConfig Mongo, server string) (*mgo.Session, error) {
	dialInfo := mgo.DialInfo{
		Addrs:   []string{server},
		Direct:  true,
		Timeout: time.Second * 5,
	}
	if len(mongoConfig.User) > 0 {
		dialInfo.Username = mongoConfig.User
		dialInfo.Password = mongoConfig.Pass
		dialInfo.Source = mongoConfig.AuthDb
	}
	session, err := mgo.DialWithInfo(&dialInfo)
	if err != nil {
		return nil, err
	}

	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)

	return session, nil
}

// GetServerStatus returns a struct of the MongoDB 'serverStatus' command response
func GetServerStatus(session *mgo.Session) *ServerStatus {
	if session == nil {
		return nil
	}

	var s *ServerStatus
	if err := session.Run("serverStatus", &s); err != nil {
		log.Printf("Error running 'serverStatus' command: %v\n", err)
		return s
	}
	return s
}

func pushConnections(client statsd.Statter, connections Connections) error {
	var err error
	// Connections
	err = client.Gauge("connections.current", int64(connections.Current), 1.0)
	if err != nil {
		return err
	}

	err = client.Gauge("connections.available", int64(connections.Available), 1.0)
	if err != nil {
		return err
	}

	err = client.Gauge("connections.created", int64(connections.TotalCreated), 1.0)
	if err != nil {
		return err
	}

	return nil
}

func pushOpcounters(client statsd.Statter, opscounters Opcounters) error {
	var err error

	// Ops Counters (non-RS)
	err = client.Gauge("ops.inserts", opscounters.Insert, 1.0)
	if err != nil {
		return err
	}

	err = client.Gauge("ops.queries", opscounters.Query, 1.0)
	if err != nil {
		return err
	}

	err = client.Gauge("ops.updates", opscounters.Update, 1.0)
	if err != nil {
		return err
	}

	err = client.Gauge("ops.deletes", opscounters.Delete, 1.0)
	if err != nil {
		return err
	}

	err = client.Gauge("ops.getmores", opscounters.GetMore, 1.0)
	if err != nil {
		return err
	}

	err = client.Gauge("ops.commands", opscounters.Command, 1.0)
	if err != nil {
		return err
	}

	return nil
}

func pushMem(client statsd.Statter, mem Mem) error {
	var err error

	err = client.Gauge("mem.resident", mem.Resident, 1.0)
	if err != nil {
		return err
	}

	err = client.Gauge("mem.virtual", mem.Virtual, 1.0)
	if err != nil {
		return err
	}

	err = client.Gauge("mem.mapped", mem.Mapped, 1.0)
	if err != nil {
		return err
	}

	err = client.Gauge("mem.mapped_with_journal", mem.MappedWithJournal, 1.0)
	if err != nil {
		return err
	}

	return nil
}

func pushGlobalLocks(client statsd.Statter, glob GlobalLock) error {
	var err error

	err = client.Gauge("global_lock.total_time", glob.TotalTime, 1.0)
	if err != nil {
		return err
	}

	err = client.Gauge("global_lock.lock_time", glob.LockTime, 1.0)
	if err != nil {
		return err
	}

	err = client.Gauge("global_lock.active_readers", glob.ActiveClients.Readers, 1.0)
	if err != nil {
		return err
	}

	err = client.Gauge("global_lock.active_writers", glob.ActiveClients.Writers, 1.0)
	if err != nil {
		return err
	}

	err = client.Gauge("global_lock.active_total", glob.ActiveClients.Total, 1.0)
	if err != nil {
		return err
	}

	err = client.Gauge("global_lock.queued_readers", glob.CurrentQueue.Readers, 1.0)
	if err != nil {
		return err
	}

	err = client.Gauge("global_lock.queued_writers", glob.CurrentQueue.Writers, 1.0)
	if err != nil {
		return err
	}

	err = client.Gauge("global_lock.queued_total", glob.CurrentQueue.Total, 1.0)
	if err != nil {
		return err
	}

	return nil
}

func pushExtraInfo(client statsd.Statter, info ExtraInfo, rinfo ReplicaInfo) error {
	var err error

	err = client.Gauge("extra.page_faults", info.PageFaults, 1.0)
	if err != nil {
		return err
	}

	err = client.Gauge("extra.heap_usage", info.HeapUsageInBytes, 1.0)
	if err != nil {
		return err
	}

	if rinfo.IsMaster {
		err = client.Gauge("extra.is_master", 1, 1.0)
	} else {
		err = client.Gauge("extra.is_master", 0, 1.0)
	}
	if err != nil {
		return err
	}

	if rinfo.Secondary {
		err = client.Gauge("extra.is_secondary", 1, 1.0)
	} else {
		err = client.Gauge("extra.is_secondary", 0, 1.0)
	}
	if err != nil {
		return err
	}

	return nil
}

func pushMetrics(client statsd.Statter, serverMetrics ServerMetrics) error {
	var err error
	for k, v := range serverMetrics.Commands {
		if v.Failed > 0 || v.Total > 0 {
			err = client.Gauge(fmt.Sprintf("metrics.commands.%s.%s", k, "failed"), v.Failed, 1.0)
			if err != nil {
				return err
			}
			err = client.Gauge(fmt.Sprintf("metrics.commands.%s.%s", k, "total"), v.Total, 1.0)
			if err != nil {
				return err
			}
		}
	}

	err = client.Gauge("metrics.cursor.timedout", serverMetrics.Cursor.TimedOut, 1.0)
	if err != nil {
		return err
	}

	for k, v := range serverMetrics.Cursor.Open {
		err = client.Gauge(fmt.Sprintf("metrics.cursor.open-%s", k), v, 1.0)
		if err != nil {
			return err
		}
	}

	for k, v := range serverMetrics.Document {
		err = client.Gauge(fmt.Sprintf("metrics.document.%s", k), v, 1.0)
		if err != nil {
			return err
		}
	}
	for k, v := range serverMetrics.Operation {
		err = client.Gauge(fmt.Sprintf("metrics.operation.%s", k), v, 1.0)
		if err != nil {
			return err
		}
	}

	for k, v := range serverMetrics.QueryExecutor {
		err = client.Gauge(fmt.Sprintf("metrics.query_executor.%s", k), v, 1.0)
		if err != nil {
			return err
		}
	}

	return nil
}

var badMetricChars = regexp.MustCompile("[^-a-zA-Z_]+")

func pushWTInfo(client statsd.Statter, wtinfo *WiredTigerInfo) error {
	var err error
	if wtinfo == nil {
		return nil // WiredTiger not enabled
	}
	log.Println("WiredTiger data present!")
	for k, v := range wtinfo.Cache {
		cleanKey := badMetricChars.ReplaceAllLiteralString(k, "_") //str.Replace(k," ","_",-1)
		err = client.Gauge(fmt.Sprintf("wiredtiger.cache.%s", cleanKey), v, 1.0)
		if err != nil {
			return err
		}

	}

	for k, v := range wtinfo.ConcurrentTransactions.Read {
		cleanKey := badMetricChars.ReplaceAllLiteralString(k, "_")
		err = client.Gauge(fmt.Sprintf("wiredtiger.conc_txn_rd.%s", cleanKey), v, 1.0)
		if err != nil {
			return err
		}
	}

	for k, v := range wtinfo.ConcurrentTransactions.Write {
		cleanKey := badMetricChars.ReplaceAllLiteralString(k, "_")
		err = client.Gauge(fmt.Sprintf("wiredtiger.conc_txn_wr.%s", cleanKey), v, 1.0)
		if err != nil {
			return err
		}
	}

	for k, v := range wtinfo.Connection {
		cleanKey := badMetricChars.ReplaceAllLiteralString(k, "_")
		err = client.Gauge(fmt.Sprintf("wiredtiger.conn.%s", cleanKey), v, 1.0)
		if err != nil {
			return err
		}
	}

	return nil
}

func PushStats(statsdConfig Statsd, status *ServerStatus, verbose bool) error {
	if status == nil {
		return nil // This means we didn't connect, so lets silently skip this cycle
	}
	prefix := statsdConfig.Env
	if len(statsdConfig.Cluster) > 0 {
		prefix = fmt.Sprintf("%s.%s", prefix, statsdConfig.Cluster)
	}
	prefix = fmt.Sprintf("%s.%s", prefix, str.Replace(str.Replace(status.Host, ":", "-", -1), ".", "_", -1))
	hostPort := fmt.Sprintf("%s:%d", statsdConfig.Host, statsdConfig.Port)
	client, err := statsd.NewClient(hostPort, prefix)
	if err != nil {
		return err
	}
	defer client.Close()
	//log.Printf("Statsd Env: %v, Statsd Cluster: %v, Statsd prefix: %v\n",statsd_config.Env, statsd_config.Cluster, prefix)
	if verbose {
		log.Println(pretty.Sprintf("Mongo ServerStatus: \n%v\n", status))
	}

	err = pushConnections(client, status.Connections)
	if err != nil {
		return err
	}

	err = pushOpcounters(client, status.Opcounters)
	if err != nil {
		return err
	}

	err = pushMem(client, status.Mem)
	if err != nil {
		return err
	}

	err = pushGlobalLocks(client, status.GlobalLocks)
	if err != nil {
		return err
	}

	err = pushExtraInfo(client, status.ExtraInfo, status.ReplicaSet)
	if err != nil {
		return err
	}

	err = pushMetrics(client, status.Metrics)
	if err != nil {
		return err
	}

	err = pushWTInfo(client, status.WiredTiger)
	if err != nil {
		return err
	}

	return nil
}
