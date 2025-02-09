package main

import (
	"fmt"
	"log"
	"runtime"

	"github.com/MikeRez0/gophkeeper/internal/client"
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
	fmt.Printf(cBuildInfoTemplate, buildVersion, buildDate, buildCommit, runtime.GOOS, runtime.GOARCH)

	err := client.Run()
	if err != nil {
		log.Fatal(err)
	}
}
