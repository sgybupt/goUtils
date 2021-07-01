package fileWatch

import (
	"fmt"
	"math/rand"
	"os"
	"sync"
	"testing"
	"time"
)

func TestFull(t *testing.T) {
	SetDebug(true)
	filePath := "./fileWatcherTest.awdawdascoad.txt"
	changeFile, err := os.OpenFile(filePath, os.O_WRONLY|os.O_CREATE|os.O_APPEND, 0666)
	if err != nil {
		panic(err)
	}
	defer func() {
		_ = os.Remove(filePath)
	}()

	defer changeFile.Close()

	watcher := NewFileWatcher(time.Millisecond*100, time.Millisecond*200)
	i := make(chan InFileInfoInter, 1024)
	oC := make(chan OutFileInfoInter, 1024)
	oS := make(chan OutFileInfoInter, 1024)

	go watcher.Run(i, oS, oC)

	var stopFlag bool
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		for ; !stopFlag; {
			_oc := <-oC
			fmt.Println("changed", _oc)
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		for ; !stopFlag; {
			_os := <-oS
			fmt.Println("stable", _os)
		}
		wg.Done()
	}()

	wg.Add(1)
	go func() {
		for ; !stopFlag; {
			i <- NewFileInfo(filePath)
			fmt.Println("input file")
			ms := rand.Int63n(500) + 1
			time.Sleep(time.Millisecond * time.Duration(ms))
		}
		wg.Done()
	}()

	for times := 0; times < 10; times++ {
		_, err := changeFile.WriteString("asdawdwe")
		fmt.Println("write something")
		if err != nil {
			panic(err)
		}
		err = changeFile.Sync()
		if err != nil {
			panic(err)
		}

		ms := rand.Int63n(500) + 200
		time.Sleep(time.Millisecond * time.Duration(ms))
	}
	stopFlag = true
	watcher.Stop()
	close(i)
	close(oC)
	close(oS)

	wg.Wait()
}
