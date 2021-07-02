package fmonitor

import (
	"github.com/fsnotify/fsnotify"
	"time"
)

type Config struct {
	TickTime      time.Duration // 过期时间检查间隔
	ToleranceTime time.Duration // Write Create后无操作的最大容忍时间  建议2-3秒
	AimDir        string
	// 最深触及的层级. 0表示只监控当前文件夹. <0 表示 监控到最深
	// 初始化的时候 添加的监听等级. 如果在监听文件夹底下继续创建文件夹, 那么无论层级都会加入监听
	DirLevel int
}

type EventWithTimestamp struct {
	fsnotify.Event
	T time.Time
}

func (ewt *EventWithTimestamp) SetT(t time.Time) {
	ewt.T = t
}
