package serve

import (
	"context"
	"fmt"
	"github.com/expgo/config"
	"github.com/thejerf/suture/v4"
	"reflect"
	"runtime"
	"testing"
	"time"
)

func init() {
	config.DefaultFile("test.yml")
}

var errNum = 0

func Error(ctx context.Context) error {
	errNum += 1
	time.Sleep(2 * time.Second)
	return fmt.Errorf("error: %d", errNum)
}

func ErrorPanic(ctx context.Context) error {
	panic(Error(ctx))
}

func ErrorTimeoutRun(ctx context.Context) error {
	time.Sleep(5 * time.Second)
	panic(Error(ctx))
}

func TestServer(t *testing.T) {
	ide1 := AddFunc(Error, "error1", suture.Spec{})
	ide2 := AddFunc(Error, "error2", suture.Spec{})
	ide3 := AddFunc(Error, "error3", suture.Spec{})

	time.Sleep(20 * time.Second)

	Remove(ide3)

	time.Sleep(20 * time.Second)

	RemoveAndWait(ide2, 3*time.Second)

	time.Sleep(20 * time.Second)

	RemoveAndWait(ide1, 3*time.Second)

	time.Sleep(3 * time.Second)

	_ = Down()
}

func TestServerPanic(t *testing.T) {
	AddFunc(ErrorPanic, "error", suture.Spec{})

	time.Sleep(1 * time.Minute)

	_ = Down()
}

func TestTimeout(t *testing.T) {
	AddFunc(ErrorTimeoutRun, "timeout", suture.Spec{})
	_ = Down()
}

type Abc struct {
}

func (a *Abc) Serve(ctx context.Context) error {
	return Error(ctx)
}

func (a *Abc) Run(ctx context.Context) error {
	return Error(ctx)
}

func (a *Abc) ReadPanic(ctx context.Context) error {
	return ErrorPanic(ctx)
}

func (a *Abc) Timeout(ctx context.Context) error {
	return ErrorTimeoutRun(ctx)
}

func TestObjectServe(t *testing.T) {
	abc := &Abc{}
	AddFunc(abc.Run, "abc_run", suture.Spec{})
	AddFunc(abc.ReadPanic, "abc_read_panic", suture.Spec{})
	AddFunc(abc.Timeout, "abc_timeout", suture.Spec{})

	time.Sleep(1 * time.Minute)

	_ = Down()
}

func testFunc1(ctx context.Context) error {
	return nil
}

func testFunc2(ctx context.Context) error {
	return nil
}

func TestForMethodName(t *testing.T) {
	methods := []func(context.Context) error{testFunc1, testFunc2}

	for i, method := range methods {
		fmt.Printf("方法 #%d: %s\n", i, runtime.FuncForPC(reflect.ValueOf(method).Pointer()).Name())
	}
}

func TestAddMethods(t *testing.T) {
	AddFuncs([]func(context.Context) error{Error, ErrorPanic, ErrorTimeoutRun}, "AddMethods", suture.Spec{})

	time.Sleep(1 * time.Minute)

	_ = Down()
}

func TestAddMethods1(t *testing.T) {
	abc := &Abc{}

	AddFuncs([]func(context.Context) error{abc.Serve, abc.Run, abc.Timeout, abc.ReadPanic}, "abc", suture.Spec{})

	time.Sleep(1 * time.Minute)

	_ = Down()
}

func TestAddServeAndFuncs(t *testing.T) {
	abc := &Abc{}

	AddServeFuncs(abc, []func(context.Context) error{abc.Run, abc.Timeout, abc.ReadPanic}, "abc serve", suture.Spec{})

	time.Sleep(1 * time.Minute)

	_ = Down()
}
