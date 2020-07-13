package config

import (
	"gitlab.com/ascenty/paloalto"
)

func LoadPaloalto() (err error) {
	for k, v := range EnvPaloalto {
		var p paloalto.Paloalto

		if p, err = paloalto.NewClient(paloalto.PaloaltoConfig{
			URL:	  v.URL,
			Username: v.Username,
			Password: v.Password,
			Vsys:	  v.Vsys,
		}); err != nil {
			return
		}

		EnvSingletons.Paloalto[k] = p
	}

	return
}
