// Package zcnp
// @Title  zcnp
// @Description  zcnp
// @Author  zxx1224@gmail.com  2021/8/11 3:42 下午
// @Update  zxx1224@gmail.com  2021/8/11 3:42 下午
package zcnp

import (
	"context"
	"fmt"
	"net"
)

type Server struct {
	manager      *Manager
	channel      *Channel
	listener     net.Listener
	protocol     Protocol
	handler      Handler
	sendChanSize int
}

type Handler interface {
	HandleSession(*Session, context.Context)
}

// 是否实现 interface
var _ Handler = HandlerFunc(nil)

type HandlerFunc func(*Session, context.Context)

func (f HandlerFunc) HandleSession(session *Session, ctx context.Context) {
	f(session, ctx)
}

func NewServer(listener net.Listener, protocol Protocol, sendChanSize int, handler Handler) *Server {
	return &Server{
		manager:      NewManager(),
		channel:      NewChannel(),
		listener:     listener,
		protocol:     protocol,
		handler:      handler,
		sendChanSize: sendChanSize,
	}
}

func (server *Server) Listener() net.Listener {
	return server.listener
}

func (server *Server) Start() error {
	fmt.Printf("Start Listening : %v \n", server.listener.Addr())
	for {
		conn, err := Accept(server.listener)
		if err != nil {
			return err
		}

		go func() {
			codec, err := server.protocol.NewCodec(conn)
			if err != nil {
				fmt.Println("创建链接报错：", err)
				conn.Close()
				return
			}
			session := server.manager.NewSession(codec, server.sendChanSize)
			server.handler.HandleSession(session, session.Ctx)
		}()
	}
}

func (server *Server) GetManager() *Manager {
	return server.manager
}

func (server *Server) GetChannel() *Channel {
	return server.channel
}

func (server *Server) GetSession(sessionID uint64) *Session {
	return server.manager.GetSession(sessionID)
}

func (server *Server) Stop() {
	server.listener.Close()
	server.manager.Dispose()
}
