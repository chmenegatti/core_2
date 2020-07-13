package config

import (
	goam "gitlab.com/ascenty/go-address-manager"
)

func LoadAddressManager() {
	EnvSingletons.AddressManager = goam.NewAddressManager(goam.Config{
		Address:  EnvConfig.AddressManagerURL,
	})
}
