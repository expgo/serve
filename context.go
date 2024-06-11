package serve

import (
	"context"
	"fmt"
	"github.com/expgo/log"
	"github.com/thejerf/suture/v4"
	"time"
)

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
	spec := SpecWithDebugLogger(c.l)
	c.Supervisor = suture.New("serve", spec)

	c.ctx, c.cancel = context.WithCancel(context.Background())
	c.errChan = c.ServeBackground(c.ctx)
}

func (c *Context) Down() error {
	c.cancel()

	return <-c.errChan
}

func SpecWithDebugLogger(l log.Logger) suture.Spec {
	return spec(func(e suture.Event) {
		l.Debug(e)
	})
}

func SpecWithInfoLogger(l log.Logger) suture.Spec {
	return spec(infoEventHook(l))
}

func spec(eventHook suture.EventHook) suture.Spec {
	return suture.Spec{
		EventHook:                eventHook,
		Timeout:                  ServiceTimeout,
		PassThroughPanics:        true,
		DontPropagateTermination: false,
	}
}

// infoEventHook prints service failures and failures to stop services at level
// info. All other events and identical, consecutive failures are logged at
// debug only.
func infoEventHook(l log.Logger) suture.EventHook {
	var prevTerminate suture.EventServiceTerminate
	return func(ei suture.Event) {
		switch e := ei.(type) {
		case suture.EventStopTimeout:
			l.Infof("%s: Service %s failed to terminate in a timely manner", e.SupervisorName, e.ServiceName)
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
