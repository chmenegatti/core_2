package core

import (
	"gitlab.com/ascenty/core/config"
	"net/http/httptest"
	"net/http"
	"testing"
	"fmt"
)

func startServerAuthenticate() *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var (
			resp  []byte
		)

		switch r.RequestURI {
		case "//api/internal/session":
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

	config.EnvConfig.VMWareURL = "https://vcsa-nsx.datacenter.local/sdk"
	config.EnvConfig.VMWareUserName = "Administrator@vsphere.local"
	config.EnvConfig.VMWarePassword = "Totvs@123"
	config.EnvConfig.VMWareInsecure = true

	config.EnvConfig.JCStackURL = "https://172.18.204.10"
	config.EnvConfig.JCStackUsername = "nemesis"
	config.EnvConfig.JCStackPassword = "xKBG}Lu8?{3.?mxc"
	config.EnvConfig.JCStackServer = "nostromo"

	config.EnvAmqp.Hosts = []string{"amqp://"}

	config.LoadLogger()
	config.LoadRedis()
}

func Test_WorkerFactory_AuthenticateRubrik(t *testing.T) {
	var (
		server	*httptest.Server
		wf	WorkerFactory
		a	Authenticate
	)

	loadConfWorker()
	server = startServerAuthenticate()

	a.Rubrik = RubrikAuthenticate{
		Clusters: map[string][]string{
			"tesp2": []string{server.URL + "/"},
		},
		Username: "MOCK",
		Password: "MOCK",
	}

	wf.Rubrik(a, "tesp2")
}

func Test_WorkerFactory_VMWare(t *testing.T) {
	var (
		wf	WorkerFactory
	)

	loadConfWorker()
	fmt.Println(wf.VMWare())
}

func Test_WorkerFactory_JCStack(t *testing.T) {
	var (
		wf	WorkerFactory
		a	Authenticate
	)

	loadConfWorker()

	a.JCStack = JCStackAuthenticate{
		URL:	  config.EnvConfig.JCStackURL,
		Username: config.EnvConfig.JCStackUsername,
		Password: config.EnvConfig.JCStackPassword,
		Server:	  config.EnvConfig.JCStackServer,
	}

	fmt.Println(wf.JCStack(a))
}
