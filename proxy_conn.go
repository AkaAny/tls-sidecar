package tls_sidecar

import (
	"bufio"
	"net"
	"time"
)

type ProxyConn struct {
	readBuffer  *bufio.Reader
	writeBuffer *bufio.Writer
}

func (pc *ProxyConn) Read(b []byte) (n int, err error) {
	return pc.readBuffer.Read(b)
}

func (pc *ProxyConn) Write(b []byte) (n int, err error) {
	return pc.writeBuffer.Write(b)
}

func (pc *ProxyConn) Close() error {
	return nil
}

func (pc *ProxyConn) LocalAddr() net.Addr {
	//TODO implement me
	panic("implement me")
}

func (pc *ProxyConn) RemoteAddr() net.Addr {
	//TODO implement me
	panic("implement me")
}

func (pc *ProxyConn) SetDeadline(t time.Time) error {
	//TODO implement me
	panic("implement me")
}

func (pc *ProxyConn) SetReadDeadline(t time.Time) error {
	//TODO implement me
	panic("implement me")
}

func (pc *ProxyConn) SetWriteDeadline(t time.Time) error {
	//TODO implement me
	panic("implement me")
}
