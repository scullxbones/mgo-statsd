package main

import (
	"fmt"
	"github.com/cactus/go-statsd-client/statsd"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"os"
	"os/signal"
	"log"
	str "strings"
	"syscall"
	"time"
	"github.com/kr/pretty"
	"regexp"
)

type Connections struct {
	Current      int64 "current"
	Available    int64 "available"
	TotalCreated int64 "totalCreated"
}

type Mem struct {
	Resident          int64 "resident"
	Virtual           int64 "virtual"
	Mapped            int64 "mapped"
	MappedWithJournal int64 "mappedWithJournal"
}

type RWT struct {
	Readers int64 "readers"
	Writers int64 "writers"
	Total   int64 "total"
}

type GlobalLock struct {
	TotalTime     int64 "totalTime"
	LockTime      int64 "lockTime"
	CurrentQueue  RWT   "currentQueue"
	ActiveClients RWT   "activeClients"
}

type Opcounters struct {
	Insert  int64 "insert"
	Query   int64 "query"
	Update  int64 "update"
	Delete  int64 "delete"
	GetMore int64 "getmore"
	Command int64 "command"
}

type ExtraInfo struct {
	PageFaults       int64 "page_faults"
	HeapUsageInBytes int64 "heap_usage_bytes"
}

type ReplicaInfo struct {
	IsMaster	bool 	"ismaster"
	Secondary	bool	"secondary"
}

type CommandCounter struct {
	Failed  int64 "failed"
	Total   int64 "total"
}

type CursorMetrics struct {
	TimedOut 	int64 	"timedOut"
	Open 		map[string]int64 "open"
}

type ServerMetrics struct {
	Commands  		map[string]CommandCounter 	"commands"
	Cursor    		CursorMetrics				"cursor"
	Document  		map[string]int64 			"document"
	Operation 		map[string]int64 			"operation"
	QueryExecutor 	map[string]int64 			"queryExecutor"
}

type ConcurrentTransactionsInfo struct {
	Write 	map[string]int64 "write"
	Read 	map[string]int64 "read"
}

type WiredTigerInfo struct {
	Cache 					 map[string]int64           "cache"
	Connection 				 map[string]int64           "connection"
	ConcurrentTransactions   ConcurrentTransactionsInfo "concurrentTransactions"
}

type ServerStatus struct {
	Host                 string              "host"
	Version              string              "version"
	Process              string              "process"
	Pid                  int64               "pid"
	Uptime               int64               "uptime"
	UptimeInMillis       int64               "uptimeMillis"
	UptimeEstimate       int64               "uptimeEstimate"
	LocalTime            bson.MongoTimestamp "localTime"
	Connections          Connections         "connections"
	ExtraInfo            ExtraInfo           "extra_info"
	Mem                  Mem                 "mem"
	GlobalLocks          GlobalLock          "globalLock"
	Opcounters           Opcounters          "opcounters"
	OpcountersReplicaSet Opcounters          "opcountersRepl"
	ReplicaSet			 ReplicaInfo		 "repl"
	Metrics 			 ServerMetrics       "metrics"
	WiredTiger			 *WiredTigerInfo	 "wiredTiger"
}

func serverStatus(mongo_config Mongo) ServerStatus {
	info := mgo.DialInfo{
		Addrs:   mongo_config.Addresses,
		Direct:  true,
		Timeout: time.Second * 5,
	}

	if len(mongo_config.User) > 0 {
		info.Username = mongo_config.User
		info.Password = mongo_config.Pass
		info.Source = mongo_config.AuthDb
	}

	session, err := mgo.DialWithInfo(&info)
	if err != nil {
		log.Printf("Error connecting to mongo %v: %v\n", info, err)
		return ServerStatus{}
	}
	defer session.Close()

	/*if len(mongo_config.User) > 0 {
		cred := mgo.Credential{Username: mongo_config.User, Password: mongo_config.Pass, Source: mongo_config.AuthDb}
		err = session.Login(&cred)
		if err != nil {
			panic(err)
		}
	}*/

	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)

	var s ServerStatus
	if err := session.Run("serverStatus", &s); err != nil {
		log.Printf("Error connecting to %v: %v\n", info,err)
		//panic(err)
		return ServerStatus{}
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
	for k,v := range serverMetrics.Commands {
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

	for k,v := range serverMetrics.Cursor.Open {
		err = client.Gauge(fmt.Sprintf("metrics.cursor.open-%s",k), v, 1.0)
		if err != nil {
			return err
		}
	}

	for k,v := range serverMetrics.Document {
		err = client.Gauge(fmt.Sprintf("metrics.document.%s",k), v, 1.0)
		if err != nil {
			return err
		}
	}
	for k,v := range serverMetrics.Operation {
		err = client.Gauge(fmt.Sprintf("metrics.operation.%s",k), v, 1.0)
		if err != nil {
			return err
		}
	}

	for k,v := range serverMetrics.QueryExecutor {
		err = client.Gauge(fmt.Sprintf("metrics.query_executor.%s",k), v, 1.0)
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
	for k,v := range wtinfo.Cache {
		cleanKey := badMetricChars.ReplaceAllLiteralString(k, "_") //str.Replace(k," ","_",-1)
		err = client.Gauge(fmt.Sprintf("wiredtiger.cache.%s", cleanKey), v, 1.0)
		if err != nil {
			return err
		}

	}

	for k,v := range wtinfo.ConcurrentTransactions.Read {
		cleanKey := badMetricChars.ReplaceAllLiteralString(k, "_")
		err = client.Gauge(fmt.Sprintf("wiredtiger.conc_txn_rd.%s", cleanKey), v, 1.0)
		if err != nil {
			return err
		}
	}

	for k,v := range wtinfo.ConcurrentTransactions.Write {
		cleanKey := badMetricChars.ReplaceAllLiteralString(k, "_")
		err = client.Gauge(fmt.Sprintf("wiredtiger.conc_txn_wr.%s", cleanKey), v, 1.0)
		if err != nil {
			return err
		}
	}

	for k,v := range wtinfo.Connection {
		cleanKey := badMetricChars.ReplaceAllLiteralString(k, "_")
		err = client.Gauge(fmt.Sprintf("wiredtiger.conn.%s", cleanKey), v, 1.0)
		if err != nil {
			return err
		}
	}

	return nil
}


func pushStats(statsd_config Statsd, status ServerStatus) error {
	if status.Host == "" {
		return nil // This means we didn't connect, so lets silently skip this cycle
	}
	prefix := statsd_config.Env
	if len(statsd_config.Cluster) > 0 {
		prefix = fmt.Sprintf("%s.%s", prefix, statsd_config.Cluster)
	}
	prefix = fmt.Sprintf("%s.%s", prefix, str.Replace(str.Replace(status.Host,":","-",-1),".","_",-1))
	host_port := fmt.Sprintf("%s:%d", statsd_config.Host, statsd_config.Port)
	client, err := statsd.NewClient(host_port, prefix)
	if err != nil {
		return err
	}
	defer client.Close()
	//log.Printf("Statsd Env: %v, Statsd Cluster: %v, Statsd prefix: %v\n",statsd_config.Env, statsd_config.Cluster, prefix)
	log.Println(pretty.Sprintf("Mongo ServerStatus: \n%v\n", status))

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

func main() {
	config := LoadConfig()


	quit := make(chan struct{})
	for i,server := range config.Mongo.Addresses {
		mgocnf := Mongo{
			Addresses: []string{server},
			User: config.Mongo.User,
			Pass: config.Mongo.Pass,
			AuthDb: config.Mongo.AuthDb,
		}
		ticker := time.NewTicker(config.Interval)
		go func(cnf Mongo, num int) {
			for {
				select {
				case <-ticker.C:
					log.Printf("[%v] Starting stats for address %v \n", num, cnf.Addresses)
					err := pushStats(config.Statsd, serverStatus(cnf))
					if err != nil {
						fmt.Printf("[%v] ERROR: %v\n",num, err)
					}
					log.Printf("[%v] Done pushing stats for address %v\n", num, cnf.Addresses)
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
	fmt.Println("Received " + sig.String())
	close(quit)
}
