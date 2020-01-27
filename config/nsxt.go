package config

import (
	"github.com/vmware/go-vmware-nsxt"
)

func LoadNsxt() (err error) {
	EnvSingletons.Nsxt, err = nsxt.NewAPIClient(&nsxt.Configuration{
		BasePath:             EnvConfig.NsxtBasePath,
		UserName:             EnvConfig.NsxtUserName,
		Password:             EnvConfig.NsxtPassword,
		Insecure:             EnvConfig.NsxtInsecure,
		RetriesConfiguration: nsxt.ClientRetriesConfiguration{
			RetryMaxDelay:  EnvConfig.NsxtRetryMaxDelay,
			MaxRetries:     EnvConfig.NsxtMaxRetries,
			RetryMinDelay:  EnvConfig.NsxtRetryMinDelay,
		},
	})

	return
}
