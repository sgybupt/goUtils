package elemwatch

import (
	"fmt"
	"log"
	"math"
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
		log.Println("watcher running")
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
				log.Println("an elem", in.GetToken())
			}
			_, has := ff.record.LoadOrStore(in.GetToken(), in)
			if has { // this file has been watched
				ff.record.Store(in.GetToken(), in) // 刷新elem的信息, token是一样的
				continue
			}
			ff.wg.Add(1)
			go func(token string) {
				defer ff.wg.Done()
				defer ff.record.Delete(token)

				if debug {
					defer func() {
						fmt.Println(fmt.Sprintf("token %s 退出", token))
					}()
				}
				var preVersion int64 = math.MinInt64
				var preTime = time.Now()
				for {
					time.Sleep(ff.loopTime)
					newVersion := changeFunc(token)
					newTime := time.Now()

					if newVersion != preVersion {
						if debug {
							log.Println(fmt.Sprintf("%s 版本更替: %d -> %d", token, preVersion, newVersion))
						}
						if ff.oC != nil {
							inElem, ok := ff.record.Load(token)
							if ok {
								ff.oC <- inElem.(ElemInter)
							}
						}
						preTime = newTime
						preVersion = newVersion
						continue
					}

					if newVersion == preVersion && newTime.Sub(preTime) >= ff.tolerateTime {
						if ff.oS != nil {
							if debug {
								log.Println("stable", token)
							}
							inElem, ok := ff.record.Load(token)
							if ok {
								ff.oS <- inElem.(ElemInter)
							}
						}
						return
					}
				}
			}(in.GetToken())

		case _, ok := <-ff.stopChan:
			if !ok {
				if debug {
					log.Println("watcher is stopping")
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
