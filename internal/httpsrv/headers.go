/*
 *  Copyright (c) 2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a GPL-3.0 license that can be found in the LICENSE file.
 */

package httpsrv

import (
	"net"
	"net/http"

	"go.osspkg.com/logx"
)

const proxyAuthHeader = "Proxy-Authorization"

func (v *Server) getHostAndCheckAuth(conn *net.TCPConn) (string, bool) {
	req, err := ReadRequest(conn)
	if err != nil {
		logx.Error("Parse http headers", "remote-addr", conn.RemoteAddr().String(), "err", err)
		return "", false
	}
	defer req.Body.Close() //nolint:errcheck

	if req.Method != http.MethodConnect {
		logx.Error("Invalid request type", "remote-addr", conn.RemoteAddr().String(), "method", req.Method)
		return "", false
	}

	if v.conf.Server.AuthRequire {
		token := req.Header.Get(proxyAuthHeader)
		if len(token) == 0 {
			SendProxyAuthRequired(conn)
			return "", false
		}
		if _, ok := v.auth[token]; !ok {
			SendProxyAuthRequired(conn)
			return "", false
		}
	}

	SendProxyOK(conn)

	return req.Host, true
}
