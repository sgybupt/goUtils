package fileWatch

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

//func (ff *FileFilter) SetInputChan(i <-chan InFileInfoInter) {
//	ff.i = i
//}
//
//func (ff *FileFilter) SetOutChangeChan(o chan<- OutFileInfoInter) {
//	ff.oC = o
//}
//
//func (ff *FileFilter) SetOutStableChan(o chan<- OutFileInfoInter) {
//	ff.oS = o
//}

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

func (ff *FileFilter) Run(i <-chan InFileInfoInter, oS, oC chan<- OutFileInfoInter) {
	if debug {
		fmt.Println("watcher running")
	}

	ff.i = i
	ff.oS = oS
	ff.oC = oC

	for ; ; {
		select {
		case in, ok := <-ff.i:
			if !ok {
				return
			}
			if debug {
				fmt.Println("input a file", in.GetFullPath())
			}
			ff.wg.Add(1)
			go func() {
				defer ff.wg.Done()
				oldSize := GetFileSize(in.GetFullPath())
				if debug {
					fmt.Println(oldSize, in.GetFullPath())
				}
				if oldSize < 0 {
					// not exist or is dir, double check
					if debug {
						fmt.Println("this file is something wrong", in.GetFullPath())
					}
					return
				}
				timeInter, has := ff.record.LoadOrStore(in.GetFullPath(), time.Now())
				if has { // this file has been watched
					if debug {
						fmt.Println("this file has being watched, pass", in.GetFullPath())
					}
					return
				}
				defer ff.record.Delete(in.GetFullPath())
				oldTime := timeInter.(time.Time)
				for ; ; {
					newSize := GetFileSize(in.GetFullPath())
					if newSize == oldSize && time.Now().Sub(oldTime) >= ff.tolerateTime {
						if ff.oS != nil {
							if debug {
								fmt.Println("stable file", in.GetFullPath())
							}
							ff.oS <- NewOutFileInfo(in.GetFullPath(), newSize)
						}
						return
					} else {
						if newSize != oldSize {
							// refresh
							oldTime = time.Now()
							oldSize = newSize
							if ff.oC != nil {
								if debug {
									fmt.Println("file size changed", in.GetFullPath())
								}
								ff.oC <- NewOutFileInfo(in.GetFullPath(), newSize)
							}
						}
					}
					time.Sleep(ff.loopTime)
				}
			}()

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
