### 处理失败自动存档的

不太清楚自动存档是存在磁盘还是存在broker（redis）中，姑且认为是存进broker中吧，思路是一样的。

只要实现一个handlefunc，针对错误的处理。在代码运行中存在：

```go
// ......
var remain int64
            s := time.Now()
            err = it.handler.ProcessMsg(it.ctx, msg)
            if it.workerWorkIntervalFunc != nil {
                remain = it.workerWorkIntervalFunc().Milliseconds() - time.Since(s).Milliseconds()
            }

            if err != nil {
                it.handleFailedMsg(msg, err)
            } else {
                it.handleSuccessMsg(msg)
            }

            if err != context.DeadlineExceeded {
                if it.workerWorkIntervalFunc != nil && remain > 0 {
                    // TODO: we should move this msg into waiting queue before it take by next ProcessMsg call?
                    time.Sleep(time.Millisecond * time.Duration(remain))
                }
            }  

// ...
```

所以只需要实现Fail方法即可。

### 支持中断暂停服务

首先是消息状态：QueueStat-->waiting

然后调用NewKeyQueueWaiting，NewKeyQueuePaused。

主要的还是通过wiki中的信号机制，设置一个信号来控制暂停与否的。
