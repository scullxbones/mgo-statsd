package mgostatsd

import (
	"testing"

	mongodb_fixtures "github.com/timvaillancourt/go-mongodb-fixtures"
)

func TestServerStatusFixtures(t *testing.T) {
	status := &ServerStatus{}
	for _, version := range mongodb_fixtures.Versions() {
		t.Logf("Testing ServerStatus{} against MongoDB %s", version)
		err := mongodb_fixtures.Load(version, "serverStatus", status)
		if err != nil {
			t.Errorf("Error loading 'serverStatus' fixture for mongodb %s: %v", version, err)
			continue
		}

		// run some checks on the ServerStatus{} struct and sub-structs
		if status.Host == "" {
			t.Error("status.Host is an empty string")
		} else if status.Version == "" {
			t.Error("status.Version is an empty string")
		} else if status.Pid < 1 {
			t.Errorf("status.Pid is less than 1: %v", status.Pid)
		} else if status.Uptime < 1 {
			t.Errorf("status.Uptime is less than 1: %v", status.Uptime)
		}

		// Connections
		if status.Connections.Current < 1 {
			t.Errorf("status.Connections.Current is < 1: %v", status.Connections.Current)
		} else if status.Connections.Available < 1 {
			t.Errorf("status.Connections.Available is <= 1: %v", status.Connections.Available)
		}

		// Mem
		if status.Mem.Resident < 32 {
			t.Errorf("status.Mem.Resident is < 32: %v", status.Mem.Resident)
		} else if status.Mem.Virtual < 512 {
			t.Errorf("status.Mem.Virtual is < 512: %v", status.Mem.Virtual)
		}

		// Opcounters
		if status.Opcounters.Command < 1 {
			t.Errorf("status.Opcounters.Command is < 1: %v", status.Opcounters.Command)
		}
	}
}
