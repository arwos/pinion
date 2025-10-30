/*
 *  Copyright (c) 2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a GPL-3.0 license that can be found in the LICENSE file.
 */

package httpsrv

import (
	"fmt"
	"net"
	"net/http"

	"github.com/miekg/dns"
	"go.osspkg.com/do"
	"go.osspkg.com/errors"
	"go.osspkg.com/goppy/v2/xdns"
	"go.osspkg.com/logx"

	"go.arwos.org/pinion/internal/config"
)

type Dial struct {
	dns     xdns.HandlerDNS
	revs    map[string]config.Reverse
	domains map[string]string
}

func NewDial(dns []string, rev []config.Reverse, domains []config.Domain) *Dial {
	d := &Dial{
		dns: xdns.DefaultExchanger(dns...),
		revs: do.Entries[config.Reverse, string, config.Reverse](rev, func(r config.Reverse) (string, config.Reverse) {
			return r.Tag, r
		}),
		domains: make(map[string]string, 1000),
	}

	for _, item := range domains {
		for _, domain := range item.Domains {
			d.domains[domain] = item.Tag
		}
	}

	return d
}

func (v *Dial) resolve(host, port string) (*net.TCPAddr, error) {
	rrs, err := v.dns.Exchange(dns.Question{
		Name:   host + ".",
		Qtype:  dns.TypeA,
		Qclass: dns.ClassINET,
	})
	if err != nil {
		return nil, err
	}

	var ip net.IP
	for _, rr := range rrs {
		vv, ok := rr.(*dns.A)
		if !ok {
			continue
		}
		ip = vv.A
		break
	}

	if ip == nil ||
		len(ip) == 0 ||
		ip.IsLoopback() ||
		ip.IsPrivate() ||
		ip.IsUnspecified() ||
		ip.IsInterfaceLocalMulticast() ||
		ip.IsLinkLocalMulticast() ||
		ip.IsLinkLocalUnicast() ||
		ip.IsMulticast() {
		return nil, fmt.Errorf("fail resolve domain: %s", host)
	}

	tcpAddr, err := net.ResolveTCPAddr("tcp", ip.String()+":"+port)
	if err != nil {
		return nil, err
	}

	return tcpAddr, nil
}

func (v *Dial) direct(host, port string) (*net.TCPConn, error) {
	tcpAddr, err := v.resolve(host, port)
	if err != nil {
		logx.Error("Resolve http domain address", "err", err, "host", host)
		return nil, err
	}

	cli, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		logx.Error("Dial domain address", "err", err, "host", host)
		return nil, err
	}

	return cli, nil
}

func (v *Dial) reverse(address string, c config.Reverse) (*net.TCPConn, error) {
	tcpAddr, err := net.ResolveTCPAddr("tcp", c.Address)
	if err != nil {
		return nil, err
	}

	cli, err := net.DialTCP("tcp", nil, tcpAddr)
	if err != nil {
		logx.Error("Dial reverse proxy", "err", err, "host", c.Address)
		return nil, err
	}

	SendProxyAuth(cli, address, c.Login, c.Passwd)

	resp, err := ReadResponse(cli)
	if err != nil || resp.StatusCode != http.StatusOK {
		if err == nil {
			resp.Body.Close() //nolint:errcheck
		}
		return nil, errors.Wrap(cli.Close(), err, fmt.Errorf("proxy `%s` status code: %d", c.Address, resp.StatusCode))
	}

	return cli, nil
}

func (v *Dial) Client(address string) (*net.TCPConn, error) {
	logx.Info("Dial", "addr", address)

	host, port, err := net.SplitHostPort(address)
	if err != nil {
		return nil, err
	}

	if tag, ok := v.domains[host]; ok {
		if c, ok := v.revs[tag]; ok {
			return v.reverse(address, c)
		}
	}

	return v.direct(host, port)
}
