package elemWatch

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

type FileFilter struct {
	loopTime     time.Duration // 每个文件的扫描时间
	tolerateTime time.Duration //
	record       sync.Map
	i            <-chan InFileInfoInter
	oS, oC       chan<- OutFileInfoInter
	stopChan     chan bool
	wg           *sync.WaitGroup
}

func NewFileWatcher(loopTime, tolerateTIme time.Duration) *FileFilter {
	return &FileFilter{
		loopTime:     loopTime,
		tolerateTime: tolerateTIme,
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
func (ff *FileFilter) Run(i <-chan InFileInfoInter, oS, oC chan<- OutFileInfoInter, changeFunc func(token string) int64) {
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
				fmt.Println("input an elem", in.GetToken())
			}
			preTimeInter, has := ff.record.LoadOrStore(in.GetToken(), time.Now())
			if has { // this file has been watched
				if debug {
					fmt.Println("this elem has being watched, pass", in.GetToken())
				}
				continue
			}
			ff.wg.Add(1)
			go func(token string) {
				defer ff.wg.Done()
				defer ff.record.Delete(token)
				preVersion := changeFunc(token)
				preTime := preTimeInter.(time.Time)
				for {
					newVersion := changeFunc(token)
					newTime := time.Now()

					if newVersion == preVersion && newTime.Sub(preTime) >= ff.tolerateTime {
						if ff.oS != nil {
							if debug {
								fmt.Println("stable elem", token)
							}
							ff.oS <- NewOutElemInfo(token, newVersion)
						}
						return
					}
					if newVersion != preVersion {
						if ff.oC != nil {
							if debug {
								fmt.Println("elem changed", token)
							}
							ff.oC <- NewOutElemInfo(token, newVersion)
						}
						// refresh
						preVersion = newVersion
						preTime = newTime
					}
					time.Sleep(ff.loopTime)
				}
			}(in.GetToken())

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

func (ff *FileFilter) Stop() {
	close(ff.stopChan)
	ff.wg.Wait()
}
