package counter

import (
	"errors"
	"sync" // 不知道能不能使用sync包，这里只用到了sync.WaitGroup功能，不然需要额外的空间记录生产者的状态，感觉不划算
	"time" // 实现心跳机制ticker
)

type value struct {
	name string
	num  int
}

type Counter struct {
	Count   map[value]int
	ch      chan value     // 信号通道
	closing chan struct{}  // 多生产者对单消费者 -- 需要一个closing状态保证数据处理
	wgSend  sync.WaitGroup // 统筹多生产者的完成情况
}

// 全局唯一，不可复制
func (*Counter) Lock()   {}
func (*Counter) Unlock() {}

func (counter *Counter) Init() {
	counter.Count = make(map[value]int, 3)
	counter.ch = make(chan value)
	counter.closing = make(chan struct{})
	counter.wgSend = sync.WaitGroup{}
	go counter.run()
}

// 消费者处理信号
func (counter *Counter) run() {
	for i := range counter.ch {
		counter.Count[i]++ // 避开map并发写
		counter.wgSend.Done()
	}
}

// 增加
var ErrorClosing error = errors.New("the Counter closed")

func (counter *Counter) Incr(key string, num int) error {
	// v := value{name: key, num: num}
	select {
	case <-counter.closing: // 如果处于关闭状态，直接返回，不阻塞
		return ErrorClosing
	default:
		counter.wgSend.Add(1)
		counter.ch <- value{name: key, num: num}
	}
	return nil
}

// 热重置
func (counter *Counter) Reset() {
	counter.stop()
	counter.Count = make(map[value]int)
	counter.closing = make(chan struct{})
}

// 第三方调用选择结束
func (counter *Counter) stop() {
	close(counter.closing) // 进入关闭状态
	counter.wgSend.Wait()  // 等待当前的生产者都被消费，防止线程堆积
	// chan ch最后被垃圾回收
}

var destory = make(chan struct{})

func (counter *Counter) Destory() {
	close(destory)
	time.Sleep(1 * time.Second) // 能用context的话好一点
	counter.stop()
}

// 周期执行函数并重置Counter
func (counter *Counter) Flush2broker(timewait time.Duration, f func()) {
	heartbeat := time.NewTicker(timewait)
	defer heartbeat.Stop()
	for {
		select {
		case <-heartbeat.C:
			f()
			counter.Reset()
		case <-destory:
			return
		case <-destory:
			return
		}
	}
}
