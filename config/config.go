package config

import (
  "os"
  _log "log"
  "strings"
  "encoding/json"
  "git-devops.totvs.com.br/intera/johdin"
  "git-devops.totvs.com.br/intera/go-environment"
  "git-devops.totvs.com.br/intera/core/log"
  logrus "github.com/Sirupsen/logrus"
  configMoiraiHttpClient "git-devops.totvs.com.br/intera/moirai-http-client/config"
)

const (
  CORE_ETCD_ENDPOINT	= "/core/env-"
  MOIRAI_HTTP_ENDPOINT	= "/moirai-http-client/env-"
  REDIS_ENDPOINT	= "/redis/env-"
  AMQP_ENDPOINT		= "/amqp/env-"
  BIGIP_ENDPOINT	= "/bigip"
  RUBRIK_ENDPOINT	= "/rubrik"
  ETCD_TIMEOUT		= 5
)

var (
  EnvConfig	      Config
  EnvRedis	      Redis
  EnvSingletons	      Singletons
  EnvAmqp	      Amqp
  EnvDB		      DB
  EnvBigip	      = make(map[string]BigIP)
  EnvRubrik	      Rubrik
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
  OpenstackPassword string  `json:",omitempty" envDefault:""`

  NuageURL	    string  `json:",omitempty" envDefault:"https://cluster-vsd.nuvem-intera.local"`
  NuageUsername	    string  `json:",omitempty" envDefault:""`
  NuagePassword	    string  `json:",omitempty" envDefault:""`
  NuageOrganization string  `json:",omitempty" envDefault:"CSP"`

  PaloaltoURL	    string  `json:",omitempty" envDefault:"https://palo-alto-api.nuvem-intera.local"`
  PaloaltoUsername  string  `json:",omitempty" envDefault:""`
  PaloaltoPassword  string  `json:",omitempty" envDefault:""`
  PaloaltoVsys	    string  `json:",omitempty" envDefault:"vsys1"`

  BigipURL	string	`json:",omitempty" envDefault:"https:///bigip-api.nuvem-intera.local"`
  BigipUsername	string	`json:",omitempty" envDefault:""`
  BigipPassword	string	`json:",omitempty" envDefault:""`

  WapAuthURL	  string  `json:",omitempty" envDefault:"http://nemesis-auth-wap.dev.nuvem-intera.local"`
  WapAdminURL	  string  `json:",omitempty" envDefault:"https://adminapiwap.dbaas.dev.intera.local"`
  WapTenantURL	  string  `json:",omitempty" envDefault:"https://tenapiwap.dbaas.dev.intera.local"`
  WapUsername	  string  `json:",omitempty" envDefault:""`
  WapPassword	  string  `json:",omitempty" envDefault:""`
  WapPlanID	  string  `json:",omitempty" envDefault:"PLANIiskudxsc"`
  WapSmaURL	  string  `json:",omitempty" envDefault:"https://smawap.dbaas.dev.intera.local"`
  WapWsURL	  string  `json:",omitempty" envDefault:"https://adminauthwap.dbaas.dev.intera.local"`
  WapPortDBSize	  string  `json:",omitempty" envDefault:""`
  WapPortSubUser  string  `json:",omitempty" envDefault:""`
  WapPortDatabase string  `json:",omitempty" envDefault:""`
  WapPortSMA	  string  `json:",omitempty" envDefault:""`

  CheckService	bool	`json:",omitempty" envDefault:"false"`
  CheckURL	string	`json:",omitempty" envDefault:"v1/health"`
  CheckPort	string	`json:",omitempty" envDefault:"80"`

  UnboundAddress	      string  `json:",omitempty" envDefault:""`
  UnboundCertificate	      string  `json:",omitempty" envDefault:""`
  UnboundServerNameAuthority  string  `json:",omitempty" envDefault:""`
  UnboundZoneDns	      string  `json:",omitempty" envDefault:""`

  RubrikURL	  string  `json:",omitempty" envDefault:""`
  RubrikUsername  string  `json:",omitempty" envDefault:""`
  RubrikPassword  string  `json:",omitempty" envDefault:""`

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

type BigIP struct {
  BigIpUrl        string  `json:",omitempty" envDefault:""`
  BigIpUser       string  `json:",omitempty" envDefault:""`
  BigIpPassword   string  `json:",omitempty" envDefault:""`
  RateLimit       int     `json:",omitempty" envDefault:""`
  ConnectionLimit int     `json:",omitempty" envDefault:""`
  PoolSnat        string  `json:",omitempty" envDefault:""`
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

type Rubrik struct {
  Product []RubrikProduct `json:",omitempty"`
}

type RubrikProduct struct {
  Name	string	    `json:",omitempty"`
  SO	[]RubrikSO  `json:",omitempty"`
}

type RubrikSO struct {
  Type		      string  `json:",omitempty"`
  SlaDomainName	      string  `json:",omitempty"`
  FilesetTemplateName string  `json:",omitempty"`
}

type Infos struct {
  Microservice	string
  DB		bool
  Rubrik	bool
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
    var (
      etcdEnv EtcdEnv
      err     error
      bigip   = make(map[string]interface{})
      body    []byte
    )

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

    if err = conf.Get(BIGIP_ENDPOINT, &bigip, false, false); err != nil {
      _log.Fatalf("Error to get conf bigip in etcd: %s\n", err)
    }

    if infos.Rubrik {
      if err = conf.Get(RUBRIK_ENDPOINT, &EnvRubrik, false, false); err != nil {
	_log.Fatalf("Error to get conf bigip in etcd: %s\n", err)
      }
    }

    if infos.DB {
      if err = conf.Get(infos.DBKey, &EnvDB, false, false); err != nil {
	_log.Fatalf("Error to get conf db resources in etcd: %s\n", err)
      }
    }

    for key, value := range bigip {
      if body, err = json.Marshal(value); err != nil {
	_log.Fatalf("Error to Marshal infos of bigip: %s\n", err.Error())
      }

      var ebp BigIP
      if err = json.Unmarshal(body, &ebp); err != nil {
	_log.Fatalf("Error to Unmarshal infos of bigip: %s\n", err.Error())
      }

      EnvBigip[key] = ebp
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
