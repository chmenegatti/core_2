package core

import (
  "gitlab-devops.totvs.com.br/golang/openstack"
  "gitlab-devops.totvs.com.br/golang/nuage"
  "gitlab-devops.totvs.com.br/golang/paloalto"
  "gitlab-devops.totvs.com.br/golang/bigip"
  "gitlab-devops.totvs.com.br/golang/go-wap-client"
  "gitlab-devops.totvs.com.br/golang/go-singleton"
  "gitlab-devops.totvs.com.br/golang/go-cache/redis"
  "gitlab-devops.totvs.com.br/microservices/core/config"
  "gitlab-devops.totvs.com.br/microservices/core/log"
)

const (
  FORCE_AUTHENTICATE  = true
)

type Authenticate struct {
  Openstack OpenstackAuthenticate
  Nuage	    NuageAuthenticate
  Paloalto  PaloaltoAuthenticate
  Bigip	    BigipAuthenticate
  Wap	    WapAuthenticate
}

type OpenstackAuthenticate struct {
  URL      string
  Username string
  Password string
  Project  string
}

type NuageAuthenticate struct {
  URL		string
  Organization	string
  Username	string
  Password	string
  EnterpriseID	string
}

type PaloaltoAuthenticate struct {
  URL	    string
  Username  string
  Password  string
  Vsys	    string
}

type BigipAuthenticate struct {
  URL	    string
  Username  string
  Password  string
}

type WapAuthenticate struct {
  AuthURL   string
  AdminURL  string
  TenantURL string
  Username  string
  Password  string
  PlanID    string
  SmaURL    string
  WsURL	    string
}

type Factorier interface {
  Openstack(Authenticate) (*openstack.Openstack, error)
  Nuage(Authenticate) (*nuage.Nuage, error)
  Paloalto(Authenticate) (*paloalto.Paloalto, error)
  Bigip(Authenticate) (*bigip.Bigip, error)
  Wap(Authenticate) (*gowapclient.WAP, error)
  GetTransactionID() string
  SetTransactionID(string)
}

type Factory struct {}

func (f *Factory) Openstack(a Authenticate) (*openstack.Openstack, error) {
  panic("Method Openstack not implemented")
}

func (f *Factory) Nuage(a Authenticate) (*nuage.Nuage, error) {
  panic("Method Nuage not implemented")
}

func (f *Factory) Paloalto(a Authenticate) (*paloalto.Paloalto, error) {
  panic("Method Paloalto not implemented")
}

func (f *Factory) Bigip(a Authenticate) (*bigip.Bigip, error) {
  panic("Method Bigip not implemented")
}

func (f *Factory) Wap(a Authenticate) (*bigip.Bigip, error) {
  panic("Method Wap not implemented")
}

func (f *Factory) GetTransactionID() string {
  panic("Method GetTransactionID not implemented")
}

func (f *Factory) SetTransactionID(id string) {
  panic("Method SetTransactionID not implemented")
}

type WorkerFactory struct {
  Factory

  openstack	*openstack.Openstack
  nuage		*nuage.Nuage
  paloalto	*paloalto.Paloalto
  bigip		*bigip.Bigip
  wap		*gowapclient.WAP
  transactionID	string
}

func (wf *WorkerFactory) Openstack(a Authenticate) (*openstack.Openstack, error) {
  var (
    client  interface{}
    err	    error
    op	    openstack.Openstack
  )

  if client, err = wf.authenticate(
    openstack.SAuth{
      a.Openstack.URL,
      a.Openstack.Username,
      a.Openstack.Password,
      a.Openstack.Project,
    },
    a.Openstack.Project,
  ); err != nil {
    return wf.openstack, err
  }

  op = client.(openstack.Openstack)

  return &op, nil
}

func (wf *WorkerFactory) Nuage(a Authenticate) (*nuage.Nuage, error) {
  var (
    client  interface{}
    err	    error
    n	    nuage.Nuage
  )

  if client, err = wf.authenticate(
    nuage.Auth{
      Url:	    a.Nuage.URL,
      Organization: a.Nuage.Organization,
      Username:	    a.Nuage.Username,
      Password:	    a.Nuage.Password,
      EnterpriseId: a.Nuage.EnterpriseID,
    },
    "nuage",
  ); err != nil {
    return wf.nuage, err
  }

  n = client.(nuage.Nuage)

  return &n, nil
}

func (wf *WorkerFactory) Paloalto(a Authenticate) (*paloalto.Paloalto, error) {
  var (
    client  interface{}
    err	    error
    pa	    paloalto.Paloalto
  )

  if client, err = wf.authenticate(
    paloalto.PaloaltoFields{
      Url:	a.Paloalto.URL,
      UserName:	a.Paloalto.Username,
      Password:	a.Paloalto.Password,
      Vsys:	a.Paloalto.Vsys,
    },
    "paloalto",
  ); err != nil {
    return wf.paloalto, err
  }

  pa = client.(paloalto.Paloalto)

  return &pa, nil
}

func (wf *WorkerFactory) Bigip(a Authenticate) (*bigip.Bigip, error) {
  var b bigip.Bigip

  b = bigip.Init(bigip.Bigip{
    Url:      a.Bigip.URL,
    User:     a.Bigip.Username,
    Password: a.Bigip.Password,
  })

  return &b, nil
}

func (wf *WorkerFactory) Wap(a Authenticate) (*gowapclient.WAP, error) {
  var (
    client  interface{}
    err	    error
    wap	    gowapclient.WAP
  )

  if client, err = wf.authenticate(
    gowapclient.WAPConfig{
      WAPAuthHost:    a.Wap.AuthURL,
      WAPAdminHost:   a.Wap.AdminURL,
      WAPTenantHost:  a.Wap.TenantURL,
      WAPUserName:    a.Wap.Username,
      WAPPassword:    a.Wap.Password,
      WAPPlanId:      a.Wap.PlanID,
      WAPSMAHost:     a.Wap.SmaURL,
    },
    "wap",
  ); err != nil {
    return wf.wap, err
  }

  wap = client.(gowapclient.WAP)

  return &wap, nil
}

func (wf *WorkerFactory) GetTransactionID() string {
  return wf.transactionID
}

func (wf *WorkerFactory) SetTransactionID(id string) {
  wf.transactionID = id
}

func (wf *WorkerFactory) authenticate(auth interface{},	control string) (interface{}, error) {
  var s singleton.Singleton = singleton.Singleton{
    Auth:		auth,
    Expire:		config.EnvRedis.Expire,
    ConfigCache:	redis.Config{},
    ForceAuthenticate:	FORCE_AUTHENTICATE,
    ClientCache:	config.EnvSingletons.RedisConnection,
  }

  return s.CacheControl(control)
}

type Worker interface {
  Create(f Factorier, a Authenticate) StatusConsumer
  Delete(f Factorier, a Authenticate) StatusConsumer
}

type Decorator struct {
  Worker
}

type Log struct {
  Decorator
}

func NewLog(w Worker) Worker {
  return &Log{Decorator{w}}
}

func (l *Log) Create(f Factorier, a Authenticate) StatusConsumer {
  var sc StatusConsumer
  config.EnvSingletons.Logger.Infof(log.TEMPLATE_LOG_CORE, f.GetTransactionID(), PACKAGE, "Create", "Log", log.INIT, l.Worker, log.EMPTY_STR)

  sc = l.Worker.Create(f, a)

  if sc.Status == COMPLETED || sc.Status == IN_PROGRESS {
    config.EnvSingletons.Logger.Infof(log.TEMPLATE_LOG_CORE, f.GetTransactionID(), PACKAGE, "Create", "Log", log.DONE, l.Worker, log.EMPTY_STR)
  } else {
    config.EnvSingletons.Logger.Errorf(log.TEMPLATE_LOG_CORE, f.GetTransactionID(), PACKAGE, "Create", "Log", log.DONE, l.Worker, sc.Error.Error())
  }

  return sc
}

func (l *Log) Delete(f Factorier, a Authenticate) StatusConsumer {
  var sc StatusConsumer
  config.EnvSingletons.Logger.Infof(log.TEMPLATE_LOG_CORE, f.GetTransactionID(), PACKAGE, "Delete", "Log", log.INIT, l.Worker, log.EMPTY_STR)

  sc = l.Worker.Delete(f, a)

  if sc.Status == COMPLETED || sc.Status == IN_PROGRESS {
    config.EnvSingletons.Logger.Infof(log.TEMPLATE_LOG_CORE, f.GetTransactionID(), PACKAGE, "Delete", "Log", log.DONE, l.Worker, log.EMPTY_STR)
  } else {
    config.EnvSingletons.Logger.Errorf(log.TEMPLATE_LOG_CORE, f.GetTransactionID(), PACKAGE, "Delete", "Log", log.DONE, l.Worker, sc.Error.Error())
  }

  return sc
}
