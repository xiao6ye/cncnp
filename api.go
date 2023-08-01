// Package zcnp
// @Title  zcnp
// @Description  zcnp
// @Author  zxx1224@gmail.com  2021/8/11 3:44 下午
// @Update  zxx1224@gmail.com  2021/8/11 3:44 下午
package zcnp

import (
	"io"
	"net"
	"strings"
	"time"
)

// Protocol 协议接口
type Protocol interface {
	NewCodec(conn net.Conn) (Codec, error)
}

type ProtocolFunc func(conn net.Conn) (Codec, error)

func (pf ProtocolFunc) NewCodec(conn net.Conn) (Codec, error) {
	return pf(conn)
}

// Codec 编解码器接口
type Codec interface {
	Receive() (interface{}, error)
	Send(interface{}) error
	Close() error
	GetMsgID() (sendMsgID int)
	GetRemoteAddr() net.Addr
}

// ClearSendChan 清除发送通道接口
type ClearSendChan interface {
	ClearSendChan(<-chan interface{})
}

func Listen(network, address string, protocol Protocol, sendChanSize int, handler Handler) (*Server, error) {
	listener, err := net.Listen(network, address)
	if err != nil {
		return nil, err
	}
	return NewServer(listener, protocol, sendChanSize, handler), nil
}

func Dial(network, address string, protocol Protocol, sendChanSize int) (*Session, error) {
	conn, err := net.Dial(network, address)
	if err != nil {
		return nil, err
	}
	codec, err := protocol.NewCodec(conn)
	if err != nil {
		return nil, err
	}
	return NewSession(codec, sendChanSize), nil
}

func DialTimeout(network, address string, timeout time.Duration, protocol Protocol, sendChanSize int) (*Session, error) {
	conn, err := net.DialTimeout(network, address, timeout)
	if err != nil {
		return nil, err
	}
	codec, err := protocol.NewCodec(conn)
	if err != nil {
		return nil, err
	}
	return NewSession(codec, sendChanSize), nil
}

func Accept(listener net.Listener) (net.Conn, error) {
	var tempDelay time.Duration
	for {
		conn, err := listener.Accept()
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Temporary() {
				if tempDelay == 0 {
					tempDelay = 5 * time.Millisecond
				} else {
					tempDelay *= 2
				}
				if max := 1 * time.Second; tempDelay > max {
					tempDelay = max
				}
				time.Sleep(tempDelay)
				continue
			}
			if strings.Contains(err.Error(), "use of closed network connection") {
				return nil, io.EOF
			}
			return nil, err
		}
		return conn, nil
	}
}
