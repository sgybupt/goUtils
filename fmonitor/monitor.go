package fmonitor

import (
	"fmt"
	"github.com/fsnotify/fsnotify"
	"io/ioutil"
	"log"
	"os"
	"sync"
	"sync/atomic"
	"time"
)

func isDir(path string) bool {
	s, err := os.Stat(path)
	if err != nil {
		return false
	}
	return s.IsDir()
}

type Monitor struct {
	config    Config
	closeChan chan struct{}
	//lock          sync.RWMutex
	wg        *sync.WaitGroup
	isRunning int64 // 0: stopped, 1: running
	watchDirs map[string]bool
}

func (m *Monitor) InitConfig(config Config) (err error) {
	m.config = config
	if m.closeChan == nil {
		m.closeChan = make(chan struct{})
	}
	if m.wg == nil {
		m.wg = new(sync.WaitGroup)
	}
	return
}

func (m Monitor) GetConfig() (c Config) {
	return m.config
}

func (m Monitor) GetHealth() int64 {
	return m.isRunning
}

func (m Monitor) GetWatchDirs() map[string]bool {
	return m.watchDirs
}

func (m *Monitor) Stop() {
	if !atomic.CompareAndSwapInt64(&m.isRunning, 1, 0) {
		return
	}
	atomic.StoreInt64(&m.isRunning, 1)
	m.closeChan <- struct{}{}
	m.wg.Wait()
	for k := range m.watchDirs { // empty the watchDirs
		delete(m.watchDirs, k)
	}
	atomic.StoreInt64(&m.isRunning, 0)
}

func (m *Monitor) Run(msgChan chan<- *EventWithTimestamp) {
	if !atomic.CompareAndSwapInt64(&m.isRunning, 0, 1) {
		return
	}
	defer func() {
		atomic.StoreInt64(&m.isRunning, 0)
	}()

	watcher, err := fsnotify.NewWatcher()
	if err != nil {
		return
	}
	defer func() {
		watcher.Close()
	}()

	// a new watchDirsMap
	m.watchDirs = make(map[string]bool, 16)

	err = watcher.Add(m.config.AimDir)
	if err != nil {
		fmt.Println("add dir fatal", err)
		log.Fatal(err)
		return
	}

	m.watchDirs[m.config.AimDir] = true

	//dirs := make([]string, 0, 32)
	//dirs = append(dirs, m.config.AimDir)

	var dirDepthCount int
	var dirDFS func(s string) error
	dirDFS = func(s string) error {
		files, err := ioutil.ReadDir(s) // filename only
		if err != nil {
			return err
		}
		pathSep := string(os.PathSeparator)
		for _, fi := range files {
			if fi.IsDir() { // 目录, 递归遍历
				subPath := s + pathSep + fi.Name()
				//fmt.Println(subPath)
				//fmt.Println(m.config.DirLevel, dirDepthCount)
				if m.config.DirLevel < 0 || dirDepthCount < m.config.DirLevel {
					dirDepthCount++
					err = watcher.Add(subPath)
					if err != nil {
						return err
					}
					m.watchDirs[subPath] = true
					err = dirDFS(subPath)
					if err != nil {
						return err
					}
					dirDepthCount--
				}

			}
		}
		return nil
	}

	err = dirDFS(m.config.AimDir)
	if err != nil {
		fmt.Println("add dir fatal", err)
		log.Fatal(err)
		return
	}

	fmt.Println(m.watchDirs)

	m.wg.Add(1)

	go func() {
		defer m.wg.Done()
		statusMap := make(map[string]*EventWithTimestamp)
		tick := time.NewTicker(m.config.TickTime)
		defer tick.Stop()
		for {
			select {
			case event, ok := <-watcher.Events:
				if !ok {
					return
				}
				//if _, ok := statusMap[event.Name]; ok {
				//	statusMap[event.Name].SetT(time.Now())
				//} else {
				//	statusMap[event.Name] = &EventWithTimestamp{
				//		Event: event,
				//		T:     time.Now(),
				//	}
				//}

				statusMap[event.Name] = &EventWithTimestamp{
					Event: event,
					T:     time.Now(),
				}

				if event.Op&fsnotify.Create == fsnotify.Create {
					if isDir(event.Name) {
						err = watcher.Add(event.Name)
						if err != nil {
							log.Println(err)
							return
						}
						m.watchDirs[event.Name] = true
					}
				}

				if event.Op&fsnotify.Write == fsnotify.Write {
				}

				if event.Op&fsnotify.Remove == fsnotify.Remove {
					if m.watchDirs[event.Name] {
						delete(m.watchDirs, event.Name)
						// here error happens because this file has been removed.
						// but it doesn't matter.
						_ = watcher.Remove(event.Name)
					}
				}

				if event.Op&fsnotify.Chmod == fsnotify.Chmod {
				}

			case <-tick.C:
				for k, v := range statusMap {
					if time.Now().Sub(v.T) >= m.config.ToleranceTime {
						fmt.Println("[Info]: 过期", v.Name, v.Op.String())
						msgChan <- v
						delete(statusMap, k)
					}
				}

			case err, ok := <-watcher.Errors:
				if !ok {
					return
				}
				log.Println("[Warning]: watcher err", err)

			case _, _ = <-m.closeChan:
				return
			}
		}

	}()

	m.wg.Wait()
}
