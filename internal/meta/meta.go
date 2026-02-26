package meta

import "fmt"

const (
	Name = "autoindex"
)

var (
	BuildDate  string
	CommitHash string
	Version    string
	Platform   string
	GoVersion  string
)

var (
	VersionShort string
)

func init() {
	VersionShort = fmt.Sprintf("%s %s", Name, Version)
}
