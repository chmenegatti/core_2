package core

import (
  "git-devops.totvs.com.br/intera/core/config"
  "git-devops.totvs.com.br/intera/openstack"
  "git-devops.totvs.com.br/intera/go-wap-client"
  "encoding/json"
  "net/http/httptest"
  "net/http"
  "testing"
)

type NuageAuth struct {
  User		string	`json:"userName"`
  Key		string	`json:"APIKey"`
  Id		string	`json:"ID"`
  TokenExpiry	int	`json:"APIKeyExpiry"`
  EnterpriseId	string	`json:"enterpriseID"`
}

func startServerAuthenticate() *httptest.Server {
  return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
    var (
      resp  []byte
      err   error
    )

    switch r.RequestURI {
    case "/:5000/v3/auth/tokens":
      var op = openstack.RespAuthenticate {
	Token:	"TOKEN",
      }

      if resp, err = json.Marshal(op); err != nil {
	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	return
      }

      w.Header().Set("X-Subject-Token", "TOKEN")
      w.WriteHeader(http.StatusCreated)
      w.Write(resp)
    case "/:8443/nuage/api/v4_0/me":
      var n = []NuageAuth{
	{User:	"USER", Key:  "KEY", Id: "ID", TokenExpiry: 3000, EnterpriseId: "ENTERPRISEID"},
      }

      if resp, err = json.Marshal(n); err != nil {
	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	return
      }

      w.WriteHeader(http.StatusOK)
      w.Write(resp)
    case "//api/?type=keygen&user=MOCK&password=MOCK":
      resp = []byte(`<response status = 'success'><result><key>KEY</key></result></response>`)

      w.WriteHeader(http.StatusOK)
      w.Write(resp)
    case "//WAPTokenGenerator.svc/":
      var wap = gowapclient.RespAuthenticate{
	Token:	"TOKEN",
      }

      if resp, err = json.Marshal(wap); err != nil {
	http.Error(w, "Internal Server Error", http.StatusInternalServerError)
	return
      }

      w.WriteHeader(http.StatusOK)
      w.Write(resp)
    case "//api/v1/session":
      resp = []byte(`{"id":"64d607b6-4104-485c-9075-ff8ba51a4fd3","token":"eyJ0eXAiOiJKV1QiLCJhbGciOiJIUzI1NiJ9.eyJzdWIiOiI5MWU3ZWQxOS1kMWIyLTQyNTAtOTM0YS1mMzFiZTBhNmUwMDIiLCJpc3MiOiJmNGQ4MWE0OC0wZDQ1LTQwODktYTYyOS0wNjNjMThlYmIwN2YiLCJqdGkiOiI2NGQ2MDdiNi00MTA0LTQ4NWMtOTA3NS1mZjhiYTUxYTRmZDMifQ.pMh8BsfJx-DPdQ6TE25xQsl1a4LMID327fk7TVsg6nE","userId":"91e7ed19-d1b2-4250-934a-f31be0a6e002","expiration":"2019-04-02T13:00:28Z","organizationId":"Organization:::acc455b4-0b0f-422b-ae2c-0da83f2814a3","Cluster":"https:g/172.18.165.20"}`)
      w.WriteHeader(http.StatusOK)
      w.Write(resp)
    }
  }))
}

func loadConfWorker() {
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

  config.EnvAmqp.Hosts = []string{"amqp://"}

  config.LoadLogger()
  config.LoadRedis()
}

func Test_WorkerFactory_AuthenticateOpenstack(t *testing.T) {
  var (
    server  *httptest.Server
    wf	    WorkerFactory
    a	    Authenticate
  )

  loadConfWorker()
  server = startServerAuthenticate()

  a.Openstack = OpenstackAuthenticate{
    URL:      server.URL + "/",
    Username: "MOCK",
    Password: "MOCK",
    Project:  "mock",
  }

  wf.Openstack(a)
}

func Test_WorkerFactory_AuthenticateNuage(t *testing.T) {
  var (
    server  *httptest.Server
    wf	    WorkerFactory
    a	    Authenticate
  )

  loadConfWorker()
  server = startServerAuthenticate()

  a.Nuage = NuageAuthenticate{
    URL:	  server.URL + "/",
    Organization: "MOCK",
    Username:	  "MOCK",
    Password:	  "MOCK",
    EnterpriseID: "MOCK",
  }

  wf.Nuage(a)
}

func Test_WorkerFactory_AuthenticatePaloalto(t *testing.T) {
  var (
    server  *httptest.Server
    wf	    WorkerFactory
    a	    Authenticate
  )

  loadConfWorker()
  server = startServerAuthenticate()

  a.Paloalto = PaloaltoAuthenticate{
    URL:	  server.URL + "/",
    Username:	  "MOCK",
    Password:	  "MOCK",
    Vsys:	  "vsys1",
  }

  wf.Paloalto(a)
}

func Test_WorkerFactory_AuthenticateWap(t *testing.T) {
  var (
    server  *httptest.Server
    wf	    WorkerFactory
    a	    Authenticate
  )

  loadConfWorker()
  server = startServerAuthenticate()

  a.Wap = WapAuthenticate{
    AuthURL:	server.URL + "/",
    AdminURL:	server.URL + "/",
    TenantURL:	server.URL + "/",
    Username:	"MOCK",
    Password:	"MOCK",
    PlanID:	"MOCK",
    SmaURL:	server.URL + "/",
  }

  wf.Wap(a)
}

func Test_WorkerFactory_AuthenticateRubrik(t *testing.T) {
  var (
    server  *httptest.Server
    wf	    WorkerFactory
    a	    Authenticate
  )

  loadConfWorker()
  server = startServerAuthenticate()

  a.Rubrik = RubrikAuthenticate{
    URL:      server.URL + "/",
    Username: "MOCK",
    Password: "MOCK",
  }

  wf.Rubrik(a)
}
