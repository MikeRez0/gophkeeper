package client

import (
	"fmt"

	"github.com/MikeRez0/gophkeeper/internal/adapter/config"
	"github.com/MikeRez0/gophkeeper/internal/adapter/logger"
	"github.com/MikeRez0/gophkeeper/internal/core/domain"
	"go.uber.org/zap"
)

func Run() error {
	conf, err := config.NewConfigClient()
	if err != nil {
		return err
	}

	log := logger.NewLogger(conf.App)

	app, err := NewApp(conf, log)
	if err != nil {
		log.Error("error creating client app", zap.Error(err))
		return err
	}

	if err = app.Connect("carl", "carl"); err != nil {
		log.Error("error connecting to server", zap.Error(err))
		return err
	}

	if err = app.FetchKeychainList(); err != nil {
		log.Error("error fetching keychain list", zap.Error(err))
		return err
	}

	kc := app.Keychains[0]
	item := kc.NewItem(domain.KCItemTypePassword)
	item.MetaDataSetItem(domain.KCMetaKeyComment, "my test comment")
	item.MetaDataSetItem(domain.KCMetaKeyLogin, "admin")
	item.MetaDataSetItem(domain.KCMetaKeySite, "google.com")

	kc.Pass = "test"
	p := "mysuper-puper-password"
	kc.StoreSecret(item, []byte(p))

	for i, k := range app.Keychains {
		fmt.Printf("Keychain %v: %v \n", i, k)

		if err = app.SyncKeychain(k); err != nil {
			log.Error("error syncing keychain", zap.Error(err))
			return err
		}

		for j, item := range k.Items {
			s, err := k.GetSecret(item)
			if err != nil {
				log.Error("error reading secret: %w", zap.Error(err))
			}

			fmt.Printf("%d: %v\n", j, string(s))
		}
	}

	return nil
}
