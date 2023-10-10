package transport

import (
	"net"
	"time"
)

type reqTimeoutConn struct {
	conn    net.Conn
	timeout time.Duration
}

func (c *reqTimeoutConn) Read(b []byte) (n int, err error) {
	err = c.SetReadDeadline(time.Now().Add(c.timeout))
	if err != nil {
		return 0, err
	}
	return c.conn.Read(b)
}

func (c *reqTimeoutConn) Write(b []byte) (n int, err error) {
	err = c.SetWriteDeadline(time.Now().Add(c.timeout))
	if err != nil {
		return 0, err
	}
	return c.conn.Write(b)
}

func (c *reqTimeoutConn) Close() error {
	return c.conn.Close()
}

func (c *reqTimeoutConn) LocalAddr() net.Addr {
	return c.conn.LocalAddr()
}

func (c *reqTimeoutConn) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

func (c *reqTimeoutConn) SetDeadline(t time.Time) error {
	return c.conn.SetDeadline(t)
}

func (c *reqTimeoutConn) SetReadDeadline(t time.Time) error {
	return c.conn.SetReadDeadline(t)
}

func (c *reqTimeoutConn) SetWriteDeadline(t time.Time) error {
	return c.conn.SetWriteDeadline(t)
}
