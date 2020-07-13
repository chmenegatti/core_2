package config

import (
	"gitlab.com/ascenty/go-nsxt"
)

func LoadNsxt() (err error) {
	EnvSingletons.Nsxt, err = nsxt.NewNSXTClient(nsxt.Configuration{
		BasePath:             EnvConfig.NsxtBasePath,
		Username:             EnvConfig.NsxtUserName,
		Password:             EnvConfig.NsxtPassword,
		Insecure:             EnvConfig.NsxtInsecure,
		/*RetriesConfiguration: nsxt.ClientRetriesConfiguration{
			RetryMaxDelay:  EnvConfig.NsxtRetryMaxDelay,
			MaxRetries:     EnvConfig.NsxtMaxRetries,
			RetryMinDelay:  EnvConfig.NsxtRetryMinDelay,
		},*/
	})

	return
}
