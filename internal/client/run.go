package client

import (
	"fmt"

	"github.com/MikeRez0/gophkeeper/internal/adapter/config"
	"github.com/MikeRez0/gophkeeper/internal/adapter/logger"
	"go.uber.org/zap"
)

func Run() error {
	conf, err := config.NewConfigClient()
	if err != nil {
		return fmt.Errorf("error reading config: %w", err)
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
	// item := kc.NewItem(domain.KCItemTypePassword)
	// item.SetLabel("my site pass")
	// item.MetaDataSetItem(domain.KCMetaKeyComment, "my test comment")
	// item.MetaDataSetItem(domain.KCMetaKeyLogin, "admin")
	// item.MetaDataSetItem(domain.KCMetaKeySite, "google.com")

	kc.Pass = "test1"
	// p := "mysuper-puper-password"
	// err = kc.StoreSecret(item, []byte(p))
	// if err != nil {
	// 	log.Error("error storing secret", zap.Error(err))
	// 	return err
	// }

	for i, k := range app.Keychains {
		if err = app.SyncKeychain(k); err != nil {
			log.Error("error syncing keychain", zap.Error(err))
			return err
		}

		fmt.Printf("Keychain %v: %v \n", i, k)

		for j, item := range k.Items {
			s, err := k.GetSecret(item)
			if err != nil {
				log.Error("error reading secret", zap.Error(err))
			}

			fmt.Printf("%d: %v\n", j, string(s))
		}
	}

	return nil
}
