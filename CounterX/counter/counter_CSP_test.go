package counter

import (
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestCounter_Copy(t *testing.T) {
	counter := Counter{}
	// println(counter)  // 这里应该要报编译错误,因为值传递会复制
	println(&counter)
}

func TestCounter_run(t *testing.T) {
	c := &Counter{}
	c.Init()
	assert.Equal(t, 0, len(c.Count))
	c.wgSend.Add(1)
	c.ch <- value{name: "a", num: 1} // 初始化完测试发送
	println("try send a value")
	time.Sleep(1 * time.Second)
	assert.Equal(t, 1, len(c.Count))
	// 通过则 run Init 通过

	// 测试Incr
	c.Incr("b", 1)
	c.Incr("b", 1)
	c.Incr("b", 2)

	time.Sleep(1 * time.Second)
	assert.Equal(t, 2, c.Count[value{"b", 1}])
	assert.Equal(t, 1, c.Count[value{"b", 2}])

	c.Incr("b", 2)
	c.Incr("b", 2)
	c.Incr("b", 2)

	// 测试Reset
	c.Reset()
	assert.Equal(t, 0, len(c.Count))
	close(c.closing)
}

func TestCounter(t *testing.T) {
	// func ExampleCounter() {
	c := &Counter{}
	c.Init()

	go c.Flush2broker(1*time.Second, func() {
		fmt.Println("输出：")
		println(&(c.Count))
		fmt.Println(c.Count)
	})

	for i := 0; i < 200; i++ {
		go c.Incr("b", i%3+1)
		time.Sleep(100 * time.Millisecond)
	}
	time.Sleep(4 * time.Second)
	c.Destory()
}

func BenchmarkRun(b *testing.B) {
	c := &Counter{}
	c.Init()

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		go c.Incr("b", i%3+1)
	}
	c.Destory()
}

// BenchmarkRun-8
//    1        1006903000 ns/op           65424 B/op        354 allocs/op
