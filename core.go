package core

import (
  "fmt"
  "sync"
  "gitlab-devops.totvs.com.br/golang/moirai-http-client/clients"
  configMoiraiHttpClient "gitlab-devops.totvs.com.br/golang/moirai-http-client/config"
  "gitlab-devops.totvs.com.br/golang/johdin"
  "gitlab-devops.totvs.com.br/microservices/core/config"
  "gitlab-devops.totvs.com.br/microservices/core/log"
  "encoding/json"
  "golang.org/x/net/context"
  "syscall"
  "os/signal"
  "os"
  "reflect"
  "strings"
)

type Client interface {
  Create(payload interface{}, headers map[string]interface{}) (ID uint, err error)
  Read(ID uint, headers map[string]interface{}, entity interface{}) error
  Update(ID uint, payload interface{}, headers map[string]interface{}) error
  Delete(ID uint, headers map[string]interface{}) error
}

type HttpClient struct {
  Clients map[string]Client
}

func NewHttpClient(mhc *clients.MoiraiHTTPClient) *HttpClient {
  var (
    v reflect.Value
    t reflect.Type
  )

  var hc = &HttpClient{}
  var c = map[string]Client{}

  v = reflect.Indirect(reflect.ValueOf(mhc).Elem())
  t = v.Type()

  for i := 0; i < v.NumField(); i++ {
    if f := v.Field(i); f.CanInterface() {
      if client, ok := f.Interface().(Client); ok {
	c[strings.ToLower(t.Field(i).Name)] = client
      }
    }
  }

  hc.Clients = c

  return hc
}

func NewMoiraiHttpClient(config configMoiraiHttpClient.Config) *clients.MoiraiHTTPClient {
  return clients.NewClient(config)
}

type Broker interface {
  Publish(infos johdin.AmqpInfos, pub <-chan johdin.Publishing, err chan<- error, done chan<- struct{}) error
  Consume(infos johdin.AmqpInfos, prefetch int, err chan<- error, messages chan<- johdin.Delivery) error
}

type AmqpBroker struct {
  amqp *johdin.Amqp
}

func (a *AmqpBroker) Publish(infos johdin.AmqpInfos, pub <-chan johdin.Publishing, err chan<- error, done chan<- struct{}) error {
  go func() {
    err <- a.amqp.Publish(infos, pub, done)
  }()

  return nil
}

func (a *AmqpBroker) Consume(infos johdin.AmqpInfos, prefetch int,  err chan<- error, messages chan<- johdin.Delivery) error {
  go func() {
    err <- a.amqp.Consume(infos, prefetch, messages)
  }()

  return nil
}

type Signature func(f Factorier, a Authenticate) error
type AnotherMethods map[string]Signature

type AmqpResourcePublish  map[string]map[string]chan johdin.Publishing
type AmqpResourcePublishError map[string]map[string]chan error
type AmqpResourceDone map[string]map[string]chan struct{}
type AmqpResourceDelivery map[string]map[string]chan johdin.Delivery
type AmqpResourceDeliveryError map[string]map[string]chan error

type AmqpResource struct {
  sync.Mutex
  AmqpResourcePublish	    AmqpResourcePublish
  AmqpResourcePublishError  AmqpResourcePublishError
  AmqpResourceDone	    AmqpResourceDone
  AmqpResourceDelivery	    AmqpResourceDelivery
  AmqpResourceDeliveryError AmqpResourceDeliveryError
}

type Message struct {
  ID  uint  `json:",omitempty"`
}

type Core struct {
  Factorier	Factorier
  Authenticate	Authenticate
  amqp		*AmqpResource
}

func (c *Core) Run(httpClient *HttpClient, worker Worker) {
  var (
    ctx	  context.Context
    done  context.CancelFunc
    an	  AnotherMethods
    sigs  = make(chan os.Signal, 1)
  )

  ctx, done = context.WithCancel(context.Background())
  signal.Notify(sigs, syscall.SIGINT, syscall.SIGTERM)
  c.resources()

  go func() {
    <-sigs
    done()
  }()

  an = c.mapMethods(worker)

  for _, values := range config.EnvAmqpResources {
    var (
      queue	[]string
      resource	string
      action	string
    )

    fmt.Println(values.QueueName)
    queue = strings.Split(values.QueueName, ".")
    resource = strings.ToLower(queue[2])
    action = strings.ToLower(queue[3])

    go func(resource, action string, values config.AmqpResourceValues) {
      for {
	select {
	case message := <-c.amqp.AmqpResourceDelivery[values.Exchange][values.BindingKey]:
	  var (
	    err	  error
	    msg	  Message
	    headers = make(map[string]interface{})
	  )

	  headers = message.Headers
	  if err = json.Unmarshal(message.Body, &msg); err != nil {
	    continue
	  }

	  fmt.Println(resource)
	  fmt.Println(httpClient.Clients[resource])
	  fmt.Println(msg.ID)

	  if err = httpClient.Clients[resource].Read(msg.ID, headers, worker); err != nil {
	    continue
	  }

	  switch action {
	  case "create":
	    err = worker.Create(c.Factorier, c.Authenticate)
	  case "delete":
	    err = worker.Delete(c.Factorier, c.Authenticate)
	  default:
	    err = an[action](c.Factorier, c.Authenticate)
	  }

	  if err != nil {
	    go func(exchange, routing string) {
	      c.amqp.AmqpResourcePublish[exchange][routing] <- johdin.Publishing {
		Headers:	headers,
		Body:	message.Body,
	      }
	    }(values.ErrorExchange, values.ErrorRoutingKey)

	    select {
	    case err = <-c.amqp.AmqpResourcePublishError[values.ErrorExchange][values.ErrorRoutingKey]:
	      fmt.Println("DEU MERDA AQUI")
	    case _ = <-c.amqp.AmqpResourceDone[values.ErrorExchange][values.ErrorRoutingKey]:
	    }

	    continue
	  }

	  if values.OkExchange != "" && values.OkRoutingKey != "" {
	    go func(exchange, routing string) {
	      c.amqp.AmqpResourcePublish[exchange][routing] <- johdin.Publishing {
		Headers:	headers,
		Body:	message.Body,
	      }

	      select {
	      case err = <-c.amqp.AmqpResourcePublishError[exchange][routing]:
		fmt.Println("DEU MERDA AQUI")
	      case _ = <-c.amqp.AmqpResourceDone[exchange][routing]:
	      }
	    }(values.OkExchange, values.OkRoutingKey)
	  }

	  message.Ack(false)
	case e := <-c.amqp.AmqpResourceDeliveryError[values.Exchange][values.BindingKey]:
	  fmt.Println(e)
	}
      }
    }(resource, action, values)
  }

  <-ctx.Done()
}

func (c *Core) resources() {
  var (
    broker  = &AmqpBroker{config.EnvSingletons.AmqpConnection}
    amqpRes = &AmqpResource{AmqpResourcePublish: make(AmqpResourcePublish), AmqpResourcePublishError: make(AmqpResourcePublishError), AmqpResourceDone: make(AmqpResourceDone), AmqpResourceDelivery: make(AmqpResourceDelivery), AmqpResourceDeliveryError: make(AmqpResourceDeliveryError)}
    err	    = make(chan error, 1)
  )

  go func() {
    select {
    case e := <-err:
      config.EnvSingletons.Logger.Fatalf(log.TEMPLATE_LOAD, PACKAGE, "Run", e.Error())
    }
  }()

  for _, resource := range config.EnvAmqpResources {
    var amqpInfos = johdin.AmqpInfos {
      ExchangeName: resource.Exchange,
      ExchangeType: "topic",
      Durable:	    true,
      QueueName:    resource.QueueName,
      RoutingKey:   resource.BindingKey,
    }

    if amqpRes.AmqpResourcePublish[resource.Exchange] == nil {
      amqpRes.AmqpResourcePublish[resource.Exchange] = make(map[string]chan johdin.Publishing)
      amqpRes.AmqpResourcePublishError[resource.Exchange] = make(map[string]chan error)
      amqpRes.AmqpResourceDone[resource.Exchange] = make(map[string]chan struct{})
      amqpRes.AmqpResourceDelivery[resource.Exchange] = make(map[string]chan johdin.Delivery)
      amqpRes.AmqpResourceDeliveryError[resource.Exchange] = make(map[string]chan error)
    }

    amqpRes.Lock()
    amqpRes.AmqpResourcePublish[resource.Exchange][resource.BindingKey] = make(chan johdin.Publishing)
    amqpRes.AmqpResourcePublishError[resource.Exchange][resource.BindingKey] = make(chan error)
    amqpRes.AmqpResourceDone[resource.Exchange][resource.BindingKey] = make(chan struct{})
    amqpRes.AmqpResourceDelivery[resource.Exchange][resource.BindingKey] = make(chan johdin.Delivery)
    amqpRes.AmqpResourceDeliveryError[resource.Exchange][resource.BindingKey] = make(chan error)

    go broker.Publish(amqpInfos, amqpRes.AmqpResourcePublish[resource.Exchange][resource.BindingKey], amqpRes.AmqpResourcePublishError[resource.Exchange][resource.BindingKey], amqpRes.AmqpResourceDone[resource.Exchange][resource.BindingKey])
    go broker.Consume(amqpInfos, 0, amqpRes.AmqpResourceDeliveryError[resource.Exchange][resource.BindingKey], amqpRes.AmqpResourceDelivery[resource.Exchange][resource.BindingKey])
    amqpRes.Unlock()
  }

  c.amqp = amqpRes
}


func (c *Core) mapMethods(worker Worker) AnotherMethods {
  var (
    an	    AnotherMethods
    exists  bool
    t1	    reflect.Type
    t2	    reflect.Type
  )

  an = make(AnotherMethods)
  t1 = reflect.TypeOf(worker)
  t2 = reflect.TypeOf((*Worker)(nil)).Elem()

  for i := 0; i < t1.NumMethod(); i++ {
    exists = false

    for j := 0; j < t2.NumMethod(); j++ {
      if t1.Method(i).Name == t2.Method(j).Name {
	exists = true
	break
      }
    }

    if !exists {
      an[strings.ToLower(t1.Method(i).Name)] = reflect.ValueOf(worker).Method(i).Interface().(Signature)
    }
  }

  return an
}
