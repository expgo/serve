package serve

import (
	"context"
	"fmt"
	"github.com/expgo/sync"
	"github.com/thejerf/suture/v4"
	"time"
)

type Service interface {
	suture.Service
	fmt.Stringer
	Error() error
}

// AsService wraps the given function to implement suture.Service. In addition
// it keeps track of the returned error and allows querying that error.
func AsService(fn func(ctx context.Context) error, creator string) Service {
	return &service{
		creator: creator,
		serve:   fn,
		mut:     sync.NewMutex(),
	}
}

func AddServe(svr suture.Service, serveName string, spec suture.Spec) suture.ServiceToken {
	sup := suture.New(serveName, spec)
	sup.Add(svr)
	return __Context().Add(sup)
}

func AddFunc(fn func(ctx context.Context) error, serveName string, spec suture.Spec) suture.ServiceToken {
	svr := AsService(fn, serveName)
	return AddServe(svr, serveName, spec)
}

func Remove(id suture.ServiceToken) error {
	return __Context().Remove(id)
}

func RemoveAndWait(id suture.ServiceToken, timeout time.Duration) error {
	return __Context().RemoveAndWait(id, timeout)
}

func Down() error {
	return __Context().Down()
}
