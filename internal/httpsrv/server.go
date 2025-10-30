/*
 *  Copyright (c) 2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a GPL-3.0 license that can be found in the LICENSE file.
 */

package httpsrv

import (
	"encoding/base64"
	"fmt"
	"net"

	"go.osspkg.com/do"
	"go.osspkg.com/errors"
	"go.osspkg.com/logx"

	"go.arwos.org/pinion/internal/config"
)

type Server struct {
	conf *config.Config
	http *net.TCPListener
	dial *Dial

	auth map[string]struct{}
}

func New(conf *config.Config) *Server {
	return &Server{
		conf: conf,
		auth: do.Entries[config.Auth, string, struct{}](conf.Auth, func(auth config.Auth) (string, struct{}) {
			return "Basic " + base64.StdEncoding.EncodeToString([]byte(auth.Login+":"+auth.Passwd)), struct{}{}
		}),
		dial: NewDial(conf.DNS, conf.Reverse, conf.Domains),
	}
}

func (v *Server) Up() (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = errors.Wrap(
				err,
				fmt.Errorf("panic: %+v", e),
			)
		}
	}()

	v.http, err = Listen(v.conf.Server.HttpServer)
	go v.handler()
	logx.Info("Start http server", "addr", v.http.Addr().String())

	return
}

func (v *Server) Down() (err error) {
	defer func() {
		if e := recover(); e != nil {
			err = errors.Wrap(
				err,
				fmt.Errorf("panic: %+v", e),
			)
		}
	}()

	err = v.http.Close()
	return
}
