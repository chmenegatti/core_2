package config

import (
	"os"
	_log "log"
	"strings"
	"context"
	"gitlab.com/ascenty/johdin"
	"gitlab.com/ascenty/core/log"
	configMoiraiHttpClient "gitlab.com/ascenty/moirai-http-client/config"
	goam "gitlab.com/ascenty/go-address-manager"
	"gitlab.com/ascenty/go-log"
	"gitlab.com/ascenty/go-etcd"
	"gitlab.com/ascenty/paloalto"
	"gitlab.com/ascenty/go-nsxt"
	"github.com/vmware/govmomi"
)

const (
	CORE_ETCD_ENDPOINT	= "/core/env-"
	MOIRAI_HTTP_ENDPOINT	= "/moirai-http-client/env-"
	REDIS_ENDPOINT		= "/redis/env-"
	AMQP_ENDPOINT		= "/amqp/env-"
	PALOALTO_ENDPOINT	= "/paloalto/env-"
	RUBRIK_ENDPOINT		= "/rubrik"
	ETCD_TIMEOUT		= 5
)

var (
	EnvConfig	    Config
	EnvRedis	    Redis
	EnvPaloalto	    map[string]Paloalto
	EnvSingletons	    Singletons
	EnvAmqp		    Amqp
	EnvDB		    DB
	EnvMoiraiHttpClient configMoiraiHttpClient.Config
	EnvAmqpResources    []AmqpResourceValues
	parsed		    = false
)

type EtcdEnv struct {
	URL	    string
	Username  string
	Password  string
}

type Config struct {
	SyslogLevel	  string  `json:",omitempty"`
	SyslogNetwork	  string  `json:",omitempty"`
	SyslogRaddr	  string  `json:",omitempty"`
	SyslogTag	  string  `json:",omitempty"`
	SyslogFacility	  string  `json:",omitempty"`

	NsxtBasePath	    string  `json:",omitempty"`
	NsxtUserName	    string  `json:",omitempty"`
	NsxtPassword	    string  `json:",omitempty"`
	NsxtRetryMaxDelay   int	    `json:",omitempty"`
	NsxtMaxRetries	    int	    `json:",omitempty"`
	NsxtRetryMinDelay   int	    `json:",omitempty"`
	NsxtInsecure	    bool    `json:",omitempty"`

	VMWareURL		  string	  `json:",omitempty"`
	VMWareUserName		  string	  `json:",omitempty"`
	VMWarePassword		  string	  `json:",omitempty"`
	VMWareInsecure		  bool		  `json:",omitempty"`
	VMWareAuthVirtualMachine  map[string]Auth `json:",omitempty"`
	VMWareDefaultDatacenter	  string	  `json:",omitempty"`

	CheckService	bool	`json:",omitempty"`
	CheckURL	string	`json:",omitempty"`
	CheckPort	string	`json:",omitempty"`

	UnboundAddress		      string  `json:",omitempty"`
	UnboundCertificate	      string  `json:",omitempty"`
	UnboundServerNameAuthority    string  `json:",omitempty"`
	UnboundZoneDns		      string  `json:",omitempty"`

	RubrikClusters	  []string  `json:",omitempty"`
	RubrikUsername	  string    `json:",omitempty"`
	RubrikPassword	  string    `json:",omitempty"`
	RubrikExpiration  int32	    `json:",omitempty"`

	AddressManagerURL string  `json:",omitempty"`

	JCStackURL	string	`json:",omitempty"`
	JCStackUsername	string	`json:",omitempty"`
	JCStackPassword	string	`json:",omitempty"`
	JCStackServer	string	`json:",omitempty"`

	DbaasURL  string  `json:",omitempty"`

	Unbound	map[string]Unbound  `json:",omitempty"`

	EnableTelescop  bool    `json:",omitempty"`
	TelescopAddress string  `json:",omitempty"`
}

type Unbound struct {
	Address		    string  `json:",omitempty"`
	Certificate	    string  `json:",omitempty"`
	ServerNameAuthority string  `json:",omitempty"`
}

type Auth struct {
	Username  string  `json:",omitempty"`
	Password  string  `json:",omitempty"`
}

type Singletons struct {
	Logger		golog.Logs
	AmqpConnection  *johdin.Amqp
	RedisConnection interface{}
	Nsxt		*nsxt.NSXTClient
	Context		context.Context
	VMWare		*govmomi.Client
	Paloalto	map[string]paloalto.Paloalto
	AddressManager	*goam.Client
}

type Amqp struct {
	Hosts		      []string	`json:",omitempty"`
	Retry		      int	`json:",omitempty"`
	DelayErrorMessage     string	`json:",omitempty"`
	DelayRequeueMessage   string	`json:",omitempty"`
	ExchangeType	      string	`json:",omitempty"`
	ExchangeRouting	      string	`json:",omitempty"`
	DeliveryMode	      uint8	`json:",omitempty"`
	Timeout		      int	`json:",omitempty"`
	SSL		      bool	`json:",omitempty"`
	SSL_Cacert	      string	`json:",omitempty"`
	SSL_Cert	      string	`json:",omitempty"`
	SSL_Key		      string	`json:",omitempty"`
}

type Paloalto struct {
	URL	  string  `json:",omitempty"`
	Username  string  `json:",omitempty"`
	Password  string  `json:",omitempty"`
	Vsys	  string  `json:",omitempty"`
}

type Redis struct {
	Hosts	    map[string]string `json:",omitempty"`
	DB	    int		      `json:",omitempty"`
	Timeout	    int		      `json:",omitempty"`
	Retries	    int		      `json:",omitempty"`
	Heartbeat   int		      `json:",omitempty"`
	Expire	    int32	      `json:",omitempty"`
}

type DB struct {
	Host		    string  `json:",omitempty"`
	Username	    string  `json:",omitempty"`
	Password	    string  `json:",omitempty"`
	Port		    string  `json:",omitempty"`
	DBName		    string  `json:",omitempty"`
	Timeout		    string  `json:",omitempty"`
	Debug		    bool    `json:",omitempty"`
	ConnsMaxIdle	    int	    `json:",omitempty"`
	ConnsMaxOpen	    int	    `json:",omitempty"`
	ConnsMaxLifetime    int	    `json:",omitempty"`
}

type AmqpResourceValues struct {
	Exchange        string  `json:",omitempty"`
	BindingKey      string  `json:",omitempty"`
	QueueName       string  `json:",omitempty"`
	OkExchange      string  `json:",omitempty"`
	OkRoutingKey    string  `json:",omitempty"`
	ErrorExchange   string  `json:",omitempty"`
	ErrorRoutingKey	string  `json:",omitempty"`
	Expiration	int32	`json:",omitempty"`
	Lock		bool	`json:",omitempty"`
	Unlock		bool	`json:",omitempty"`
	DelayMessage	string	`json:",omitempty"`
}

type Infos struct {
	Microservice	string
	DB		bool
	Nsxt		bool
	VMWare		bool
	Paloalto	bool
	AddressManager	bool
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
			c     etcd.Client
			err   error
			url   = getEnvironment("ETCD_URL", "http://127.0.0.1:2379")
		)

		if c, err = etcd.NewClient(etcd.Config{
			Endpoints:  []string{url},
			Username:   getEnvironment("ETCD_USER", "root"),
			Password:   getEnvironment("ETCD_PASSWORD", "123456"),
			Timeout:    5,
		}); err != nil {
			_log.Fatalf("Error to load etcd: %s\n", err)
		}

		if err = c.Get(CORE_ETCD_ENDPOINT + getEnvironment("ENV", "prod"), &EnvConfig); err != nil {
			_log.Fatalf("Error to get conf config in etcd: %s\n", err)
		}

		if err = c.Get(REDIS_ENDPOINT + getEnvironment("ENV", "prod"), &EnvRedis); err != nil {
			_log.Fatalf("Error to get conf redis in etcd: %s\n", err)
		}

		if err = c.Get(AMQP_ENDPOINT + getEnvironment("ENV", "prod"), &EnvAmqp); err != nil {
			_log.Fatalf("Error to get conf amqp in etcd: %s\n", err)
		}

		if err = c.Get(MOIRAI_HTTP_ENDPOINT + getEnvironment("ENV", "prod"), &EnvMoiraiHttpClient); err != nil {
			_log.Fatalf("Error to get conf moirai http client in etcd: %s\n", err)
		}

		if err = c.Get(PALOALTO_ENDPOINT + getEnvironment("ENV", "prod"), &EnvPaloalto); err != nil {
			_log.Fatalf("Error to get conf paloalto in etcd: %s\n", err)
		}

		if err = c.Get(infos.Microservice, &EnvAmqpResources); err != nil {
			_log.Fatalf("Error to get conf amqp resources in etcd: %s\n", err)
		}

		if infos.DB {
			if err = c.Get(infos.DBKey, &EnvDB); err != nil {
				_log.Fatalf("Error to get conf db resources in etcd: %s\n", err)
			}
		}

		if infos.Nsxt {
			if err = LoadNsxt(); err != nil {
				_log.Fatalf("Error to init nsxt: %s\n", err)
			}
		}

		if infos.VMWare {
			if err = LoadVMWare(); err != nil {
				_log.Fatalf("Error to init vmware: %s\n", err)
			}
		}

		if infos.Paloalto {
			EnvSingletons.Paloalto = make(map[string]paloalto.Paloalto)

			if err = LoadPaloalto(); err != nil {
				_log.Fatalf("Error to init paloalto: %s\n", err)
			}
		}

		if infos.AddressManager {
			LoadAddressManager()
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
