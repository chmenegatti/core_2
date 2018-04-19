package config

import (
  "golang.org/x/net/context"
  "gitlab-devops.totvs.com.br/golang/johdin"
  "gitlab-devops.totvs.com.br/microservices/core/log"
)

func LoadAmqp() {
  EnvSingletons.Logger.Infof(log.TEMPLATE_LOAD, PACKAGE, "LoadAmqp", INIT)

  var ctx context.Context
  ctx, _ = context.WithCancel(context.Background())

  if EnvSingletons.AmqpConnection = johdin.New(EnvAmqp.Hosts, ctx); EnvSingletons.AmqpConnection == nil {
    EnvSingletons.Logger.Fatalf(log.TEMPLATE_LOAD, PACKAGE, "LoadAmqp", "Could not connect in rabbitmq-server")
  }

  EnvSingletons.Logger.Infof(log.TEMPLATE_LOAD, PACKAGE, "LoadAmqp", DONE)
}
