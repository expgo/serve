package serve

import (
	"context"
	"fmt"
	"github.com/expgo/log"
	"github.com/thejerf/suture/v4"
	"time"
)

//go:generate ag

const ServiceTimeout = 10 * time.Second

// @Singleton(localGetter)
type Context struct {
	*suture.Supervisor
	l       log.Logger `new:""`
	cancel  context.CancelFunc
	ctx     context.Context
	errChan <-chan error
}

func (c *Context) Init() {
	c.Supervisor = suture.New("serve", spec(infoEventHook(c.l)))
	c.ctx, c.cancel = context.WithCancel(context.Background())
	c.errChan = c.ServeBackground(c.ctx)
}

func (c *Context) Down() error {
	c.cancel()
	return <-c.errChan
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
			l.Infof("%s: Service '%s' failed to terminate in time", e.SupervisorName, e.ServiceName)
		case suture.EventServicePanic:
			l.Warn("Caught a service panic, which shouldn't happen")
			l.Info(e)
		case suture.EventServiceTerminate:
			msg := fmt.Sprintf("%s: service %s failed: %s", e.SupervisorName, e.ServiceName, e.Err)
			if e.ServiceName == prevTerminate.ServiceName && e.Err == prevTerminate.Err {
				l.Debug(msg)
			} else {
				l.Info(msg)
			}
			prevTerminate = e
			l.Debug(e) // Contains some backoff statistics
		case suture.EventBackoff:
			l.Debugf("%s: exiting the backoff state.", e.SupervisorName)
		case suture.EventResume:
			l.Debugf("%s: too many service failures - entering the backoff state.", e.SupervisorName)
		default:
			l.Warn("Unknown suture supervisor event type", e.Type())
			l.Info(e)
		}
	}
}
