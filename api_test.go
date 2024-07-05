package serve

import (
	"context"
	"fmt"
	"github.com/expgo/config"
	"reflect"
	"runtime"
	"testing"
	"time"
)

//@Log
//go:generate ag

func init() {
	config.DefaultFile("test.yml")
}

var errNum = 0

func Normal(ctx context.Context) error {
	logger.Info("normal text")
	return nil
}

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
	ide1 := AddFuncs("error1", Error)
	ide2 := AddFuncs("error2", Error)
	ide3 := AddFuncs("error3", Error)
	ide4 := AddFuncs("normal", Normal)

	time.Sleep(20 * time.Second)

	Remove(ide3)

	time.Sleep(20 * time.Second)

	RemoveAndWait(ide2, 3*time.Second)

	time.Sleep(20 * time.Second)

	RemoveAndWait(ide1, 3*time.Second)

	time.Sleep(3 * time.Second)

	RemoveAndWait(ide4, 3*time.Second)

	time.Sleep(3 * time.Second)

	_ = Down()
}

func TestServerPanic(t *testing.T) {
	AddFuncs("error", ErrorPanic)

	time.Sleep(1 * time.Minute)

	_ = Down()
}

func TestTimeout(t *testing.T) {
	AddFuncs("timeout", ErrorTimeoutRun)
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
	AddFuncs("abc_run", abc.Run)
	AddFuncs("abc_read_panic", abc.ReadPanic)
	AddFuncs("abc_timeout", abc.Timeout)

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
	AddFuncs("AddMethods", Error, ErrorPanic, ErrorTimeoutRun)

	time.Sleep(1 * time.Minute)

	_ = Down()
}

func TestAddMethods1(t *testing.T) {
	abc := &Abc{}

	AddFuncs("abc", abc.Serve, abc.Run, abc.Timeout, abc.ReadPanic)

	time.Sleep(1 * time.Minute)

	_ = Down()
}

func TestAddServeAndFuncs(t *testing.T) {
	abc := &Abc{}

	AddServe("abc serve", abc, abc.Run, abc.Timeout, abc.ReadPanic)

	time.Sleep(1 * time.Minute)

	_ = Down()
}
