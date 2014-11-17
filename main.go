package main

import (
	"fmt"
	"github.com/cactus/go-statsd-client/statsd"
	"gopkg.in/mgo.v2"
	"gopkg.in/mgo.v2/bson"
	"time"
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
}

func serverStatus() ServerStatus {
	session, err := mgo.Dial("localhost")
	if err != nil {
		panic(err)
	}
	defer session.Close()

	// Optional. Switch the session to a monotonic behavior.
	session.SetMode(mgo.Monotonic, true)

	var s ServerStatus
	if err := session.Run("serverStatus", &s); err != nil {
		panic(err)
	}
	return s
}

func pushConnections(client statsd.Client, connections Connections) error {
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

func pushOpcounters(client statsd.Client, opscounters Opcounters) error {
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

func pushMem(client statsd.Client, mem Mem) error {
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

func pushGlobalLocks(client statsd.Client, glob GlobalLock) error {
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

func pushExtraInfo(client statsd.Client, info ExtraInfo) error {
	var err error

	err = client.Gauge("extra.page_faults", info.PageFaults, 1.0)
	if err != nil {
		return err
	}

	err = client.Gauge("extra.heap_usage", info.HeapUsageInBytes, 1.0)
	if err != nil {
		return err
	}

	return nil
}

func pushStats(status ServerStatus) error {
	prefix := fmt.Sprintf("statsd.mongodb.%s", status.Host)
	client, err := statsd.New("127.0.0.1:8125", prefix)
	if err != nil {
		return err
	}
	defer client.Close()

	err = pushConnections(*client, status.Connections)
	if err != nil {
		return err
	}

	err = pushOpcounters(*client, status.Opcounters)
	if err != nil {
		return err
	}

	err = pushMem(*client, status.Mem)
	if err != nil {
		return err
	}

	err = pushGlobalLocks(*client, status.GlobalLocks)
	if err != nil {
		return err
	}

	err = pushExtraInfo(*client, status.ExtraInfo)
	if err != nil {
		return err
	}

	return nil
}

func main() {
	ticker := time.NewTicker(5 * time.Second)
	quit := make(chan struct{})
	go func() {
		for {
			select {
			case <-ticker.C:
				err := pushStats(serverStatus())
				if err != nil {
					fmt.Println(err)
				}
			case <-quit:
				ticker.Stop()
				return
			}
		}
	}()
	time.Sleep(30 * time.Second)
	close(quit)
}
