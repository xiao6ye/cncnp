// Package codec
// @Title  codec
// @Description  codec
// @Author  zxx1224@gmail.com  2022/1/7 4:54 下午
// @Update  zxx1224@gmail.com  2022/1/7 4:54 下午
package codec

//
//import (
//	"bufio"
//	"io"
//	"net"
//	"zcnp"
//)
//
//func Bufio(base zcnp.Protocol, readBuf, writeBuf int) zcnp.Protocol {
//	return &bufioProtocol{
//		base:     base,
//		readBuf:  readBuf,
//		writeBuf: writeBuf,
//	}
//}
//
//type bufioProtocol struct {
//	base     zcnp.Protocol
//	readBuf  int
//	writeBuf int
//}
//
//func (b *bufioProtocol) NewCodec(conn net.Conn) (cc zcnp.Codec, err error) {
//	codec := new(bufioCodec)
//
//	if b.writeBuf > 0 {
//		codec.stream.w = bufio.NewWriterSize(rw, b.writeBuf)
//		codec.stream.Writer = codec.stream.w
//	} else {
//		codec.stream.Writer = rw
//	}
//
//	if b.readBuf > 0 {
//		codec.stream.Reader = bufio.NewReaderSize(rw, b.readBuf)
//	} else {
//		codec.stream.Reader = rw
//	}
//
//	codec.stream.c, _ = rw.(io.Closer)
//
//	codec.base, err = b.base.NewCodec(&codec.stream)
//	if err != nil {
//		return
//	}
//	cc = codec
//	return
//}
//
//type bufioStream struct {
//	io.Reader
//	io.Writer
//	c io.Closer
//	w *bufio.Writer
//}
//
//func (s *bufioStream) Flush() error {
//	if s.w != nil {
//		return s.w.Flush()
//	}
//	return nil
//}
//
//func (s *bufioStream) close() error {
//	if s.c != nil {
//		return s.c.Close()
//	}
//	return nil
//}
//
//type bufioCodec struct {
//	base   zcnp.Codec
//	stream bufioStream
//}
//
//func (c *bufioCodec) Send(msg interface{}) error {
//	if err := c.base.Send(msg); err != nil {
//		return err
//	}
//	return c.stream.Flush()
//}
//
//func (c *bufioCodec) Receive() (interface{}, error) {
//	return c.base.Receive()
//}
//
//func (c *bufioCodec) Close() error {
//	err1 := c.base.Close()
//	err2 := c.stream.close()
//	if err1 != nil {
//		return err1
//	}
//	return err2
//}
