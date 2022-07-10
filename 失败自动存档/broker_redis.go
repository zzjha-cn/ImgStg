package gmq

import (
	"context"
)

// 出队：
// return &Msg{
// 	Payload:     []byte(payload),
// 	Id:          msgId,
// 	Queue:       queueName,
// 	State:       state,
// 	Created:     gstr.Atoi64(created),
// 	Processedat: gstr.Atoi64(processedat),
// }, nil

// scriptEnqueue enqueues a message.
//
// Input:
// KEYS[1] -> gmq:<queueName>:msg:<msgId>
// KEYS[2] -> gmq:<queueName>:pending
// --
// ARGV[1] -> <message payload>
// ARGV[2] -> "pending"
// ARGV[3] -> <current unix time in milliseconds>
// ARGV[4] -> <msgId>
//
// Output:
// Returns 0 if successfully enqueued
// Returns 1 if task ID already exists
// var scriptEnqueue = redis.NewScript(`
// if redis.call("EXISTS", KEYS[1]) == 1 then
// 	return 1
// end
// redis.call("HSET", KEYS[1],
//            "payload", ARGV[1],
//            "state",   ARGV[2],
//            "created", ARGV[3])
// redis.call("LPUSH", KEYS[2], ARGV[4])
// return 0
// `)

func (it *BrokerRedis) Fail(ctx context.Context, msg IMsg, errFail error) (err error) {

	payload := msg.GetPayload()
	msgId := msg.GetId()
	queueName := msg.GetQueue()

	// 针对失败的，要不要unique？ -失败记录需要统计，以及支持后续的重试策略，所以，应该不用unique。
	keys := []string{
		NewKeyMsgDetail(it.namespace, queueName, msgId),
		NewKeyQueueFailed(it.namespace, queueName),
	}
	args := []interface{}{
		payload,
		MsgStateFailed,
		it.clock.Now().UnixMilli(), // 用的是这个时间？
		msgId,
	}
	resI, err := scriptEnqueue.Run(ctx, it.cli, keys, args...).Result()

	if err != nil {
		return
	}
	// errors. 错误的处理应该形成一条链
	rt, ok := resI.(int64)
	if !ok {
		err = ErrInternal
		return
	}

	if rt == LuaReturnCodeError {
		err = ErrMsgIdConflict
		return
	}

	return
}
