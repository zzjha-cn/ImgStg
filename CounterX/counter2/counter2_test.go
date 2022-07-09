package counter2

import (
	"fmt"
	"sync"
	"testing"
	"time"
)

func BenchmarkRun(b *testing.B) {
	for i := 0; i < b.N; i++ {
		foo(200)
	}
}

func foo(n int) {
	c := &Counter{}
	c.Init()

	wg := sync.WaitGroup{}

	go c.Flush2broker(1*time.Second, func() {
		fmt.Println("输出：")
		println(&(c.count))
		fmt.Println(c.count)
	})
	for i := 0; i < n; i++ {
		wg.Add(1)
		go func(i int) {
			defer wg.Done()
			c.Incr("b", i%3+1)
		}(i)
	}
	wg.Wait()
	c.Destory()
}
