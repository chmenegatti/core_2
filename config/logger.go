package config

import (
	moiraicommonslogging "git-devops.totvs.com.br/intera/moirai-commons-logging"
	_log "log"
)

func LoadLogger() {
	var (
		logs  moiraicommonslogging.Logs
		err	    error
	)

	if logs, err = moiraicommonslogging.NewLog(moiraicommonslogging.Config{
		Level:		EnvConfig.SyslogLevel,
		SyslogNetwork:	EnvConfig.SyslogNetwork,
		SyslogRaddr:	EnvConfig.SyslogRaddr,
		Priority:	EnvConfig.SyslogFacility,
		SyslogTag:	EnvConfig.SyslogTag,
	}); err != nil {
		_log.Fatalf("Error to iniciate NewSyslog: ", err)
	}

	EnvSingletons.Logger = logs
}
