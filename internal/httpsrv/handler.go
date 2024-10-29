/*
 *  Copyright (c) 2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a GPL-3.0 license that can be found in the LICENSE file.
 */

package httpsrv

import (
	"context"
	"io"
	"net"

	"go.osspkg.com/logx"
)

func (v *Server) handler() {
	for {
		conn, err := v.http.AcceptTCP()
		if err != nil {
			if !IsNormalErr(err) {
				logx.Error("Failed to accept connection", "err", err)
			}
			return
		}

		if err = conn.SetNoDelay(true); err != nil {
			logx.Error("Failed to set no delay", "err", err)
		}

		logx.Info("Accept connect", "remote-addr", conn.RemoteAddr().String())

		go v.runPipe(conn)
	}
}

func (v *Server) runPipe(conn *net.TCPConn) {
	defer func() {
		if err := conn.Close(); !IsNormalErr(err) {
			logx.Error("Close connect", "remote-addr", conn.RemoteAddr().String(), "err", err)
		} else {
			logx.Info("Close connect", "remote-addr", conn.RemoteAddr().String())
		}
	}()

	host, ok := v.getHostAndCheckAuth(conn)
	if !ok {
		return
	}

	cli, err := v.dial.Client(host)
	if err != nil {
		logx.Error("Dial host", "host", host, "remote-addr", conn.RemoteAddr().String(), "err", err)
		return
	}
	defer func() {
		if err := cli.Close(); !IsNormalErr(err) {
			logx.Error("Close dial", "remote-addr", conn.RemoteAddr().String(), "host", host, "err", err)
		} else {
			logx.Error("Close dial", "remote-addr", conn.RemoteAddr().String(), "host", host)
		}
	}()

	ctx, cncl := context.WithCancel(context.Background())

	go func() {
		defer cncl()

		for {
			n, err := io.Copy(cli, conn)
			if !IsNormalErr(err) {
				logx.Error("Receive data", "host", host, "remote-addr", conn.RemoteAddr().String(), "err", err)
				return
			}
			if n == 0 {
				return
			}
		}
	}()

	go func() {
		defer cncl()

		for {
			n, err := io.Copy(conn, cli)
			if !IsNormalErr(err) {
				logx.Error("Send data", "host", host, "remote-addr", conn.RemoteAddr().String(), "err", err)
				return
			}
			if n == 0 {
				return
			}
		}
	}()

	<-ctx.Done()
	return
}
