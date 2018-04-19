package config

import (
  "time"
  "gitlab-devops.totvs.com.br/golang/go-singleton"
  "gitlab-devops.totvs.com.br/golang/go-cache/redis"
  "gitlab-devops.totvs.com.br/microservices/core/log"
)

func LoadRedis() {
  EnvSingletons.Logger.Infof(log.TEMPLATE_LOAD, PACKAGE, "LoadRedis", INIT)

  var s = singleton.Singleton{
    ConfigCache:      redis.Config{
      RedisHosts:     EnvRedis.Hosts,
      RedisDB:	      EnvRedis.DB,
      RedisTimeout:   EnvRedis.Timeout,
      RedisRetries:   EnvRedis.Retries,
      RedisHeartbeat: time.Duration(EnvRedis.Heartbeat) * time.Second,
    },
  }

  var err error

  if err = s.Init(); err != nil {
    EnvSingletons.Logger.Fatalf(log.TEMPLATE_LOAD, PACKAGE, "LoadRedis", err.Error())
  }

  EnvSingletons.RedisConnection = s.ClientCache

  EnvSingletons.Logger.Infof(log.TEMPLATE_LOAD, PACKAGE, "LoadRedis", DONE)
}
