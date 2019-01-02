package config

import (
  "time"
  "git-devops.totvs.com.br/intera/go-singleton"
  "git-devops.totvs.com.br/intera/go-cache/redis"
  "git-devops.totvs.com.br/intera/core/log"
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
