package serve

import (
	"context"
	"fmt"
	"github.com/expgo/config"
	"github.com/thejerf/suture/v4"
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
	time.Sleep(15 * time.Second)
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
