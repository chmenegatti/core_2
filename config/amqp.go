package config

import (
  "git-devops.totvs.com.br/intera/johdin"
  "git-devops.totvs.com.br/intera/core/log"
)

const (
  TIMEOUT_AMQP_DEFAULT	= 5
)

func LoadAmqp() {
  EnvSingletons.Logger.Infof(log.TEMPLATE_LOAD, PACKAGE, "LoadAmqp", INIT)

  var err error

  if EnvAmqp.Timeout == 0 {
    EnvAmqp.Timeout = TIMEOUT_AMQP_DEFAULT
  }

  var config = johdin.Config{
    Hosts:    EnvAmqp.Hosts,
    Timeout:  EnvAmqp.Timeout,
    Logger:   EnvSingletons.Logger,
    Certs:    johdin.Certs{
      Cacert: EnvAmqp.SSL_Cacert,
      Cert:   EnvAmqp.SSL_Cert,
      Key:    EnvAmqp.SSL_Key,
    },
  }

  if EnvSingletons.AmqpConnection, err = johdin.Connect(config); err != nil {
    EnvSingletons.Logger.Fatalf(log.TEMPLATE_LOAD, PACKAGE, "LoadAmqp", err.Error())
  }

  EnvSingletons.Logger.Infof(log.TEMPLATE_LOAD, PACKAGE, "LoadAmqp", DONE)
}
