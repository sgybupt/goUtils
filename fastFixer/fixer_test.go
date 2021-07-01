package fastFixer

import (
	"fmt"
	"math/rand"
	"sync"
	"testing"
	"time"
)

type testCond struct {
	C bool
}

func (t testCond) OK() bool {
	return t.C
}

func TestDo(t *testing.T) {
	wg := new(sync.WaitGroup)

	do := func() {
		fmt.Println("do something")
		time.Sleep(time.Duration(rand.Int63n(10)) * time.Millisecond)
	}

	fix := func() {
		fmt.Println("fixing")
		time.Sleep(time.Duration(rand.Int63n(10)) * time.Millisecond)
		fmt.Println("fixed")
	}

	for i := 0; i < 100; i++ {
		for j := 0; j < 100; j++ {
			wg.Add(1)
			go func() {
				defer wg.Done()
				good := true
				randI := rand.Int63n(100)
				if randI < 50 {
					good = false
				}
				Do(good, do, fix)
			}()
		}
		wg.Wait()
	}
}
