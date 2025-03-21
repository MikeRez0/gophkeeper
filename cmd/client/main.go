// Client application for GophKeeper.
// Stores secret items locally and syncing items with server.
package main

import (
	"fmt"
	"runtime"

	"github.com/MikeRez0/gophkeeper/cmd/client/cmd"
)

var buildVersion string
var buildDate string
var buildCommit string

const cBuildInfoTemplate = `GophKeeper client
Build version: %s
Build date: %s
Build commit: %s
OS/Arch: %s/%s
`

func main() {
	if buildVersion == "" {
		buildVersion = "N/A"
	}
	if buildDate == "" {
		buildDate = "N/A"
	}
	if buildCommit == "" {
		buildCommit = "N/A"
	}

	cmd.Execute(fmt.Sprintf(cBuildInfoTemplate, buildVersion, buildDate, buildCommit, runtime.GOOS, runtime.GOARCH))
}
