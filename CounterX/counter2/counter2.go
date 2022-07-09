package counter2

import (
	"sync"
	"sync/atomic"
	"time"
)

type value struct {
	name string
	num  int
}

type counter struct {
	v   value
	num int32
}

type Counter struct {
	index   map[value]int
	count   []counter
	Wg      sync.WaitGroup
	mu      sync.RWMutex
	destory chan struct{}
}

// 全局唯一，不可复制
func (*Counter) Lock()   {}
func (*Counter) Unlock() {}

func (c *Counter) Init() {
	c.index = make(map[value]int)
	c.count = make([]counter, 0, 10)
	c.Wg = sync.WaitGroup{}
	c.mu = sync.RWMutex{}
	c.destory = make(chan struct{})
}

func (c *Counter) Incr(name string, num int) {
	v := value{name: name, num: num}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.Wg.Add(1)

	if _, ok := c.index[v]; !ok {
		c.count = append(c.count, counter{v, 0})
		c.index[v] = len(c.count) - 1
	}
	atomic.AddInt32(&c.count[c.index[v]].num, 1)
	c.Wg.Done()
}

func (c *Counter) Reset() {
	c.Wg.Wait()
	c.index = make(map[value]int)
	c.count = make([]counter, 0, 10)
}

func (c *Counter) Destory() {
	c.Wg.Wait()
	close(c.destory)
}

// 周期执行函数并重置Counter
func (c *Counter) Flush2broker(timewait time.Duration, f func()) {
	heartbeat := time.NewTicker(timewait)
	defer heartbeat.Stop()
	for {
		select {
		case <-heartbeat.C:
			f()
			c.Reset()
		case <-c.destory:
			return
		case <-c.destory:
			return
		}
	}
}

// func main() {
// 	c := &Counter{}
// 	c.Init()

// 	go c.Flush2broker(1*time.Second, func() {
// 		fmt.Println("输出：")
// 		fmt.Println(c.count)
// 	})
// 	for i := 0; i < 20; i++ {
// 		go c.Incr("b", i%3+1)
// 	}
// 	c.Destory()
// 	fmt.Println(c.count)
// }
