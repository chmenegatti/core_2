package tasks

import (
	"gitlab.com/ascenty/core/config"
	"gitlab.com/ascenty/go-cache"
	"github.com/vmware/govmomi/vim25"
	"github.com/vmware/govmomi/view"
	"github.com/vmware/govmomi/vim25/types"

	"strings"
	"context"
	"fmt"
)

const (
	DEFAULT_EXPIRATION_TASK	= 3600
	MAX_TASKS		= 200
)

func RegisterVirtualMachineTasks(virtualMachineName, taskId string) (err error) {
	var (
		c	cache.Cache
		value	interface{}
		exists	bool
	)

	c = cache.Cache{
		Key:	virtualMachineName,
		Client:	config.EnvSingletons.RedisConnection,
	}

	if value, err, exists = c.Get(); err != nil {
		return
	}

	if !exists {
		c.Value = taskId
		c.Time = DEFAULT_EXPIRATION_TASK

		if _, err = c.Set(); err != nil {
			return
		}
	} else {
		var tasks = value.(string)

		c.Value = fmt.Sprintf("%s,%s", tasks, taskId)
		c.Time = DEFAULT_EXPIRATION_TASK

		if err = c.Edit(); err != nil {
			return
		}
	}

	return
}

func CheckIfExitsTasksRunningOrQueued(virtualMachineName string, vsphereClient *vim25.Client, ctx context.Context) (existsTask bool, err error) {
	var (
		c	  cache.Cache
		value	  interface{}
		existsKey bool
	)

	c = cache.Cache{
		Key:	virtualMachineName,
		Client:	config.EnvSingletons.RedisConnection,
	}

	if value, err, existsKey = c.Get(); err != nil {
		return
	}

	if existsKey {
		var tasksVirtualMachine = value.(string)
		var m = vsphereClient.ServiceContent.TaskManager
		var watch = *m

		var tv *view.TaskView

		if tv, err = view.NewManager(vsphereClient).CreateTaskView(ctx, &watch); err != nil {
			return
		}

		defer func() {
			_ = tv.Destroy(context.Background())
		}()

		err = tv.Collect(ctx, func(tasks []types.TaskInfo) {
			if len(tasks) > MAX_TASKS {
			        tasks = tasks[len(tasks)-MAX_TASKS:]
			}

			for _, t := range tasks {
				fmt.Println(t.Key)
				for _, tvm := range strings.Split(tasksVirtualMachine, ",") {
					if strings.ToLower(t.Key) == strings.ToLower(tvm) && (t.State == types.TaskInfoStateQueued || t.State == types.TaskInfoStateRunning) {
						existsTask = true
						break
					}
				}
			}
		})
	}

	if existsKey && !existsTask {
		go DeleteVirtualMachineTasks(virtualMachineName)
	}

	return
}

func DeleteVirtualMachineTasks(virtualMachineName string) (err error) {
	var (
		c	cache.Cache
		exists	bool
	)

	c = cache.Cache{
		Key:	virtualMachineName,
		Client:	config.EnvSingletons.RedisConnection,
	}

	if _, err, exists = c.Get(); err != nil {
		return
	}

	if exists {
		err = c.Del()
	}

	return
}
