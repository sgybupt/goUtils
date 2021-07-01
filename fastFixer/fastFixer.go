// 一个高效的维修器, 如果状态OK, 那么所有处理都可以极快并行下去
// 如果状态不OK, 或者 有维修器在工作, 那么到来的线程都会等待, 直至维修器工作完成

package fastFixer

import (
	"fmt"
	"sync/atomic"
)

var fixingFlag int64
var boardcastChan chan bool

type Condition interface {
	OK() bool
}

func init() {
	boardcastChan = make(chan bool)
}

//func Do(c Condition, do func(), fixer func()) {
//	if !c.OK() || atomic.LoadInt64(&fixingFlag) == 1 {
//		doWG.Add(1)
//		if atomic.CompareAndSwapInt64(&fixingFlag, 0, 1) {
//			fixer()
//			atomic.StoreInt64(&fixingFlag, 0)
//			doWG.Done()
//		} else {
//			doWG.Done()
//			doWG.Wait()
//		}
//	}
//	do()
//}

func Do(bad bool, do func(), fixer func()) {
	if bad || atomic.LoadInt64(&fixingFlag) == 1 {
		if atomic.CompareAndSwapInt64(&fixingFlag, 0, 1) {
			fixer()
			do()
			atomic.StoreInt64(&fixingFlag, 0)
			close(boardcastChan)
			boardcastChan = make(chan bool)
			return // 此处  fixer结束以后 就立刻do 然后在清理之前 return. 防止经过make chan等操作以后 在底下的do()之前 又需要fix
		} else {
			fmt.Println("in")
			<-boardcastChan
			fmt.Println("out")
		}
	}
	do()
}
