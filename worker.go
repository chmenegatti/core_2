package core

import (
  "errors"
  "sync"

  "git-devops.totvs.com.br/intera/openstack"
  "git-devops.totvs.com.br/intera/nuage"
  "git-devops.totvs.com.br/intera/paloalto"
  "git-devops.totvs.com.br/intera/bigip"
  "git-devops.totvs.com.br/intera/go-wap-client"
  "git-devops.totvs.com.br/intera/go-singleton"
  "git-devops.totvs.com.br/intera/go-cache/redis"
  "git-devops.totvs.com.br/intera/core/config"
  "git-devops.totvs.com.br/intera/core/log"

  _ "github.com/go-sql-driver/mysql"
  "github.com/jinzhu/gorm"
)

const (
  FORCE_AUTHENTICATE  = true
)

type Authenticate struct {
  sync.RWMutex

  Openstack OpenstackAuthenticate
  Nuage	    NuageAuthenticate
  Paloalto  PaloaltoAuthenticate
  Bigip	    map[string]BigipAuthenticate
  Wap	    WapAuthenticate
  DB	    DBAuthenticate
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

type DBAuthenticate struct {
  Connection  *gorm.DB
}

type Factorier interface {
  Openstack(Authenticate) (*openstack.Openstack, error)
  Nuage(Authenticate) (*nuage.Nuage, error)
  Paloalto(Authenticate) (*paloalto.Paloalto, error)
  Bigip(Authenticate) (*bigip.Bigip, error)
  Wap(Authenticate) (*gowapclient.WAP, error)
  DB(Authenticate) (*gorm.DB, error)
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

func (f *Factory) DB(a Authenticate) (*gorm.DB, error) {
  panic("Method DB not implemented")
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
  db		*gorm.DB
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

func (wf *WorkerFactory) Bigip(a Authenticate, context string) (*bigip.Bigip, error) {
  var (
    b	bigip.Bigip
    err	error
  )

  if _, ok := a.Bigip[context]; ok {
    b = bigip.Init(bigip.Bigip{
      Url:      a.Bigip[context].URL,
      User:     a.Bigip[context].Username,
      Password: a.Bigip[context].Password,
    })
  } else {
    err = errors.New("Context not exists")
  }

  return &b, err
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

func (wf *WorkerFactory) DB(a Authenticate) (*gorm.DB, error) {
  var tx *gorm.DB = a.DB.Connection.Begin()

  if tx.Error != nil {
    return a.DB.Connection, tx.Error
  }

  return tx, nil
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
  Custom(f Factorier, a Authenticate) StatusConsumer
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
    var message = log.EMPTY_STR

    if sc.Error != nil {
      message = sc.Error.Error()
    }

    config.EnvSingletons.Logger.Errorf(log.TEMPLATE_LOG_CORE, f.GetTransactionID(), PACKAGE, "Create", "Log", log.DONE, l.Worker, message)
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
    var message = log.EMPTY_STR

    if sc.Error != nil {
      message = sc.Error.Error()
    }

    config.EnvSingletons.Logger.Errorf(log.TEMPLATE_LOG_CORE, f.GetTransactionID(), PACKAGE, "Delete", "Log", log.DONE, l.Worker, message)
  }

  return sc
}

func (l *Log) Custom(f Factorier, a Authenticate) StatusConsumer {
  var sc StatusConsumer
  config.EnvSingletons.Logger.Infof(log.TEMPLATE_LOG_CORE, f.GetTransactionID(), PACKAGE, "Custom", "Log", log.INIT, l.Worker, log.EMPTY_STR)

  sc = l.Worker.Custom(f, a)

  if sc.Status == COMPLETED || sc.Status == IN_PROGRESS {
    config.EnvSingletons.Logger.Infof(log.TEMPLATE_LOG_CORE, f.GetTransactionID(), PACKAGE, "Custom", "Log", log.DONE, l.Worker, log.EMPTY_STR)
  } else {
    var message = log.EMPTY_STR

    if sc.Error != nil {
      message = sc.Error.Error()
    }

    config.EnvSingletons.Logger.Errorf(log.TEMPLATE_LOG_CORE, f.GetTransactionID(), PACKAGE, "Custom", "Log", log.DONE, l.Worker, message)
  }

  return sc
}
