/*
 *  Copyright (c) 2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a GPL-3.0 license that can be found in the LICENSE file.
 */

package httpsrv

import (
	"bufio"
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net"
	"net/http"
	"strings"

	"go.osspkg.com/ioutils"
	"go.osspkg.com/ioutils/pool"
	"go.osspkg.com/logx"
)

var poolBuff = pool.New[*bytes.Buffer](func() *bytes.Buffer {
	return bytes.NewBuffer(make([]byte, 0, 1024))
})

func Listen(addr string) (*net.TCPListener, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", addr)
	if err != nil {
		logx.Error("Resolve http server address", "err", err, "addr", addr)
		return nil, err
	}

	tcp, err := net.ListenTCP("tcp", tcpAddr)
	if err != nil {
		logx.Error("Listen http server", "err", err, "addr", addr)
		return nil, err
	}

	return tcp, nil
}

func IsNormalErr(err error) bool {
	if err == nil ||
		strings.Contains(err.Error(), "use of closed network connection") ||
		strings.Contains(err.Error(), "connection reset by peer") {
		return true
	}
	return false
}

func SendProxyOK(w io.Writer) {
	fmt.Fprintf(w, "HTTP/1.0 200 Connection established\r\n") //nolint:errcheck
	fmt.Fprintf(w, "\r\n")                                    //nolint:errcheck
}

func SendProxyAuthRequired(w io.Writer) {
	fmt.Fprintf(w, "HTTP/1.0 407 Proxy Authentication Required\r\n") //nolint:errcheck
	fmt.Fprintf(w, "Proxy-Authenticate: Basic realm=\"proxy\"\r\n")  //nolint:errcheck
	fmt.Fprintf(w, "Proxy-Connection: close\r\n")                    //nolint:errcheck
	fmt.Fprintf(w, "Content-type: text/html; charset=utf-8\r\n")     //nolint:errcheck
	fmt.Fprintf(w, "\r\n")                                           //nolint:errcheck
}

func SendProxyAuth(w io.Writer, address, login, passwd string) {
	fmt.Fprintf(w, "CONNECT %s HTTP/1.0\r\n", address)                                                               //nolint:errcheck
	fmt.Fprintf(w, "Host: %s\r\n", address)                                                                          //nolint:errcheck
	fmt.Fprintf(w, "Proxy-Authorization: Basic %s\r\n", base64.StdEncoding.EncodeToString([]byte(login+":"+passwd))) //nolint:errcheck
	fmt.Fprintf(w, "Proxy-Connection: Keep-Alive\r\n")                                                               //nolint:errcheck
	fmt.Fprintf(w, "\r\n")                                                                                           //nolint:errcheck
}

func ReadRequest(r io.Reader) (*http.Request, error) {
	buf := poolBuff.Get()
	defer func() {
		poolBuff.Put(buf)
	}()

	n, err := ioutils.Copy(buf, r)
	if err != nil || n == 0 {
		if err == nil {
			err = fmt.Errorf("got zero bytes headers")
		}
		return nil, err
	}

	br := bufio.NewReader(buf)
	req, err := http.ReadRequest(br)
	if err != nil {
		return nil, err
	}

	return req, nil
}

func ReadResponse(r io.Reader) (*http.Response, error) {
	buf := poolBuff.Get()
	defer func() {
		poolBuff.Put(buf)
	}()

	n, err := ioutils.Copy(buf, r)
	if err != nil || n == 0 {
		if err == nil {
			err = fmt.Errorf("got zero bytes headers")
		}
		return nil, err
	}

	br := bufio.NewReader(buf)
	req, err := http.ReadResponse(br, nil)
	if err != nil {
		return nil, err
	}

	return req, nil
}
