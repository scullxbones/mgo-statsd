package fixtures

import (
	"io/ioutil"
	"os"
	"path/filepath"
	"runtime"

	version "github.com/hashicorp/go-version"
	"gopkg.in/mgo.v2/bson"
)

func VersionsDir() string {
	_, filename, _, ok := runtime.Caller(0)
	if !ok {
		return ""
	}
	return filepath.Join(filepath.Dir(filename), "versions")
}

func Load(versionStr, command string, out interface{}) error {
	filePath := filepath.Join(VersionsDir(), versionStr, command+".bson")
	bytes, err := ioutil.ReadFile(filePath)
	if err != nil {
		return err
	}
	return bson.Unmarshal(bytes, out)
}

func Write(versionStr, command string, data []byte) error {
	versionDir := filepath.Join(VersionsDir(), versionStr)
	if _, err := os.Stat(versionDir); os.IsNotExist(err) {
		err = os.Mkdir(versionDir, 0755)
		if err != nil {
			return err
		}
	}
	filePath := filepath.Join(versionDir, command+".bson")
	return ioutil.WriteFile(filePath, data, 0644)
}

func Versions() []string {
	var versions []string
	subdirs, err := ioutil.ReadDir(VersionsDir())
	if err != nil {
		return versions
	}
	for _, subdir := range subdirs {
		if subdir.IsDir() {
			versions = append(versions, subdir.Name())
		}
	}
	return versions
}

func VersionsFilter(filter string) []string {
	var versions []string
	for _, versionStr := range Versions() {
		if IsVersionMatch(versionStr, filter) {
			versions = append(versions, versionStr)
		}
	}
	return versions
}

func IsVersionMatch(versionStr, filter string) bool {
	constraints, err := version.NewConstraint(filter)
	if err != nil {
		return false
	}
	v, err := version.NewVersion(versionStr)
	if err != nil {
		return false
	}
	return constraints.Check(v)
}
