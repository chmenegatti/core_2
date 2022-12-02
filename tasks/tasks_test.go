package tasks

import (
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/object"
	"github.com/vmware/govmomi/find"
	"gitlab.com/ascenty/go-cache/redis"
	"gitlab.com/ascenty/core/config"
	"github.com/vmware/govmomi/vim25/soap"
	"net/url"
	"context"
	"testing"
	"time"
)

var (
	client  *govmomi.Client
)

func loadConfCore() {
	config.EnvConfig.SyslogLevel = "DEBUG"
	config.EnvConfig.SyslogNetwork = "udp"
	config.EnvConfig.SyslogRaddr = "localhost:514"
	config.EnvConfig.SyslogTag = "core"
	config.EnvConfig.SyslogFacility = "local6"
	config.EnvConfig.CheckURL = "v1/health"
	config.EnvConfig.CheckPort = "8000"
	config.EnvConfig.Edge = "Teste"
	config.EnvConfig.SlackToken = ""
	config.EnvConfig.SlackChannel = ""

	config.EnvRedis.Hosts = map[string]string{"localhost": "127.0.0.1:6379"}
	config.EnvRedis.Timeout = 5
	config.EnvRedis.Retries = 10
	config.EnvRedis.Heartbeat = 60
	config.EnvRedis.Expire = 3000

	config.EnvConfig.VMWareDefaultDatacenter = "NSX-DATACENTER"
	config.EnvSingletons.Context = context.Background()
	config.EnvConfig.VMWareURL = "https://vcsa-nsx.datacenter.local/sdk"
	config.EnvConfig.VMWareUserName = "Administrator@vsphere.local"
	config.EnvConfig.VMWarePassword = "Totvs@123"
	config.EnvConfig.VMWareInsecure = true
	config.EnvConfig.VMWareVirtualMachineTimeout = 900

	config.LoadLogger()

	var (
		err	error
		ctx     context.Context
		u       *url.URL
	)

	if config.EnvSingletons.RedisConnection, err = redis.Init(redis.Config{
		RedisHosts:	config.EnvRedis.Hosts,
		RedisTimeout:	config.EnvRedis.Timeout,
		RedisRetries:	config.EnvRedis.Retries,
		RedisHeartbeat:	time.Duration(config.EnvRedis.Heartbeat) * time.Second,
	}); err != nil {
		panic(err)
	}

	if u, err = soap.ParseURL(config.EnvConfig.VMWareURL); err != nil {
		panic(err)
	}

	u.User = url.User(config.EnvConfig.VMWareUserName)
	u.User = url.UserPassword(u.User.Username(), config.EnvConfig.VMWarePassword)
	ctx = context.WithValue(context.Background(), "datacenter", config.EnvConfig.VMWareDefaultDatacenter)
	config.EnvSingletons.Context = ctx

	if client, err = govmomi.NewClient(config.EnvSingletons.Context, u, true); err != nil {
		panic(err)
	}

	var finder = find.NewFinder(client.Client, false)
	var datacenter *object.Datacenter

	if datacenter, err = finder.Datacenter(config.EnvSingletons.Context, config.EnvConfig.VMWareDefaultDatacenter); err != nil {
		panic(err)
	}
	finder.SetDatacenter(datacenter)

	return
}

func Test_RegisterVirtualMachineTasks(t *testing.T) {
	loadConfCore()

	if err := RegisterVirtualMachineTasks("teste-vm", "task-67403"); err != nil {
		t.Fatal(err)
	}
}

func Test_CheckIfExitsTasksRunningOrQueued(t *testing.T) {
	loadConfCore()

	if _, err := CheckIfExitsTasksRunningOrQueued("teste-vm", client.Client, config.EnvSingletons.Context); err != nil {
		t.Fatal(err)
	}
}

func Test_DeleteVirtualMachineTasks(t *testing.T) {
	loadConfCore()

	if err := DeleteVirtualMachineTasks("teste-vm"); err != nil {
		t.Fatal(err)
	}
}
