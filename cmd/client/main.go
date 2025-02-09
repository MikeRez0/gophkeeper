package main

import (
	"log"

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

	err := client.Run()
	if err != nil {
		log.Fatal(err)
	}

	// fmt.Printf(cBuildInfoTemplate, buildVersion, buildDate, buildCommit, runtime.GOOS, runtime.GOARCH)
	// kc := keychain.NewKeychain(nil)

	// item := kc.NewItem(domain.KCItemTypePassword)

	// // item.
	// item.MetaDataSetItem(domain.KCMetaKeyLogin, "admin")
	// item.MetaDataSetItem(domain.KCMetaKeySite, "google.com")

	// err = kc.StoreSecret(item, []byte("password"))
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// val, err := kc.GetSecret(kc.Items[0])
	// if err != nil {
	// 	log.Fatal(err)
	// }
	// fmt.Printf("%v \nSecret: %s", kc.Items[0], val)
}
