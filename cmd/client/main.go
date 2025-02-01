package main

import (
	"fmt"
	"log"

	"github.com/MikeRez0/gophkeeper/internal/core/domain"
	"github.com/MikeRez0/gophkeeper/internal/core/keychain"
)

func main() {
	kc := keychain.NewKeyChain(nil)

	item := kc.NewItem(domain.KCItemTypePassword)

	item.Data.MetaData[domain.KCMetaKeyLogin] = "admin"
	item.Data.MetaData[domain.KCMetaKeySite] = "google.com"

	err := kc.StoreSecret(item, []byte("password"))
	if err != nil {
		log.Fatal(err)
	}

	fmt.Println(string(kc.Items[0].Data.Value))
}
