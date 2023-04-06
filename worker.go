package core

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/vim25/soap"
	"gitlab.com/ascenty/core/config"
	"gitlab.com/ascenty/core/log"
	"gitlab.com/ascenty/go-cache/redis"
	"gitlab.com/ascenty/go-jcstack"
	"gitlab.com/ascenty/go-singleton"
	"gitlab.com/ascenty/paloalto"
	"gitlab.com/ascenty/rubrik-golang"

	_ "github.com/go-sql-driver/mysql"
	"github.com/jinzhu/gorm"
)

const (
	FORCE_AUTHENTICATE = true
)

type Authenticate struct {
	sync.RWMutex

	DB       DBAuthenticate
	Dbaas    DbaasAuthenticate
	Paloalto PaloaltoAuthenticate
	JCStack  JCStackAuthenticate
}

type DBAuthenticate struct {
	Connection *gorm.DB
}

type DbaasAuthenticate struct {
	URL string
}

type PaloaltoAuthenticate struct {
	URL      string `json:",omitempty"`
	Username string `json:"-"`
	Password string `json:"-"`
	Vsys     string `json:"-"`
}

type JCStackAuthenticate struct {
	URL      string `json:",omitempty"`
	Username string `json:",omitempty"`
	Password string `json:",omitempty"`
	Server   string `json:",omitempty"`
}

type Factorier interface {
	DB(Authenticate) (*gorm.DB, error)
	Rubrik(string) (*rubrik.Rubrik, error)
	Paloalto(Authenticate) (paloalto.Paloalto, error)
	JCStack(Authenticate) (*jcstack.JCStack, error)
	VMWare() (*govmomi.Client, error)
	SessionS3() (*session.Session, error)
	GetTransactionID() string
	SetTransactionID(string)
}

type Factory struct{}

func (f *Factory) DB(a Authenticate) (*gorm.DB, error) {
	panic("Method DB not implemented")
}

func (f *Factory) Rubrik(cluster string) (*rubrik.Rubrik, error) {
	panic("Method Rubrik not implemented")
}

func (f *Factory) Paloalto(a Authenticate) (paloalto.Paloalto, error) {
	panic("Method Paloalto not implemented")
}

func (f *Factory) JCStack(a Authenticate) (*jcstack.JCStack, error) {
	panic("Method JCStack not implemented")
}

func (f *Factory) VMWare() (*govmomi.Client, error) {
	panic("Method VMWare not implemented")
}

func (f *Factory) SessionS3() (*session.Session, error) {
	panic("Method SessionS3 not implemented")
}

func (f *Factory) GetTransactionID() string {
	panic("Method GetTransactionID not implemented")
}

func (f *Factory) SetTransactionID(id string) {
	panic("Method SetTransactionID not implemented")
}

type WorkerFactory struct {
	Factory

	db            *gorm.DB
	rubrik        *rubrik.Rubrik
	jcstack       *jcstack.JCStack
	vmware        *govmomi.Client
	sess          *session.Session
	transactionID string
}

func (wf *WorkerFactory) Rubrik(cluster string) (*rubrik.Rubrik, error) {
	var (
		client interface{}
		err    error
		r      *rubrik.Rubrik
		ok     bool
	)

	if _, ok = config.EnvConfig.Rubrik[cluster]; !ok {
		return nil, errors.New(fmt.Sprintf("%s is not mapped", cluster))
	}

	if client, err = wf.authenticate(
		rubrik.RubrikFields{
			Cluster:    config.EnvConfig.Rubrik[cluster].Cluster,
			Username:   config.EnvConfig.Rubrik[cluster].Username,
			Password:   config.EnvConfig.Rubrik[cluster].Password,
			Expiration: config.EnvConfig.Rubrik[cluster].Expiration,
		},
		fmt.Sprintf("rubrik-%s", cluster),
	); err != nil {
		return wf.rubrik, err
	}

	r = client.(*rubrik.Rubrik)

	return r, nil
}

func (wf WorkerFactory) JCStack(a Authenticate) (*jcstack.JCStack, error) {
	var (
		client interface{}
		err    error
		jc     *jcstack.JCStack
	)

	if client, err = wf.authenticate(
		jcstack.JCStackConfig{
			URL:      a.JCStack.URL,
			Username: a.JCStack.Username,
			Password: a.JCStack.Password,
			Server:   a.JCStack.Server,
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
		URL:      a.Paloalto.URL,
		Username: a.Paloalto.Username,
		Password: a.Paloalto.Password,
		Vsys:     a.Paloalto.Vsys,
	})

	return p, err
}

func (wf *WorkerFactory) VMWare() (*govmomi.Client, error) {
	var (
		u      *url.URL
		client *govmomi.Client
		err    error
	)

	if u, err = soap.ParseURL(config.EnvConfig.VMWareURL); err != nil {
		return nil, err
	}

	u.User = url.User(config.EnvConfig.VMWareUserName)
	u.User = url.UserPassword(u.User.Username(), config.EnvConfig.VMWarePassword)
	config.EnvSingletons.Context = context.WithValue(context.Background(), "datacenter", config.EnvConfig.VMWareDefaultDatacenter)

	client, err = govmomi.NewClient(config.EnvSingletons.Context, u, config.EnvConfig.VMWareInsecure)

	return client, err
}

func (wf *WorkerFactory) SessionS3() (*session.Session, error) {
	var (
		sess *session.Session
		err  error
	)

	tr := &http.Transport{}
	client := &http.Client{}

	if strings.Contains(config.EnvConfig.OntapS3URL, "tesp4infra.local") {
		tr = &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		}
		client = &http.Client{Transport: tr}
	} else {
		var clientCertFile, clientKeyFile []byte

		clientKeyFile = []byte(`-----BEGIN PRIVATE KEY-----
MIIEvQIBADANBgkqhkiG9w0BAQEFAASCBKcwggSjAgEAAoIBAQDXr0CDF7Fyns10
U3B6Mk1eQyKju4tkXkugmxJRSa0evedxaJEz6I/yTbx/SuVuc2pY+UnQbTsaYt/W
SJpCmJys3/xTeiNQZZqU5zhxGv/jsIXyRof6nFsoH19TKfz8maYP8/KzZ1ntO6dv
JUsOEm9avoFUfgqIagK3db17e2zC6h2Hqreth5C16Z9GLdONITabhteo58Q2Ay9o
HmERHOx4huibLAsomiaG4zKOoAnvEP2x4VHnQdeTXGmlQu5GYYxcmGaMgjw15NxH
yPkRIltdZlEUTqJP/cexjWg/+GKTAMlir/mV4/c39KKK2I3cjDLexJdqOex395Ag
JvPKNKcvAgMBAAECggEANFFNa78eivgcTCBFQ07aV4gfaen/KOx6mc2jxtYBSVi1
QFwmBJpf+F2E4LexKXaTkFt/3S/xyze6pgbcbFUKhCCc2z7gPAs5UO85UK5E7d0O
0LLtHww4TGY3wDuKI1e94VbCQ2RJevMibSJ1r0cnfbKCOpWMRoS4ffnwaXiJ7Eld
rVYZZXqdTOeFiX8CEg4xZAEkTi0PmylhiFya/GlWqn4/N0Od932jLOwB4ettjXra
8fqxyXHMTo1E1AJcZEmfKjp0nyrH/p1v5qJjAj7vcXbukx5l4tI77YS+K1kfeh7r
8960eJv8RIneh8U1gU3aySQ9UAPaVyh7rMnLQzaJyQKBgQDyS3gBqJrEyXyFYV6O
cm76ETbpJb8QsW95ACcVbJb5TEj/yFBFM/DJSJOXjlMQLQUd+n0HRNmyPSXvR49s
HpsiwWzVvlL4N2kIjCBfJEwXibFrTc+DAUklJqRiYzlqeYitxwcpbLp9TVpeiLho
Qbwm+aga9JdNTc5+gMcXfoO/xQKBgQDj4nS5aGWsnBJdsJ//HeFpvXcK4cMbUYtw
rm70y5oDiW9mlcesE0wQoRaBSOg5NTq2beJjE9kFPIzmdyh2gIGLw7FwFr3DrHX1
AIeDlESIId4njtvc/S/ZHzbXQb5T6y5Jll3o05iLkt4b+vrQbJbggwHgJzlMS4ey
WVDiIAVmYwKBgQC6gYtDw5Q27F20kgnmHoH8benVt9+GKv8CpjJvlGH2TllWwsu/
KUcsuXgazkx0BCOPOvSo9r+YKuebc3scH8cNDtUHBvT46jYohSyZ7+e+qpfuYDve
vMugFEmvDm/w3NJv6edCZNZ8g2GPqTBB2G/LlJto/GPG9qB/0gLuu80QfQKBgBQS
5a9TZ+ltkhyYq29gpXOYEf1uZ65nX34cj3+W08lN+PczaHRa+s9YMVBQKMypSLO7
OD78B5jzfYHrqy/NIBw8r8us5Shwb6+WTVmk3OiU+ynG5s0vrGBW5JDPpMFlrR0k
Vr9krEJXPKOAV0m21w+N2sNxERYMbTajsYFJys/7AoGAQSHJALmNdGPwCPL4a2R/
LxBflkNg4FUsScFcLjXiyRfMG11/DX2L+PZaT917HOBuc9vrDpc0oflMTlq9R7vf
f6IlAEDzTJsMHWlmmHi0tmb2+531iVYaOSk6W+R2dzCDG+anDayku8nEVW5BPHs4
9svMfLAA0W58IF2l+uF15Yc=
-----END PRIVATE KEY-----`)

		clientCertFile = []byte(`-----BEGIN CERTIFICATE-----
MIIEFTCCAv2gAwIBAgIUM1oIMu1wpYn6/Eg2wj07k8pfG34wDQYJKoZIhvcNAQEL
BQAwgagxCzAJBgNVBAYTAlVTMRMwEQYDVQQIEwpDYWxpZm9ybmlhMRYwFAYDVQQH
Ew1TYW4gRnJhbmNpc2NvMRkwFwYDVQQKExBDbG91ZGZsYXJlLCBJbmMuMRswGQYD
VQQLExJ3d3cuY2xvdWRmbGFyZS5jb20xNDAyBgNVBAMTK01hbmFnZWQgQ0EgODkx
NDNmOTgwNWNjNzMyMWZiZmRjYWVmY2NjYjkwMTcwHhcNMjMwMzEzMTc0OTAwWhcN
MzgwMzA5MTc0OTAwWjAiMQswCQYDVQQGEwJVUzETMBEGA1UEAxMKQ2xvdWRmbGFy
ZTCCASIwDQYJKoZIhvcNAQEBBQADggEPADCCAQoCggEBANevQIMXsXKezXRTcHoy
TV5DIqO7i2ReS6CbElFJrR6953FokTPoj/JNvH9K5W5zalj5SdBtOxpi39ZImkKY
nKzf/FN6I1BlmpTnOHEa/+OwhfJGh/qcWygfX1Mp/PyZpg/z8rNnWe07p28lSw4S
b1q+gVR+CohqArd1vXt7bMLqHYeqt62HkLXpn0Yt040hNpuG16jnxDYDL2geYREc
7HiG6JssCyiaJobjMo6gCe8Q/bHhUedB15NcaaVC7kZhjFyYZoyCPDXk3EfI+REi
W11mURROok/9x7GNaD/4YpMAyWKv+ZXj9zf0oorYjdyMMt7El2o57Hf3kCAm88o0
py8CAwEAAaOBuzCBuDATBgNVHSUEDDAKBggrBgEFBQcDAjAMBgNVHRMBAf8EAjAA
MB0GA1UdDgQWBBSg2I/i93KEwa/OZFkoj7wnR3MXmTAfBgNVHSMEGDAWgBTEyrHq
4uXgUkGS2rtaRhSb+y57FzBTBgNVHR8ETDBKMEigRqBEhkJodHRwOi8vY3JsLmNs
b3VkZmxhcmUuY29tLzA0MWM4YzQzLWIzZTYtNGVlNi1hOThjLTUyMWY0OWI3NmRl
OS5jcmwwDQYJKoZIhvcNAQELBQADggEBACIfAPn7GbBI4PvklwgECmuQwc2CVlyA
CTIBf4J7Ww4aOulCAT9KlrzaeiH1UpXwrvKzX5blGBWcLTue4pFgNL+1UKHSSOgC
5RLfQDn9ENOaSn7/Telk/5O/wEW29VsLNhW0thGmFGC3bNgdE8djy95zACSJqMvj
LAFPI9Gp+bMbmTBy2uWnDZPiK1Boll5ruyOP4ipAo3TH2ophQYW6jHRH33Wfa7f7
498JTAy9wheVRfFSvXGnTzWt/S8s7HhXmtHx8pRLXxOnDld0MdokTAI7JvA7toTA
hnQ+uv4D2eYW3COBKfdxC/N2SqjV6tJRXSCmdOq3s+lqWKpue4jhZfE=
-----END CERTIFICATE-----`)

		clientCert, err := tls.X509KeyPair(clientCertFile, clientKeyFile)
		if err != nil {
			return sess, err
		}

		tlsCfg := &tls.Config{
			Certificates: []tls.Certificate{
				clientCert,
			},
		}

		tlsCfg.BuildNameToCertificate()

		tr = &http.Transport{
			TLSClientConfig: tlsCfg,
		}
		client = &http.Client{Transport: tr}
	}

	sess, err = session.NewSession(&aws.Config{
		CredentialsChainVerboseErrors: aws.Bool(true),
		Region:                        aws.String("*"),
		Endpoint:                      aws.String(config.EnvConfig.OntapS3URL),
		S3ForcePathStyle:              aws.Bool(true),
		DisableSSL:                    aws.Bool(config.EnvConfig.OntapS3DisableSsl),
		Credentials: credentials.NewStaticCredentialsFromCreds(credentials.Value{
			AccessKeyID:     config.EnvConfig.OntapS3AccessKeyID,
			SecretAccessKey: config.EnvConfig.OntapS3SecretAccessKey,
		}),
		HTTPClient: client,
	})

	return sess, err
}

func (wf *WorkerFactory) GetTransactionID() string {
	return wf.transactionID
}

func (wf *WorkerFactory) SetTransactionID(id string) {
	wf.transactionID = id
}

func (wf *WorkerFactory) authenticate(auth interface{}, control string) (interface{}, error) {
	var s singleton.Singleton = singleton.Singleton{
		Auth:              auth,
		Expire:            config.EnvRedis.Expire,
		ConfigCache:       redis.Config{},
		ForceAuthenticate: FORCE_AUTHENTICATE,
		ClientCache:       config.EnvSingletons.RedisConnection,
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
