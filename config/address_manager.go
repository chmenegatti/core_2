package config

import (
	goam "git-devops.totvs.com.br/ascenty/go-address-manager"
)

func LoadAddressManager() {
	EnvSingletons.AddressManager = goam.NewAddressManager(goam.Config{
		Address:  Env.AddressManagerURL,
	})
}
