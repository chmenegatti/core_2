package config

import (
	"gitlab.com/ascenty/go-ontap"
)

func LoadOntap() (err error) {
	EnvSingletons.Ontap, err = ontap.NewOntapClient(ontap.Configuration{
		BasePath: EnvConfig.OntapAddress,
		Username: EnvConfig.OntapUsername,
		Password: EnvConfig.OntapPassword,
	})

	return
}
