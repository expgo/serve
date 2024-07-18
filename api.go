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

func AddServe(serveName string, svr suture.Service, funcs ...func(ctx context.Context) error) suture.ServiceToken {
	return AddServeWithSpec(serveName, svr, suture.Spec{}, funcs...)
}

func AddServeWithSpec(serveName string, svr suture.Service, spec suture.Spec, funcs ...func(ctx context.Context) error) suture.ServiceToken {
	sup := suture.New(serveName, spec)
	sup.Add(svr)

	for _, f := range funcs {
		sup.Add(AsService(f, _getMethodName(f)))
	}

	return __Context().Add(sup)
}

func AddFuncs(serveName string, funcs ...func(ctx context.Context) error) suture.ServiceToken {
	return AddFuncsWithSpec(serveName, suture.Spec{}, funcs...)
}

func AddFuncsWithSpec(serveName string, spec suture.Spec, funcs ...func(ctx context.Context) error) suture.ServiceToken {
	if len(funcs) == 0 {
		panic("AddFuncs must with a func")
	}

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

func AddFuncsOnServe(id suture.ServiceToken, funcs ...func(ctx context.Context) error) {
	__Context().supervisorLock.RLock()
	sup, ok := __Context().supervisorMap[id]
	__Context().supervisorLock.RUnlock()

	if ok {
		for _, f := range funcs {
			sup.Add(AsService(f, _getMethodName(f)))
		}
	}
}

func Down() error {
	return __Context().Down()
}
