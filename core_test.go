package core

import (
  configMoiraiHttpClient "gitlab-devops.totvs.com.br/golang/moirai-http-client/config"
  "gitlab-devops.totvs.com.br/microservices/core/config"
  "golang.org/x/net/context"
  "net/http/httptest"
  "os/signal"
  "net/http"
  "syscall"
  "testing"
  //"errors"
  "fmt"
  "os"
)

type MockProject struct {
  Error			string	`json:",omitempty"`
  OpenstackProjectName	string	`json:",omitempty"`
}

func (mp *MockProject) Create(f Factorier, a Authenticate) StatusConsumer {
  fmt.Printf("Transaction ID: %s, Args: %s\n", f.GetTransactionID(), mp)
  return StatusConsumer{Status: COMPLETED}
  //return StatusConsumer{Status:	ERROR, Error: errors.New("Error Mock")}
}

func (mp *MockProject) Delete(f Factorier, a Authenticate) StatusConsumer {
  fmt.Printf("Transaction ID: %s, Args: %s\n", f.GetTransactionID(), mp)
  return StatusConsumer{Status: COMPLETED}
}

func (mp *MockProject) Get(f Factorier, a Authenticate) StatusConsumer {
  fmt.Printf("Transaction ID: %s, Args: %s\n", f.GetTransactionID(), mp)
  return StatusConsumer{Status: COMPLETED}
}

func (mp *MockProject) CreateError(f Factorier, a Authenticate) StatusConsumer {
  fmt.Printf("Transaction ID: %s, Args: %s\n", f.GetTransactionID(), mp)
  return StatusConsumer{Status: COMPLETED}
}

func startServerHttpClient() *httptest.Server {
  return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    switch r.RequestURI {
    case "//projects/1":
      var resp = []byte(`{"OpenstackProjectName":"MOCK"}`)

      w.WriteHeader(http.StatusOK)
      w.Write(resp)
    }
  }))
}

func loadConfCore() {
  config.EnvConfig.SyslogLevel = "DEBUG"
  config.EnvConfig.SyslogNetwork = "udp"
  config.EnvConfig.SyslogRaddr = "localhost:514"
  config.EnvConfig.SyslogTag = "core"
  config.EnvConfig.SyslogFacility = "local6"

  config.EnvRedis.Hosts = map[string]string{"localhost": "127.0.0.1:6379"}
  config.EnvRedis.Timeout = 5
  config.EnvRedis.Retries = 10
  config.EnvRedis.Heartbeat = 60
  config.EnvRedis.Expire = 3000

  config.EnvAmqp.Hosts	      = []string{"amqp://"}
  config.EnvAmqp.Retry	      = 10
  config.EnvAmqp.DeliveryMode = 2
  config.EnvAmqp.ExchangeType = "x-delayed-message"
  config.EnvAmqp.ExchangeRouting = "topic"
  config.EnvAmqp.DelayErrorMessage = "100"
  config.EnvAmqp.DelayRequeueMessage = "100"

  config.EnvAmqpResources = []config.AmqpResourceValues{
    {
      Exchange:		"broker.topic.project.create",
      BindingKey:	"v1.1.project.create",
      QueueName:	"broker.queue.project.create.v1.1.project.create",
      OkExchange:	"broker.topic.project.create",
      OkRoutingKey:	"v1.1.success",
      ErrorExchange:	"broker.topic.project.create",
      ErrorRoutingKey:	"v1.1.project.createerror",
    },
    {
      Exchange:		"broker.topic.project.create",
      BindingKey:	"v1.1.project.createerror",
      QueueName:	"broker.queue.project.create.v1.1.project.createerror",
      OkExchange:	"broker.topic.project.create",
      OkRoutingKey:	"v1.1.success",
      ErrorExchange:	"broker.topic.project.create",
      ErrorRoutingKey:	"v1.1.error",
    },
  }

  config.LoadLogger()
  config.LoadAmqp()
  config.LoadRedis()
}

func Test_Core_Run(t *testing.T) {
  var (
    shc		      = startServerHttpClient()
    sa		      = startServerAuthenticate()
    configHttpClient  = newConfigMoiraiHttpClient(shc.URL + "/")
    httpClient	      = NewHttpClient(NewMoiraiHttpClient(configHttpClient))
    factory	      = NewWorkerFactory()
    auth	      Authenticate
    c		      Core
    m		      Worker
    ctx		      context.Context
    done	      context.CancelFunc
    sigs	      = make(chan os.Signal, 1)
  )

  ctx, done = context.WithCancel(context.Background())
  signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)

  go func() {
    <-sigs
    done()
  }()

  loadConfCore()

  auth.Openstack = OpenstackAuthenticate{
    URL:      sa.URL + "/",
    Username: "MOCK",
    Password: "MOCK",
    Project:  "mock",
  }

  c = Core{factory, auth, &AmqpResource{}}
  m = &MockProject{}

  c.Run(ctx, httpClient, m)
}

func newConfigMoiraiHttpClient(url string) configMoiraiHttpClient.Config {
  return configMoiraiHttpClient.Config {
    APIVIP:           url,
    APIProjectPath:   "projects",
    LogLevel:         "DEBUG",
    SyslogNetwork:    "udp",
    SyslogRaddr:      "localhost:514",
    SyslogTag:        "moiraihttpclient",
    SyslogFacility:   "local6",
    Timeout:          100000,
  }
}
