// Package zcnp
// @Title  zcnp
// @Description  zcnp
// @Author  zxx1224@gmail.com  2021/8/11 3:43 下午
// @Update  zxx1224@gmail.com  2021/8/11 3:43 下午
package zcnp

import (
	"context"
	"errors"
	"fmt"
	gopubsub "github.com/xiao6ye/go-pubsub"
	"log"
	"sync"
	"sync/atomic"
	"time"
	"zcnp/logger"
)

var SessionClosedError = errors.New("session Closed")
var SessionBlockedError = errors.New("session Blocked")

var globalSessionId uint64

type Session struct {
	sync.RWMutex
	id                 uint64
	codec              Codec // 编解码
	manager            *Manager
	channel            *Channel //增加连接的别名，便于转发消息
	sendChan           chan interface{}
	recvMutex          sync.Mutex
	Ctx                context.Context
	Cancel             context.CancelFunc
	closeFlag          int32
	closeMutex         sync.Mutex
	firstCloseCallback *closeCallback
	lastCloseCallback  *closeCallback
	Property           interface{}
	PropertyLock       sync.Mutex //保护当前property的锁
	State              interface{}
	Subscriber         *gopubsub.Subscriber
}

func NewSession(codec Codec, sendChanSize int) *Session {
	return newSession(nil, codec, sendChanSize)
}

// 创建会话
func newSession(manager *Manager, codec Codec, sendChanSize int) *Session {
	session := &Session{
		codec:    codec,
		manager:  manager,
		id:       atomic.AddUint64(&globalSessionId, 1),
		sendChan: make(chan interface{}, sendChanSize),
	}
	session.Ctx, session.Cancel = context.WithCancel(context.Background())
	go session.sendLoop()
	go session.Start()
	return session
}

func (session *Session) Start() {
	log.Printf("开启tcp新链接：%v ，sessionID: %v \n", session.Codec().GetRemoteAddr(), session.ID())
	select {
	case <-session.Ctx.Done():
		session.Close()
		log.Println("关闭tcp链接")
		return
	}
}

func (session *Session) ID() uint64 {
	return session.id
}

// IsClosed 是否关闭
func (session *Session) IsClosed() bool {
	return atomic.LoadInt32(&session.closeFlag) == 1
}

// Close 关闭会话
func (session *Session) Close() {
	log.Println("TCP链接 ConnID:", session.id, " 断开链接！")
	logger.SugarLogger.Info("TCP链接 ConnID:", session.id, " 断开链接！")
	session.Lock()
	defer session.Unlock()
	session.Cancel()
	if atomic.CompareAndSwapInt32(&session.closeFlag, 0, 1) {
		close(session.sendChan)
		//close(session.CmdMsg)
		if clear, ok := session.codec.(ClearSendChan); ok {
			clear.ClearSendChan(session.sendChan)
		}

		err := session.codec.Close()
		if err != nil {
			logger.SugarLogger.Error("链接关闭失败：", err)
		}

		session.invokeCloseCallbacks()
		if session.manager != nil {
			session.manager.delSession(session)
		}
	}
}

func (session *Session) GetChannel() *Channel {
	return session.channel
}

func (session *Session) Codec() Codec {
	return session.codec
}

func (session *Session) Receive() interface{} {
	session.recvMutex.Lock()
	defer session.recvMutex.Unlock()
	msg, err := session.codec.Receive()
	if err != nil {
		session.Cancel()
	}
	return msg
}

// 写队列
func (session *Session) sendLoop() {
	for {
		select {
		case <-session.Ctx.Done():
			fmt.Println("会话退出：", session.id)
			logger.SugarLogger.Warn("会话退出：", session.id)
			return
		case msg, ok := <-session.sendChan:
			if !ok || session.codec.Send(msg) != nil {
				return
			}
		}
	}
}

// Send 发送消息到消息队列
func (session *Session) Send(msg interface{}) error {
	session.RLock()
	defer session.RUnlock()
	//发送超时
	idleTimeout := time.NewTimer(5 * time.Millisecond)
	defer idleTimeout.Stop()

	if session.IsClosed() {
		return SessionClosedError
	}

	select {
	case <-idleTimeout.C:
		return errors.New("send buff msg timeout")
	case session.sendChan <- msg:
		return nil
	}
}

type closeCallback struct {
	Handler interface{}
	Key     interface{}
	Func    func()
	Next    *closeCallback
}

// AddCloseCallback 关闭链接回调
func (session *Session) AddCloseCallback(handler, key interface{}, callback func()) {
	if session.IsClosed() {
		return
	}

	session.closeMutex.Lock()
	defer session.closeMutex.Unlock()

	newItem := &closeCallback{handler, key, callback, nil}

	if session.firstCloseCallback == nil {
		session.firstCloseCallback = newItem
	} else {
		session.lastCloseCallback.Next = newItem
	}
	session.lastCloseCallback = newItem
}

func (session *Session) RemoveCloseCallback(handler, key interface{}) {
	if session.IsClosed() {
		return
	}

	session.closeMutex.Lock()
	defer session.closeMutex.Unlock()

	var prev *closeCallback
	for callback := session.firstCloseCallback; callback != nil; prev, callback = callback, callback.Next {
		if callback.Handler == handler && callback.Key == key {
			if session.firstCloseCallback == callback {
				session.firstCloseCallback = callback.Next
			} else {
				prev.Next = callback.Next
			}
			if session.lastCloseCallback == callback {
				session.lastCloseCallback = prev
			}
			return
		}
	}
}

func (session *Session) invokeCloseCallbacks() {
	session.closeMutex.Lock()
	defer session.closeMutex.Unlock()
	for callback := session.firstCloseCallback; callback != nil; callback = callback.Next {
		callback.Func()
	}
}
