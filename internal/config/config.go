/*
 *  Copyright (c) 2024 Mikhail Knyazhev <markus621@yandex.com>. All rights reserved.
 *  Use of this source code is governed by a GPL-3.0 license that can be found in the LICENSE file.
 */

package config

type Server struct {
	HttpServer  string `yaml:"http-server"`
	AuthRequire bool   `yaml:"auth-require"`
}

type Auth struct {
	Login  string `yaml:"login"`
	Passwd string `yaml:"passwd"`
}

type Reverse struct {
	Tag     string `yaml:"tag"`
	Address string `yaml:"address"`
	Login   string `yaml:"login"`
	Passwd  string `yaml:"passwd"`
}

type Domain struct {
	Tag     string   `yaml:"tag"`
	Domains []string `yaml:"domain"`
}

type Config struct {
	Server  Server    `yaml:"server"`
	DNS     []string  `yaml:"dns"`
	Auth    []Auth    `yaml:"auth"`
	Reverse []Reverse `yaml:"reverse"`
	Domains []Domain  `yaml:"domains"`
}

func (c *Config) Default() {
	c.Server.HttpServer = "0.0.0.0:55555"
	c.DNS = append(c.DNS, "1.1.1.1", "1.0.0.1")
	c.Auth = append(c.Auth, Auth{
		Login:  "user",
		Passwd: "pwd",
	})
	c.Reverse = append(c.Reverse, Reverse{
		Tag:     "p1",
		Address: "1.1.1.1:1234",
		Login:   "p_user",
		Passwd:  "p_passwd",
	})
	c.Domains = append(c.Domains, Domain{
		Tag: "p1",
		Domains: []string{
			"google.com",
			"example.com",
		},
	})
}
