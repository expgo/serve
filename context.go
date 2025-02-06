package serve

import (
	"context"
	"fmt"
	"time"

	sync "github.com/expgo/sync/wire"

	"github.com/expgo/log"
	"github.com/thejerf/suture/v4"
)

//go:generate ag

const ServiceTimeout = 10 * time.Second

// @Singleton(localGetter)
type Context struct {
	*suture.Supervisor
	l              log.Logger `new:""`
	cancel         context.CancelFunc
	ctx            context.Context
	errChan        <-chan error
	supervisorMap  map[suture.ServiceToken]*suture.Supervisor
	supervisorLock sync.RWMutex `new:""`
}

func (c *Context) Init() {
	c.Supervisor = suture.New("serve", spec(infoEventHook(c.l)))
	c.ctx, c.cancel = context.WithCancel(context.Background())
	c.errChan = c.ServeBackground(c.ctx)
	c.supervisorMap = make(map[suture.ServiceToken]*suture.Supervisor)
}

func (c *Context) Down() error {
	c.cancel()
	c.supervisorMap = nil
	return <-c.errChan
}

func (c *Context) Add(sup *suture.Supervisor) suture.ServiceToken {
	id := c.Supervisor.Add(sup)

	c.supervisorLock.Lock()
	defer c.supervisorLock.Unlock()

	c.supervisorMap[id] = sup

	return id
}

func (c *Context) Remove(id suture.ServiceToken) error {
	err := c.Supervisor.Remove(id)

	c.supervisorLock.Lock()
	defer c.supervisorLock.Unlock()

	delete(c.supervisorMap, id)

	return err
}

func (c *Context) RemoveAndWait(id suture.ServiceToken, timeout time.Duration) error {
	err := c.Supervisor.RemoveAndWait(id, timeout)

	c.supervisorLock.Lock()
	defer c.supervisorLock.Unlock()

	delete(c.supervisorMap, id)

	return err
}

func spec(eventHook suture.EventHook) suture.Spec {
	return suture.Spec{
		EventHook:                eventHook,
		Timeout:                  ServiceTimeout,
		PassThroughPanics:        true,
		DontPropagateTermination: false,
	}
}

func infoEventHook(l log.Logger) suture.EventHook {
	var prevTerminate suture.EventServiceTerminate
	return func(ei suture.Event) {
		switch e := ei.(type) {
		case suture.EventStopTimeout:
			l.Infof("%s: '%s' failed to terminate in time", e.SupervisorName, e.ServiceName)
		case suture.EventServicePanic:
			l.Warnf("%s: '%s' panic", e.SupervisorName, e.ServiceName)
			l.Warn(e)
		case suture.EventServiceTerminate:
			msg := fmt.Sprintf("%s:%s failed: %v", e.SupervisorName, e.ServiceName, e.Err)
			if e.ServiceName == prevTerminate.ServiceName && e.Err == prevTerminate.Err {
				l.Debug(msg)
			} else {
				l.Info(msg)
			}
			prevTerminate = e
			l.Debug(e) // Contains some backoff statistics
		case suture.EventBackoff:
			l.Debugf("%s: too many service failures - entering the backoff state.", e.SupervisorName)
		case suture.EventResume:
			l.Debugf("%s: resume - exiting the backoff state.", e.SupervisorName)
		default:
			l.Warn("Unknown suture supervisor event type", e.Type())
			l.Info(e)
		}
	}
}
