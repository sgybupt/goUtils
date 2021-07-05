package logger

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"log"
	"os"
	"path"
	"sync"
	"time"
	"unsafe"
)

var InfoLog *log.Logger
var ErrorLog *log.Logger
var WarningLog *log.Logger
var FatalLog *log.Logger

const DEBUG = true

func LogInit(logConfig LogConfig) {
	f, err := os.Open(logConfig.LogPath)
	if err != nil {
		if err == os.ErrNotExist {
			//err = nil
			err = os.MkdirAll(logConfig.LogPath, 0666)
			if err != nil {
				log.Fatal(err)
			}
		} else {
			log.Fatal(err)
		}
	}
	defer f.Close()
	fInfo, err := f.Stat()
	if err != nil {
		log.Fatal(err)
	}
	if !fInfo.IsDir() {
		log.Fatalln("[Error]: ", errors.New("log path is not a dir"))
	}
	L.InitConfig(LogConfig{
		LogPath: logConfig.LogPath,
	})
	InfoLog = log.New(io.MultiWriter(nil, os.Stdout), "", log.Ldate|log.Ltime|log.Lshortfile)
	WarningLog = log.New(io.MultiWriter(nil, os.Stdout), "", log.Ldate|log.Ltime|log.Lshortfile)
	ErrorLog = log.New(io.MultiWriter(nil, os.Stderr), "", log.Ldate|log.Ltime|log.Lshortfile)
	FatalLog = log.New(io.MultiWriter(nil, os.Stderr), "", log.Ldate|log.Ltime|log.Lshortfile)
	return
}

func LogStart() {
	var outFile *os.File
	var fileName string
	var startFlag bool
	wg := sync.WaitGroup{}
	wg.Add(1)
	go func() {
		for {
			newFileName := path.Join(L.LogPath, time.Now().Format("2006-01-02")+".log")
			if newFileName == fileName {
				time.Sleep(time.Second * 10)
				continue
			}
			fileName = newFileName
			outFileNew, err := os.OpenFile(fileName, os.O_CREATE|os.O_WRONLY|os.O_APPEND, 0666)
			if err != nil {
				_ = outFile.Close()
				log.Fatalln("failed to open error logger file:", err)
			}
			// 无缝切换输出目录
			if DEBUG {
				InfoLog = log.New(io.MultiWriter(outFileNew, os.Stdout), "", log.Ldate|log.Ltime|log.Lshortfile)
				WarningLog = log.New(io.MultiWriter(outFileNew, os.Stdout), "", log.Ldate|log.Ltime|log.Lshortfile)
				ErrorLog = log.New(io.MultiWriter(outFileNew, os.Stderr), "", log.Ldate|log.Ltime|log.Lshortfile)
				FatalLog = log.New(io.MultiWriter(outFileNew, os.Stderr), "", log.Ldate|log.Ltime|log.Lshortfile)
			} else {
				InfoLog = log.New(io.MultiWriter(outFileNew, nil), "", log.Ldate|log.Ltime|log.Lshortfile)
				WarningLog = log.New(io.MultiWriter(outFileNew, nil), "", log.Ldate|log.Ltime|log.Lshortfile)
				ErrorLog = log.New(io.MultiWriter(outFileNew, os.Stderr), "", log.Ldate|log.Ltime|log.Lshortfile)
				FatalLog = log.New(io.MultiWriter(outFileNew, os.Stderr), "", log.Ldate|log.Ltime|log.Lshortfile)
			}

			// 关闭之前的文件
			if outFile != nil {
				_ = outFile.Close()
				outFile = nil
			}
			outFile = outFileNew

			if !startFlag {
				wg.Done()
				startFlag = true
			}
		}
	}()
	wg.Wait()
}

func Info(s string) {
	InfoLog.Println("[INFO]:", s)
}

func Warning(w string) {
	WarningLog.Println("[WARNING]:", w)
}

func Error(e error) {
	ErrorLog.Println("[ERROR]:", e.Error())
}

func Fatal(e error) {
	FatalLog.Fatal("[FATAL]:", e.Error())
}

func BufInfos(c <-chan bool) io.Writer {
	var buff bytes.Buffer
	//bufWrite := bufio.NewWriter(&buff)
	go func() {
		_ = <-c
		b, err := ioutil.ReadAll(&buff)
		if err != nil {
			Error(err)
			return
		}
		str := fmt.Sprintf("stdout: \n %s", *(*string)(unsafe.Pointer(&b)))
		Info(str)
		Info("stdout finished")
	}()
	return &buff
}
