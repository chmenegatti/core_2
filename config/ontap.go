package config

import (
	"gitlab.com/ascenty/go-ontap"
)

func LoadOntap() (err error) {
	for key, values := range EnvConfig.Ontap {
		var o *ontap.OntapClient

		if o, err = ontap.NewOntapClient(ontap.Configuration{
			BasePath: values.Address,
			Username: values.Username,
			Password: values.Password,
		}); err != nil {
			return
		}

		EnvSingletons.Ontap[key] = o
	}

	return
}
