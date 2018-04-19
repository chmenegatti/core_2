package config

import (
  moiraicommonslogging "gitlab-devops.totvs.com.br/golang/moirai-commons-logging"
  logrus "github.com/Sirupsen/logrus"
  _log "log"
)

func LoadLogger() {
  var (
    logger  *logrus.Logger
    err	    error
  )

  if logger, err = moiraicommonslogging.NewSyslog(EnvConfig.SyslogLevel, EnvConfig.SyslogNetwork, EnvConfig.SyslogRaddr, EnvConfig.SyslogFacility, EnvConfig.SyslogTag); err != nil {
    _log.Fatalf("Error to iniciate NewSyslog: ", err)
  }

  EnvSingletons.Logger = logger
}
