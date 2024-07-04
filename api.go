package serve

import (
	"context"
	"fmt"
	"github.com/expgo/sync"
	"github.com/thejerf/suture/v4"
	"reflect"
	"runtime"
	"strings"
	"time"
)

type Service interface {
	suture.Service
	fmt.Stringer
	Error() error
}

func _getMethodName(method func(ctx context.Context) error) string {
	methodName := runtime.FuncForPC(reflect.ValueOf(method).Pointer()).Name()
	lastDotIndex := strings.LastIndex(methodName, ".")

	if lastDotIndex > 0 && lastDotIndex < len(methodName) {
		return strings.TrimSuffix(methodName[lastDotIndex+1:], "-fm")
	} else if len(methodName) > 0 {
		return methodName
	}

	return "err: unknown method name"
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

func AddServeFuncs(svr suture.Service, funcs []func(ctx context.Context) error, serveName string, spec suture.Spec) suture.ServiceToken {
	sup := suture.New(serveName, spec)
	sup.Add(svr)

	for _, f := range funcs {
		sup.Add(AsService(f, _getMethodName(f)))
	}

	return __Context().Add(sup)
}

func AddFunc(fn func(ctx context.Context) error, serveName string, spec suture.Spec) suture.ServiceToken {
	sup := suture.New(serveName, spec)

	sup.Add(AsService(fn, _getMethodName(fn)))

	return __Context().Add(sup)
}

func AddFuncs(funcs []func(ctx context.Context) error, serveName string, spec suture.Spec) suture.ServiceToken {
	sup := suture.New(serveName, spec)

	for _, f := range funcs {
		sup.Add(AsService(f, _getMethodName(f)))
	}

	return __Context().Add(sup)
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
