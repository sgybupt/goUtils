package fileMonitor

import (
	"fmt"
	"testing"
	"time"
)

func TestFull(t *testing.T) {
	var monitor Monitor
	monitor.InitConfig(Config{
		TickTime:      time.Millisecond * 50,
		ToleranceTime: time.Millisecond * 100,
		AimDir:        "/Users/su/tempDir",
		DirLevel:      0, // 初始化的时候 添加的监听等级. 如果在监听文件夹底下继续创建文件夹, 那么无论层级都会加入监听
	})
	msgChan := make(chan EventInter, 1024)

	go monitor.Run(msgChan)

	ticker := time.NewTicker(time.Second * 10)
	for ; ; {
		select {
		case ev := <-msgChan:
			fmt.Println(ev)
		case <-ticker.C:
			goto stop
		}
	}
stop:
	ticker.Stop()
	monitor.Stop()
}
