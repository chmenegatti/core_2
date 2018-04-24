package config

import (
  "os"
  _log "log"
  "strings"
  "gitlab-devops.totvs.com.br/golang/johdin"
  "gitlab-devops.totvs.com.br/golang/go-environment"
  "gitlab-devops.totvs.com.br/microservices/core/log"
  logrus "github.com/Sirupsen/logrus"
)

const (
  CORE_ETCD_ENDPOINT  = "/core/env-"
  REDIS_ENDPOINT      = "/redis/env-"
  AMQP_ENDPOINT	      = "/amqp/env-"
  ETCD_TIMEOUT	      = 5
)

var (
  EnvConfig	    Config
  EnvRedis	    Redis
  EnvSingletons	    Singletons
  EnvAmqp	    Amqp
  EnvAmqpResources  []AmqpResourceValues
  parsed	    = false
)

type EtcdEnv struct {
  URL	    string
  Username  string
  Password  string
}

type Config struct {
  SyslogLevel	  string  `env:"SYSLOG_LEVEL" envDefault:"DEBUG"`
  SyslogNetwork	  string  `env:"SYSLOG_NETWORK" envDefault:"udp"`
  SyslogRaddr	  string  `env:"SYSLOG_RADDR" envDefault:"localhost:514"`
  SyslogTag	  string  `env:"SYSLOG_TAG" envDefault:"moirai"`
  SyslogFacility  string  `env:"SYSLOG_FACILITY" envDefault:"local6"`
}

type Singletons struct {
  Logger	  *logrus.Logger
  AmqpConnection  *johdin.Amqp
  RedisConnection  interface{}
}

type Amqp struct {
  Hosts		      []string	`env:"HOSTS" envDefault:"amqp://"`
  Retry		      int	`env:"RETRY" envDefault:"10"`
  DelayErrorMessage   string	`env:"DELAY_ERROR_MESSAGE" envDefault:"1000"`
  DelayRequeueMessage string	`env:"DELAY_REQUEUE_MESSAGE" envDefault:"1000"`
  ExchangeType	      string	`env:"EXCHANGE_TYPE" envDefault:"x-delayed-message"`
  ExchangeRouting     string	`env:"EXCHANGE_ROUTING" envDefault:"topic"`
  DeliveryMode	      uint8	`env:"DELIVERY_MODE" envDefault:"2"`
  Timeout	      int	`env:"TIMEOUT" envDefault:"2000"`
  SSL		      bool	`env:"SSL" envDefault:"false"`
  SSL_Cacert	      string	`env:"SSL_CACERT" envDefault:"/etc/pki/rabbitmq/ssl/cacert.pem"`
  SSL_Cert	      string	`env:"SSL_CERT" envDefault:"/etc/pki/rabbitmq/ssl/nuvem-intera.local.pem"`
  SSL_Key	      string	`env:"SSL_KEY" envDefault:"/etc/pki/rabbitmq/ssl/nuvem-intera.local.key"`
}

type Redis struct {
  Hosts	    map[string]string `env:"HOSTS" envDefault:"localhost"`
  DB	    int		      `env:"DB" envDefault:"0"`
  Timeout   int		      `env:"TIMEOUT" envDefault:"0"`
  Retries   int		      `env:"RETRIES" envDefault:"0"`
  Heartbeat int		      `env:"HEARTBEAT" envDefault:"0"`
  Expire    int32	      `env:"EXPIRE" envDefault:"3000"`
}

type AmqpResourceValues struct {
  Exchange        string
  BindingKey      string
  QueueName       string
  OkExchange      string
  OkRoutingKey    string
  ErrorExchange   string
  ErrorRoutingKey string
}

func LoadConfig(microservice string) {
  loadEtcd(microservice)
  LoadRedis()
  LoadAmqp()
  LoadLogger()

  EnvSingletons.Logger.Infof(log.TEMPLATE_LOAD, PACKAGE, "LoadConfig", DONE)
}

func loadEtcd(microservice string) {
  if !parsed {
    var etcdEnv EtcdEnv
    var err error

    etcdEnv.URL = getEnvironment("ETCD_URL", "http://127.0.0.1:2379")
    etcdEnv.Username = getEnvironment("ETCD_USERNAME", "root")
    etcdEnv.Password = getEnvironment("ETCD_PASSWORKD", "123456")

    var conf = environment.Config{
      Username:	  etcdEnv.Username,
      Password:	  etcdEnv.Password,
      Type:	  environment.ETCD,
      TimeOut:	  ETCD_TIMEOUT,
      EndPoints:  []string{etcdEnv.URL},
    }

    if err = conf.Get(CORE_ETCD_ENDPOINT + getEnvironment("ENV", "prod"), &EnvConfig, true, false); err != nil {
      _log.Fatalf("Error to get conf config in etcd: %s\n", err)
    }

    if err = conf.Get(REDIS_ENDPOINT + getEnvironment("ENV", "prod"), &EnvRedis, true, false); err != nil {
      _log.Fatalf("Error to get conf redis in etcd: %s\n", err)
    }

    if err = conf.Get(AMQP_ENDPOINT + getEnvironment("ENV", "prod"), &EnvAmqp, true, false); err != nil {
      _log.Fatalf("Error to get conf amqp in etcd: %s\n", err)
    }

    if err = conf.Get(microservice, &EnvAmqpResources, true, false); err != nil {
      _log.Fatalf("Error to get conf amqp resources in etcd: %s\n", err)
    }

    parsed = true
  }
}

func getEnvironment(env string, def string) string {
  if e := os.Getenv(env); e != "" {
    return strings.ToLower(os.Getenv(env))
  }

  return def
}
