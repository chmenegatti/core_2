package config

import (
	"github.com/vmware/govmomi"
	"github.com/vmware/govmomi/vim25/soap"
	"context"
	"net/url"
)

func LoadVMWare() (err error) {
	var (
		u *url.URL
	)

	if u, err = soap.ParseURL(EnvConfig.VMWareURL); err != nil {
		return
	}

	u.User = url.User(EnvConfig.VMWareUserName)
	u.User = url.UserPassword(u.User.Username(), EnvConfig.VMWarePassword)
	EnvSingletons.Context = context.WithValue(context.Background(), "datacenter", EnvConfig.VMWareDefaultDatacenter)

	EnvSingletons.VMWare, err = govmomi.NewClient(EnvSingletons.Context, u, EnvConfig.VMWareInsecure)
	return
}
