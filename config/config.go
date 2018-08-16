package config

import (
  "os"
  _log "log"
  "strings"
  "gitlab-devops.totvs.com.br/golang/johdin"
  "gitlab-devops.totvs.com.br/golang/go-environment"
  "gitlab-devops.totvs.com.br/microservices/core/log"
  logrus "github.com/Sirupsen/logrus"
  configMoiraiHttpClient "gitlab-devops.totvs.com.br/golang/moirai-http-client/config"
)

const (
  CORE_ETCD_ENDPOINT	= "/core/env-"
  MOIRAI_HTTP_ENDPOINT	= "/moirai-http-client/env-"
  REDIS_ENDPOINT	= "/redis/env-"
  AMQP_ENDPOINT		= "/amqp/env-"
  ETCD_TIMEOUT		= 5
)

var (
  EnvConfig	      Config
  EnvRedis	      Redis
  EnvSingletons	      Singletons
  EnvAmqp	      Amqp
  EnvDB		      DB
  EnvMoiraiHttpClient configMoiraiHttpClient.Config
  EnvAmqpResources    []AmqpResourceValues
  parsed	      = false
)

type EtcdEnv struct {
  URL	    string
  Username  string
  Password  string
}

type Config struct {
  SyslogLevel	  string  `json:",omitempty" envDefault:"DEBUG"`
  SyslogNetwork	  string  `json:",omitempty" envDefault:"udp"`
  SyslogRaddr	  string  `json:",omitempty" envDefault:"localhost:514"`
  SyslogTag	  string  `json:",omitempty" envDefault:"moirai"`
  SyslogFacility  string  `json:",omitempty" envDefault:"local6"`

  OpenstackURL	    string  `json:",omitempty" envDefault:"https://services.nuvem-intera.local"`
  OpenstackUsername string  `json:",omitempty" envDefault:"admin"`
  OpenstackPassword string  `json:",omitempty" envDefault:"7Jm&7iiq4zyW4TFu"`

  NuageURL	    string  `json:",omitempty" envDefault:"https://cluster-vsd.nuvem-intera.local"`
  NuageUsername	    string  `json:",omitempty" envDefault:"apiuser"`
  NuagePassword	    string  `json:",omitempty" envDefault:"apiuser@123"`
  NuageOrganization string  `json:",omitempty" envDefault:"CSP"`

  PaloaltoURL	    string  `json:",omitempty" envDefault:"https://palo-alto-api.nuvem-intera.local"`
  PaloaltoUsername  string  `json:",omitempty" envDefault:"apiuser"`
  PaloaltoPassword  string  `json:",omitempty" envDefault:"apiuser@123"`
  PaloaltoVsys	    string  `json:",omitempty" envDefault:"vsys1"`

  BigipURL	string	`json:",omitempty" envDefault:"https:///bigip-api.nuvem-intera.local"`
  BigipUsername	string	`json:",omitempty" envDefault:"apiuser"`
  BigipPassword	string	`json:",omitempty" envDefault:"apiuser@123"`

  WapAuthURL	string	`json:",omitempty" envDefault:"http://nemesis-auth-wap.dev.nuvem-intera.local"`
  WapAdminURL	string	`json:",omitempty" envDefault:"https://adminapiwap.dbaas.dev.intera.local"`
  WapTenantURL	string	`json:",omitempty" envDefault:"https://tenapiwap.dbaas.dev.intera.local"`
  WapUsername	string	`json:",omitempty" envDefault:"wapapiuser"`
  WapPassword	string	`json:",omitempty" envDefault:"hJblx*%w?%BQ=mFca/.o"`
  WapPlanID	string	`json:",omitempty" envDefault:"PLANIiskudxsc"`
  WapSmaURL	string	`json:",omitempty" envDefault:"https://smawap.dbaas.dev.intera.local"`
  WapWsURL	string	`json:",omitempty" envDefault:"https://adminauthwap.dbaas.dev.intera.local"`

  NuageRulesProtocol  map[string]string	`json:",omitempty" envDefault:""`
}

type Singletons struct {
  Logger	  *logrus.Logger
  AmqpConnection  *johdin.Amqp
  RedisConnection  interface{}
}

type Amqp struct {
  Hosts		      []string	`json:",omitempty" envDefault:"amqp://"`
  Retry		      int	`json:",omitempty" envDefault:"10"`
  DelayErrorMessage   string	`json:",omitempty" envDefault:"1000"`
  DelayRequeueMessage string	`json:",omitempty" envDefault:"1000"`
  ExchangeType	      string	`json:",omitempty" envDefault:"x-delayed-message"`
  ExchangeRouting     string	`json:",omitempty" envDefault:"topic"`
  DeliveryMode	      uint8	`json:",omitempty" envDefault:"2"`
  Timeout	      int	`json:",omitempty" envDefault:"2000"`
  SSL		      bool	`json:",omitempty" envDefault:"false"`
  SSL_Cacert	      string	`json:",omitempty" envDefault:"/etc/pki/rabbitmq/ssl/cacert.pem"`
  SSL_Cert	      string	`json:",omitempty" envDefault:"/etc/pki/rabbitmq/ssl/nuvem-intera.local.pem"`
  SSL_Key	      string	`json:",omitempty" envDefault:"/etc/pki/rabbitmq/ssl/nuvem-intera.local.key"`
}

type Redis struct {
  Hosts	    map[string]string `json:",omitempty" envDefault:"localhost"`
  DB	    int		      `json:",omitempty" envDefault:"0"`
  Timeout   int		      `json:",omitempty" envDefault:"0"`
  Retries   int		      `json:",omitempty" envDefault:"0"`
  Heartbeat int		      `json:",omitempty" envDefault:"0"`
  Expire    int32	      `json:",omitempty" envDefault:"3000"`
}

type DB struct {
  Host		    string  `json:",omitempty" envDefault:""`
  Username	    string  `json:",omitempty" envDefault:""`
  Password	    string  `json:",omitempty" envDefault:""`
  Port		    string  `json:",omitempty" envDefault:""`
  DBName	    string  `json:",omitempty" envDefault:""`
  Timeout	    string  `json:",omitempty" envDefault:""`
  Debug		    bool    `json:",omitempty" envDefault:""`
  ConnsMaxIdle	    int	    `json:",omitempty" envDefault:""`
  ConnsMaxOpen	    int	    `json:",omitempty" envDefault:""`
  ConnsMaxLifetime  int	    `json:",omitempty" envDefault:""`
}

type AmqpResourceValues struct {
  Exchange        string  `json:",omitempty"`
  BindingKey      string  `json:",omitempty"`
  QueueName       string  `json:",omitempty"`
  OkExchange      string  `json:",omitempty"`
  OkRoutingKey    string  `json:",omitempty"`
  ErrorExchange   string  `json:",omitempty"`
  ErrorRoutingKey string  `json:",omitempty"`
  Expiration	  int32	  `json:",omitempty"`
  Lock		  bool	  `json:",omitempty"`
  Unlock	  bool	  `json:",omitempty"`
}

type Infos struct {
  Microservice	string
  DB		bool
  DBKey		string
}

func LoadConfig(infos Infos) {
  loadEtcd(infos)
  LoadLogger()
  LoadRedis()
  LoadAmqp()

  EnvSingletons.Logger.Infof(log.TEMPLATE_LOAD, PACKAGE, "LoadConfig", DONE)
}

func loadEtcd(infos Infos) {
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

    EnvConfig.NuageRulesProtocol = make(map[string]string)

    if err = conf.Get(CORE_ETCD_ENDPOINT + getEnvironment("ENV", "prod"), &EnvConfig, false, false); err != nil {
      _log.Fatalf("Error to get conf config in etcd: %s\n", err)
    }

    if err = conf.Get(REDIS_ENDPOINT + getEnvironment("ENV", "prod"), &EnvRedis, false, false); err != nil {
      _log.Fatalf("Error to get conf redis in etcd: %s\n", err)
    }

    if err = conf.Get(AMQP_ENDPOINT + getEnvironment("ENV", "prod"), &EnvAmqp, false, false); err != nil {
      _log.Fatalf("Error to get conf amqp in etcd: %s\n", err)
    }

    if err = conf.Get(MOIRAI_HTTP_ENDPOINT + getEnvironment("ENV", "prod"), &EnvMoiraiHttpClient, false, false); err != nil {
      _log.Fatalf("Error to get conf moirai http client in etcd: %s\n", err)
    }

    if err = conf.Get(infos.Microservice, &EnvAmqpResources, false, false); err != nil {
      _log.Fatalf("Error to get conf amqp resources in etcd: %s\n", err)
    }

    if infos.DB {
      if err = conf.Get(infos.DBKey, &EnvDB, false, false); err != nil {
	_log.Fatalf("Error to get conf db resources in etcd: %s\n", err)
      }
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
