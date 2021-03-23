package core

import (
  "encoding/json"
  "database/sql"
  "reflect"
  "strings"
  "strconv"
  "errors"
  "sync"
  "fmt"
  "os"

  configMoiraiHttpClient "gitlab.com/ascenty/moirai-http-client/config"
  "gitlab.com/ascenty/moirai-http-client/clients"
  "gitlab.com/ascenty/go-cache"
  "gitlab.com/ascenty/core/config"
  "gitlab.com/ascenty/core/utils"
  "gitlab.com/ascenty/core/log"
  "gitlab.com/ascenty/johdin"
  "github.com/satori/go.uuid"
  "github.com/streadway/amqp"
  "golang.org/x/net/context"
)

const (
  DEFAULT_VALUE		    = "0"
  IN_PROGRESS		    = "IN_PROGRESS"
  ERROR			    = "ERROR"
  COMPLETED		    = "COMPLETED"
  REDELIVERE_LOCK	    = "REDELIVERE_LOCK"
  HEADER_DELAY_MESSAGE	    = "x-delay"
  HEADER_REDELIVERED_AMOUNT = "x-redelivered-amount"
  HEADER_EXCHANGE_ROUTING   = "x-delayed-type"
  DEFAULT_EXPIRATION	    = 60
)

type NewWorker func(string) Worker

type Client interface {
  Create(payload interface{}, headers map[string]interface{}) (ID uint, err error)
  Read(ID uint, headers map[string]interface{}, entity interface{}) error
  Update(ID uint, payload interface{}, headers map[string]interface{}) error
  Delete(ID uint, headers map[string]interface{}) error
}

type Base struct {
  Action  string	  `json:"-"`
  Error	  sql.NullString  `json:",omitempty"`
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
  Publish(ctx context.Context, infos johdin.Infos, pub <-chan johdin.Publishing, err chan<- error, done chan<- bool) error
  Consume(infos johdin.Infos, prefetch int, err chan<- error, messages chan<- johdin.Delivery) error
}

type AmqpBroker struct {
  amqp *johdin.Amqp
}

func (a *AmqpBroker) Publish(ctx context.Context, infos johdin.Infos, pub chan johdin.Publishing, err chan<- error, done chan<- bool) error {
  go func() {
    err <- a.amqp.Publish(ctx, infos, pub, done)
  }()

  return nil
}

func (a *AmqpBroker) Consume(infos johdin.Infos, prefetch int,  err chan<- error, messages chan<- johdin.Delivery) error {
  go func() {
    err <- a.amqp.Consume(infos, prefetch, messages)
  }()

  return nil
}

type StatusConsumer struct {
  Status  string
  Error	  error
  Update  bool
  Delete  bool
}

type Signature func(f Factorier, a Authenticate) StatusConsumer
type AnotherMethods map[string]Signature

type AmqpResourcePublish  map[string]map[string]chan johdin.Publishing
type AmqpResourcePublishError map[string]map[string]chan error
type AmqpResourceDone map[string]map[string]chan bool
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
  ID	uint	`json:",omitempty"`
  Error	string	`json:",omitempty"`
}

type SetError struct {
  Error	sql.NullString	`json:",omitempty"`
}

type Core struct {
  Factorier	Factorier
  Authenticate	Authenticate
  amqp		*AmqpResource
}

type Publish struct {
  id	      uint
  msg	      johdin.Delivery
  headers     utils.Headers
  values      config.AmqpResourceValues
  sc	      StatusConsumer
  httpClient  *HttpClient
  resource    string
}

func NewWorkerFactory() Factorier {
  return &WorkerFactory{}
}

func clone(org interface{}) interface{} {
  return reflect.New(reflect.ValueOf(org).Elem().Type()).Interface()
}

func clear(w interface{}) {
  v := reflect.ValueOf(w).Elem()
  v.Set(reflect.Zero(v.Type()))
}

func (c *Core) Run(ctx	context.Context, httpClient *HttpClient, fw NewWorker) {
  c.resources(ctx)

  go func() {
    if config.EnvConfig.CheckService == true && config.EnvConfig.CheckURL != log.EMPTY_STR {
      if err := healthCheck(); err != nil {
	config.EnvSingletons.Logger.Errorf(log.TEMPLATE_CORE, log.EMPTY_STR, PACKAGE, "Core", "healthCheck", err.Error())
	os.Exit(1)
      }
    }
  }()

  for _, values := range config.EnvAmqpResources {
    var (
      queue	[]string
      resource	string
      action	string
      rollback	string
    )

    config.EnvSingletons.Logger.Infof(log.TEMPLATE_LOAD, PACKAGE, "Load Queue", values.QueueName)

    queue = strings.Split(values.QueueName, ".")
    resource = strings.ToLower(queue[2])
    action = strings.ToLower(queue[len(queue) - 1])
    rollback = strings.ToLower(queue[len(queue) - 2])

    go func(resource, action string, values config.AmqpResourceValues) {
      for {
	select {
	case message := <-c.amqp.AmqpResourceDelivery[values.Exchange][values.BindingKey]:
	  var (
	    err	    error
	    msg	    Message
	    sc	    StatusConsumer
	    headers utils.Headers
	    w	    Worker
	    p	    Publish
	    worker  Worker
	    key	    = resource
	  )

	  worker = fw(action)
	  headers = utils.GetHeader(message.Headers)
	  p = Publish{msg: message, headers: headers, values: values, httpClient: httpClient, resource: resource}

	  if values.Lock {
	    var (
	      id      string
	      set     bool
	      err     error
	    )

	    if headers.LockTag != log.EMPTY_STR {
	      key = headers.LockTag
	    }

	    if id, set, err = lock(key, values.Expiration); err != nil {
	      config.EnvSingletons.Logger.Errorf(log.TEMPLATE_CORE, headers.TransactionID, PACKAGE, "Core", "lock", err.Error())
	      p.sc = StatusConsumer{Status: ERROR, Error: err}
	      c.publish(p)
	      continue
	    }

	    if !set {
	      p.sc = StatusConsumer{Status: REDELIVERE_LOCK}
	      c.publish(p)
	      continue
	    }

	    p.msg.Headers[utils.LOCK_ID] = id
	  }

	  if err = json.Unmarshal(message.Body, &msg); err != nil {
	    config.EnvSingletons.Logger.Errorf(log.TEMPLATE_CORE, headers.TransactionID, PACKAGE, "Core", "json.Unmarshal", err.Error())
	    p.sc = StatusConsumer{Status: ERROR, Error: err}
	    c.publish(p)
	    continue
	  }

	  p.id = msg.ID

	  if err = httpClient.Clients[resource].Read(msg.ID, message.Headers, worker); err != nil {
	    config.EnvSingletons.Logger.Errorf(log.TEMPLATE_CORE, headers.TransactionID, PACKAGE, "Core", "Read", err.Error())
	    p.sc = StatusConsumer{Status: ERROR, Error: err}
	    c.publish(p)
	    continue
	  }

	  c.Factorier.SetTransactionID(headers.TransactionID)
	  w = NewLog(worker)

	  switch action {
	  case "create":
	    sc = w.Create(c.Factorier, c.Authenticate)
	  case "delete":
	    sc = w.Delete(c.Factorier, c.Authenticate)
	  default:
	    sc = w.Custom(c.Factorier, c.Authenticate)
	  }

	  if sc.Update {
	    if err = httpClient.Clients[resource].Update(msg.ID, worker, message.Headers); err != nil {
	      config.EnvSingletons.Logger.Errorf(log.TEMPLATE_CORE, headers.TransactionID, PACKAGE, "Core", "Update", err.Error())
	      p.sc = StatusConsumer{Status: ERROR, Error: err}
	      c.publish(p)
	      continue
	    }
	  }

	  if sc.Delete {
	    if err = httpClient.Clients[resource].Delete(msg.ID, message.Headers); err != nil {
	      config.EnvSingletons.Logger.Errorf(log.TEMPLATE_CORE, headers.TransactionID, PACKAGE, "Core", "Delete", err.Error())
	      p.sc = StatusConsumer{Status: ERROR, Error: err}
	      c.publish(p)
	      continue
	    }
	  }

	  if rollback != ERROR && msg.Error != log.EMPTY_STR && (sc.Status == COMPLETED || sc.Status == IN_PROGRESS) {
	    msg.Error = log.EMPTY_STR
	  }

	  if values.Unlock {
	    if headers.LockTag != log.EMPTY_STR {
	      key = headers.LockTag
	    }

	    if err = unlock(key, headers.LockID); err != nil {
	      config.EnvSingletons.Logger.Errorf(log.TEMPLATE_CORE, headers.TransactionID, PACKAGE, "Core", "unlock", err.Error())
	      p.sc = StatusConsumer{Status: ERROR, Error: err}
	      c.publish(p)
	      continue
	    }
	  }

	  if sc.Error != nil && sc.Error.Error() == log.EMPTY_STR {
		  sc.Error = errors.New("Unknown error!")
	  }

	  p.sc = sc
	  c.publish(p)
	case e := <-c.amqp.AmqpResourceDeliveryError[values.Exchange][values.BindingKey]:
	  config.EnvSingletons.Logger.Errorf(log.TEMPLATE_PUBLISH, log.EMPTY_STR, PACKAGE, "consume", "johdin", values.Exchange, values.BindingKey, e.Error())
	  os.Exit(1)
	}
      }
    }(resource, action, values)
  }

  <-ctx.Done()
}

func lock(key string, expiration int32) (string, bool, error) {
  var (
    id	    string
    err	    error
    exists  bool
    c	    cache.Cache
  )

  id = uuid.NewV4().String()

  if expiration == 0 {
    expiration = DEFAULT_EXPIRATION
  }

  c = cache.Cache{
    Key:    key,
    Time:   expiration,
    Client: config.EnvSingletons.RedisConnection,
    Value:  id,
  }

  exists, err = c.Set()

  return id, exists, err
}

func unlock(key string, id string) error {
  var (
    err	    error
    c	    cache.Cache
    value   interface{}
    exists  bool
  )

  if id != "" {
    c = cache.Cache{
      Key:    key,
      Client: config.EnvSingletons.RedisConnection,
    }

    if value, err, exists = c.Get(); err != nil {
      return err
    }

    if exists {
      if value.(string) == id {
	return c.Del()
      }
    }
  }

  return nil
}

func (c *Core) checkMethods(action string, an AnotherMethods) bool {
  if action == "create" || action == "delete" {
    return true
  }

  if _, ok := an[action]; ok {
    return true
  }

  return false
}

func (c *Core) publish(p Publish) {
  var (
    ctx		  context.Context
    cancel	  context.CancelFunc
    exchange	  string
    routing	  string
  )

  ctx, cancel = context.WithCancel(context.Background())
  p.msg.Ack(false)

  switch p.sc.Status {
  case COMPLETED:
    if p.values.OkExchange == "" || p.values.OkRoutingKey == "" {
      return
    }

    exchange = p.values.OkExchange
    routing = p.values.OkRoutingKey

    p.msg.Headers[HEADER_REDELIVERED_AMOUNT] = DEFAULT_VALUE
    p.msg.Headers[HEADER_DELAY_MESSAGE] = DEFAULT_VALUE
  case ERROR:
    var (
      retry int
      err   error
    )

    if v, ok := p.msg.Headers[HEADER_REDELIVERED_AMOUNT]; ok {
      if retry, err = strconv.Atoi(v.(string)); err != nil {
	retry = config.EnvAmqp.Retry + 1
      }
    } else {
      retry = config.EnvAmqp.Retry + 1
    }

    if retry <= config.EnvAmqp.Retry {
      exchange = p.values.Exchange
      routing = p.values.BindingKey

      p.msg.Headers[HEADER_DELAY_MESSAGE] = config.EnvAmqp.DelayErrorMessage
      p.msg.Headers[HEADER_REDELIVERED_AMOUNT] = strconv.Itoa(retry + 1)
    } else {
      if p.values.ErrorExchange == "" || p.values.ErrorRoutingKey == "" {
	return
      }

      exchange = p.values.ErrorExchange
      routing = p.values.ErrorRoutingKey

      p.msg.Body = setErrorPublish(p.msg.Body, p.sc.Error)
      p.msg.Headers[HEADER_DELAY_MESSAGE] = config.EnvAmqp.DelayRequeueMessage
      p.msg.Headers[HEADER_REDELIVERED_AMOUNT] = DEFAULT_VALUE

      if p.values.DelayRequeueMessage != "" {
	p.msg.Headers[HEADER_DELAY_MESSAGE] = p.values.DelayRequeueMessage
      }

      if p.id != 0 {
	if err = p.httpClient.Clients[p.resource].Update(p.id, &SetError{Error: sql.NullString{String: p.sc.Error.Error(), Valid: true}}, p.msg.Headers); err != nil {
	  config.EnvSingletons.Logger.Errorf(log.TEMPLATE_CORE, p.headers.TransactionID, PACKAGE, "publish", "Update", err.Error())
	}
      }
    }
  case IN_PROGRESS:
    exchange = p.values.Exchange
    routing = p.values.BindingKey

    p.msg.Headers[HEADER_REDELIVERED_AMOUNT] = DEFAULT_VALUE

    if p.values.DelayMessage != "" {
      p.msg.Headers[HEADER_DELAY_MESSAGE] = p.values.DelayMessage
    } else {
      p.msg.Headers[HEADER_DELAY_MESSAGE] = config.EnvAmqp.DelayRequeueMessage
    }
  case REDELIVERE_LOCK:
    var delay = p.values.Expiration

    exchange = p.values.Exchange
    routing = p.values.BindingKey

    if delay == 0 {
      delay = DEFAULT_EXPIRATION
    }

    delay = (delay * 500) / 2

    p.msg.Headers[HEADER_REDELIVERED_AMOUNT] = DEFAULT_VALUE
    p.msg.Headers[HEADER_DELAY_MESSAGE] = strconv.Itoa(int(delay))
  }

  go func(exchange, routing string, transationID interface{}, msg []byte, headers map[string]interface{}) {
    var correlationID = fmt.Sprintf("%s.%s", exchange, routing)

    c.amqp.AmqpResourcePublish[exchange][routing] <- johdin.Publishing {
      Headers:	      headers,
      Body:	      msg,
      DeliveryMode:   config.EnvAmqp.DeliveryMode,
      CorrelationId:  correlationID,
    }

    select {
    case err := <-c.amqp.AmqpResourcePublishError[exchange][routing]:
      config.EnvSingletons.Logger.Errorf(log.TEMPLATE_PUBLISH, transationID, PACKAGE, "publish", "johdin", exchange, routing, err.Error())
    case _ = <-c.amqp.AmqpResourceDone[exchange][routing]:
      config.EnvSingletons.Logger.Infof(log.TEMPLATE_PUBLISH, transationID, PACKAGE, "publish", "johdin", exchange, routing, log.EMPTY_STR)
    }

    cancel()
  }(exchange, routing, p.headers.TransactionID, p.msg.Body, p.msg.Headers)

  <-ctx.Done()
}

func (c *Core) resources(ctx context.Context) {
  var (
    broker  = &AmqpBroker{config.EnvSingletons.AmqpConnection}
    amqpRes = &AmqpResource{AmqpResourcePublish: make(AmqpResourcePublish), AmqpResourcePublishError: make(AmqpResourcePublishError), AmqpResourceDone: make(AmqpResourceDone), AmqpResourceDelivery: make(AmqpResourceDelivery), AmqpResourceDeliveryError: make(AmqpResourceDeliveryError)}
    err	    = make(chan error, 1)
    args    = amqp.Table{}
  )

  if config.EnvAmqp.ExchangeRouting != "" {
    args[HEADER_EXCHANGE_ROUTING] = config.EnvAmqp.ExchangeRouting
  }

  go func() {
    select {
    case e := <-err:
      config.EnvSingletons.Logger.Fatalf(log.TEMPLATE_LOAD, PACKAGE, "Run", e.Error())
    }
  }()

  for _, resource := range config.EnvAmqpResources {
    var infos []johdin.Infos

    infos = append(infos, johdin.Infos{ExchangeName: resource.Exchange, ExchangeType: config.EnvAmqp.ExchangeType, Durable: true, QueueName: resource.QueueName, RoutingKey: resource.BindingKey, Args: args, CorrelationID: fmt.Sprintf("%s.%s", resource.Exchange, resource.BindingKey)})

    if resource.OkExchange != "" && resource.OkRoutingKey != "" {
      infos = append(infos, johdin.Infos{ExchangeName: resource.OkExchange, ExchangeType: config.EnvAmqp.ExchangeType, Durable: true, RoutingKey: resource.OkRoutingKey, Args: args, CorrelationID: fmt.Sprintf("%s.%s", resource.OkExchange, resource.OkRoutingKey)})
    }

    if resource.ErrorExchange != "" && resource.ErrorRoutingKey != "" {
      infos = append(infos, johdin.Infos{ExchangeName: resource.ErrorExchange, ExchangeType: config.EnvAmqp.ExchangeType, Durable: true, RoutingKey: resource.ErrorRoutingKey, Args: args, CorrelationID: fmt.Sprintf("%s.%s", resource.ErrorExchange, resource.ErrorRoutingKey)})
    }

    for _, values := range infos {
      if amqpRes.AmqpResourcePublish[values.ExchangeName] == nil {
	amqpRes.AmqpResourcePublish[values.ExchangeName] = make(map[string]chan johdin.Publishing)
	amqpRes.AmqpResourcePublishError[values.ExchangeName] = make(map[string]chan error)
	amqpRes.AmqpResourceDone[values.ExchangeName] = make(map[string]chan bool)

	if values.QueueName != "" {
	  amqpRes.AmqpResourceDelivery[values.ExchangeName] = make(map[string]chan johdin.Delivery)
	  amqpRes.AmqpResourceDeliveryError[values.ExchangeName] = make(map[string]chan error)
	}
      }

      if amqpRes.AmqpResourcePublish[values.ExchangeName][values.RoutingKey] == nil {
	amqpRes.Lock()
	amqpRes.AmqpResourcePublish[values.ExchangeName][values.RoutingKey] = make(chan johdin.Publishing)
	amqpRes.AmqpResourcePublishError[values.ExchangeName][values.RoutingKey] = make(chan error)
	amqpRes.AmqpResourceDone[values.ExchangeName][values.RoutingKey] = make(chan bool)
	amqpRes.Unlock()
      }

      amqpRes.Lock()
      go broker.Publish(ctx, values, amqpRes.AmqpResourcePublish[values.ExchangeName][values.RoutingKey], amqpRes.AmqpResourcePublishError[values.ExchangeName][values.RoutingKey], amqpRes.AmqpResourceDone[values.ExchangeName][values.RoutingKey])

      if values.QueueName != "" {
	amqpRes.AmqpResourceDelivery[values.ExchangeName][values.RoutingKey] = make(chan johdin.Delivery)
	amqpRes.AmqpResourceDeliveryError[values.ExchangeName][values.RoutingKey] = make(chan error)
	values.Args = amqp.Table{}

	go broker.Consume(values, 0, amqpRes.AmqpResourceDeliveryError[values.ExchangeName][values.RoutingKey], amqpRes.AmqpResourceDelivery[values.ExchangeName][values.RoutingKey])
      }
      amqpRes.Unlock()
    }
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
      an[strings.ToLower(t1.Method(i).Name)] = reflect.ValueOf(worker).Method(i).Interface().(func(Factorier, Authenticate) StatusConsumer)
    }
  }

  return an
}

func setErrorPublish(msg []byte, e error) []byte {
  var (
    m	  Message
    err	  error
    body  []byte
  )

  if err = json.Unmarshal(msg, &m); err != nil {
    return msg
  }

  m.Error = e.Error()

  if body, err = json.Marshal(m); err != nil {
    return msg
  }

  return body
}
