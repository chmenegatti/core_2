package config

import (
  "gitlab-devops.totvs.com.br/golang/johdin"
  "gitlab-devops.totvs.com.br/microservices/core/log"
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
