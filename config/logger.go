package config

import (
	"gitlab.com/ascenty/go-log"
	"log"
)

func LoadLogger() {
	var (
		logs  golog.Logs
		err	    error
	)

	if logs, err = golog.NewLog(golog.Config{
		Level:		EnvConfig.SyslogLevel,
		SyslogNetwork:	EnvConfig.SyslogNetwork,
		SyslogRaddr:	EnvConfig.SyslogRaddr,
		Priority:	EnvConfig.SyslogFacility,
		SyslogTag:	EnvConfig.SyslogTag,
	}); err != nil {
		log.Fatalf("Error to iniciate NewSyslog: ", err)
	}

	EnvSingletons.Logger = logs
}
