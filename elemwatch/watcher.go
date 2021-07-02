package elemwatch

import (
	"fmt"
	"os"
	"sync"
	"time"
)

var debug bool

func SetDebug(d bool) {
	debug = d
}

type FilterConfig struct {
	LoopTime, TolerateTime time.Duration
}

type ElemFilter struct {
	loopTime     time.Duration // 每个文件的扫描时间
	tolerateTime time.Duration //
	record       sync.Map
	i            <-chan ElemInter
	oS, oC       chan<- ElemInter
	stopChan     chan bool
	wg           *sync.WaitGroup
}

func NewFileWatcher(eConfig FilterConfig) *ElemFilter {
	return &ElemFilter{
		loopTime:     eConfig.LoopTime,
		tolerateTime: eConfig.TolerateTime,
		record:       sync.Map{},
		stopChan:     make(chan bool),
		wg:           new(sync.WaitGroup),
	}
}

// return -1 if file not exist or it is a dir
func GetFileSize(fp string) int64 {
	fi, err := os.Stat(fp)
	if err != nil {
		if err == os.ErrNotExist {
			return -1
		} else {
			panic(err)
		}
	}
	if fi.IsDir() {
		return -1
	}
	return fi.Size()
}

// changeFunc 用token算出一个版本号 若版本未推进, 则认为stable
func (ff *ElemFilter) Run(i <-chan ElemInter, oS, oC chan<- ElemInter, changeFunc func(token string) int64) {
	if debug {
		fmt.Println("watcher running")
	}

	ff.i = i
	ff.oS = oS
	ff.oC = oC

	for {
		select {
		case in, ok := <-ff.i:
			if !ok {
				return
			}
			if debug {
				fmt.Println("an elem", in.GetToken())
			}
			preTimeInter, has := ff.record.LoadOrStore(in.GetToken(), time.Now())
			if has { // this file has been watched
				if debug {
					fmt.Println("under watched, pass", in.GetToken())
				}
				continue
			}
			ff.wg.Add(1)
			go func(inElem ElemInter) {
				defer ff.wg.Done()
				defer ff.record.Delete(inElem.GetToken())
				preVersion := changeFunc(inElem.GetToken())
				preTime := preTimeInter.(time.Time)
				for {
					newVersion := changeFunc(inElem.GetToken())
					newTime := time.Now()

					if newVersion == preVersion && newTime.Sub(preTime) >= ff.tolerateTime {
						if ff.oS != nil {
							if debug {
								fmt.Println("stable", inElem.GetToken())
							}
							ff.oS <- inElem
						}
						return
					}
					if newVersion != preVersion {
						if ff.oC != nil {
							if debug {
								fmt.Println("changed", inElem.GetToken())
							}
							ff.oC <- inElem
						}
						// refresh
						preVersion = newVersion
						preTime = newTime
					}
					time.Sleep(ff.loopTime)
				}
			}(in)

		case _, ok := <-ff.stopChan:
			if !ok {
				if debug {
					fmt.Println("watcher is stopping")
				}
				return
			}
		}
	}
}

func (ff *ElemFilter) Stop() {
	close(ff.stopChan)
	ff.wg.Wait()
}
