package config

import (
	"gitlab.com/ascenty/go-nsxt"
)

func LoadNsxt() (err error) {
	for key, values := range EnvConfig.Nsxt {
		var n *nsxt.NSXTClient

		if n, err = nsxt.NewNSXTClient(nsxt.Configuration{
			BasePath: values.Path,
			Username: values.Username,
			Password: values.Password,
			Insecure: values.Insecure,
		}); err != nil {
			return
		}

		EnvSingletons.Nsxt[key] = n
	}

	return
}
