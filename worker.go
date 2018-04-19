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
)

const (
  FORCE_AUTHENTICATE  = true
)

type Authenticate struct {
  Openstack OpenstackAuthenticate
  Nuage	    NuageAuthenticate
  Paloalto  PaloaltoAuthenticate
  BigIP	    BigIPAuthenticate
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
}

type BigIPAuthenticate struct {
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
}

type Factorier interface {
  Openstack(Authenticate) (*openstack.Openstack, error)
  Nuage(Authenticate) (*nuage.Nuage, error)
  Paloalto(Authenticate) (*paloalto.Paloalto, error)
  Bigip(Authenticate) (*bigip.Bigip, error)
  Wap(Authenticate) (*gowapclient.WAP, error)
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

type WorkerFactory struct {
  Factory

  openstack *openstack.Openstack
  nuage	    *nuage.Nuage
  paloalto  *paloalto.Paloalto
  bigip	    *bigip.Bigip
  wap	    *gowapclient.WAP
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
  return wf.paloalto, nil
}

func (wf *WorkerFactory) Bigip(a Authenticate) (*bigip.Bigip, error) {
  return wf.bigip, nil
}

func (wf *WorkerFactory) Wap(a Authenticate) (*gowapclient.WAP, error) {
  return wf.wap, nil
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
  Create(f Factorier, a Authenticate) error
  Delete(f Factorier, a Authenticate) error
}
