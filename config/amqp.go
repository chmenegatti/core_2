package config

import (
  "git-devops.totvs.com.br/intera/johdin"
  "git-devops.totvs.com.br/intera/core/log"
)

func LoadAmqp() {
  EnvSingletons.Logger.Infof(log.TEMPLATE_LOAD, PACKAGE, "LoadAmqp", INIT)

  var err error

  var config = johdin.Config{
    Hosts:    EnvAmqp.Hosts,
    Timeout:  EnvAmqp.Timeout,
    Logger:   EnvSingletons.Logger,
  }

  if EnvSingletons.AmqpConnection, err = johdin.Connect(config); err != nil {
    EnvSingletons.Logger.Fatalf(log.TEMPLATE_LOAD, PACKAGE, "LoadAmqp", err.Error())
  }

  EnvSingletons.Logger.Infof(log.TEMPLATE_LOAD, PACKAGE, "LoadAmqp", DONE)
}
