package core

import (
	"sync"

	"gitlab.com/intera/dbaas"
	"gitlab.com/intera/rubrik-golang"
	"gitlab.com/ascenty/paloalto"
	"gitlab.com/ascenty/go-singleton"
	"gitlab.com/ascenty/go-cache/redis"
	"gitlab.com/ascenty/go-jcstack"
	"gitlab.com/ascenty/core/config"
	"gitlab.com/ascenty/core/log"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

const (
	FORCE_AUTHENTICATE  = true
)

type Authenticate struct {
	sync.RWMutex

	DB	  DBAuthenticate
	Rubrik	  RubrikAuthenticate
	Dbaas	  DbaasAuthenticate
	Paloalto  PaloaltoAuthenticate
	JCStack	  JCStackAuthenticate
}

type RubrikAuthenticate struct {
	Clusters    []string  `json:",omitempty"`
	Username    string    `json:",omitempty"`
	Password    string    `json:",omitempty"`
	Expiration  int32     `json:",omitempty"`
}

type DBAuthenticate struct {
	Connection  *gorm.DB
}

type DbaasAuthenticate struct {
	URL string
}

type PaloaltoAuthenticate struct {
	URL	  string  `json:",omitempty"`
	Username  string  `json:"-"`
	Password  string  `json:"-"`
	Vsys	  string  `json:"-"`
}

type JCStackAuthenticate struct {
	URL	  string  `json:",omitempty"`
	Username  string  `json:",omitempty"`
	Password  string  `json:",omitempty"`
}

type Factorier interface {
	DB(Authenticate) (*gorm.DB, error)
	Rubrik(Authenticate) (*rubrik.Rubrik, error)
	Dbaas(Authenticate) (*dbaas.Dbaas, error)
	Paloalto(Authenticate) (paloalto.Paloalto, error)
	JCStack(Authenticate) (*jcstack.JCStack, error)
	GetTransactionID() string
	SetTransactionID(string)
}

type Factory struct {}

func (f *Factory) DB(a Authenticate) (*gorm.DB, error) {
	panic("Method DB not implemented")
}

func (f *Factory) Rubrik(a Authenticate) (*rubrik.Rubrik, error) {
	panic("Method Rubrik not implemented")
}

func (f *Factory) Dbaas(a Authenticate) (*dbaas.Dbaas, error) {
	panic("Method Dbaas not implemented")
}

func (f *Factory) Paloalto(a Authenticate) (paloalto.Paloalto, error) {
	panic("Method Paloalto not implemented")
}

func (f *Factory) JCStack(a Authenticate) (*jcstack.JCStack, error) {
	panic("Method Paloalto not implemented")
}

func (f *Factory) GetTransactionID() string {
	panic("Method GetTransactionID not implemented")
}

func (f *Factory) SetTransactionID(id string) {
	panic("Method SetTransactionID not implemented")
}

type WorkerFactory struct {
	Factory

	db		*gorm.DB
	rubrik		*rubrik.Rubrik
	dbaas		*dbaas.Dbaas
	jcstack		*jcstack.JCStack
	transactionID	string
}

func (wf *WorkerFactory) Dbaas(a Authenticate) (*dbaas.Dbaas, error) {
	return dbaas.Authenticate(dbaas.DbaasConfig{
		URL:  a.Dbaas.URL,
	})
}

func (wf *WorkerFactory) Rubrik(a Authenticate) (*rubrik.Rubrik, error) {
	var (
		client  interface{}
		err	error
		r	*rubrik.Rubrik
	)

	if client, err = wf.authenticate(
		rubrik.RubrikFields{
			Cluster:    a.Rubrik.Clusters,
			Username:   a.Rubrik.Username,
			Password:   a.Rubrik.Password,
			Expiration: a.Rubrik.Expiration,
		},
		"rubrik",
	); err != nil {
		return wf.rubrik, err
	}

	r = client.(*rubrik.Rubrik)

	return r, nil
}

func (wf WorkerFactory) JCStack(a Authenticate) (*jcstack.JCStack, error) {
	var (
		client	interface{}
		err	error
		jc	*jcstack.JCStack
	)

	if client, err = wf.authenticate(
		jcstack.JCStackConfig{
			URL:  a.JCStack.URL,
			Username: a.JCStack.Username,
			Password: a.JCStack.Password,
		},
		"jcstack",
	); err != nil {
		return wf.jcstack, err
	}

	jc = client.(*jcstack.JCStack)

	return jc, nil
}

func (wf *WorkerFactory) DB(a Authenticate) (*gorm.DB, error) {
	var tx *gorm.DB = a.DB.Connection.Begin()

	if tx.Error != nil {
		return a.DB.Connection, tx.Error
	}

	return tx, nil
}

func (wf *WorkerFactory) Paloalto(a Authenticate) (paloalto.Paloalto, error) {
	var (
		p   paloalto.Paloalto
		err error
	)

	p, err = paloalto.NewClient(paloalto.PaloaltoConfig{
		URL:	  a.Paloalto.URL,
		Username: a.Paloalto.Username,
		Password: a.Paloalto.Password,
		Vsys:	  a.Paloalto.Vsys,
	})

	return p, err
}

func (wf *WorkerFactory) GetTransactionID() string {
	return wf.transactionID
}

func (wf *WorkerFactory) SetTransactionID(id string) {
	wf.transactionID = id
}

func (wf *WorkerFactory) authenticate(auth interface{},	control string) (interface{}, error) {
	var s singleton.Singleton = singleton.Singleton{
		Auth:		    auth,
		Expire:		    config.EnvRedis.Expire,
		ConfigCache:	    redis.Config{},
		ForceAuthenticate:  FORCE_AUTHENTICATE,
		ClientCache:	    config.EnvSingletons.RedisConnection,
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
