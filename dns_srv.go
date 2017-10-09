package clientbuilder

import (
	"crypto/tls"
	"net"
	"net/url"
	"strconv"

	etcd "github.com/coreos/etcd/clientv3"
)

/*
FillConfigFromDNS queries DNS records for endpoints and adds them to the
specified etcd client configuration.
*/
func FillConfigFromDNS(conf *etcd.Config, srvtype, domain string) error {
	var srvs []*net.SRV
	var srv *net.SRV
	var u url.URL
	var err error

	if srvtype == "etcd-client-ssl" {
		u.Scheme = "https"
	} else {
		u.Scheme = "http"
	}

	_, srvs, err = net.LookupSRV(srvtype, "tcp", domain)
	if err != nil {
		return err
	}

	for _, srv = range srvs {
		var port string = strconv.FormatUint(uint64(srv.Port), 10)
		u.Host = net.JoinHostPort(srv.Target, port)
		conf.Endpoints = append(conf.Endpoints, u.String())
	}

	return nil
}

/*
NewFromDNS creates a new etcd client configured from a DNS domain name, with
an optional user name and password.
*/
func NewFromDNS(domain, user, pass string) (*etcd.Client, error) {
	var config etcd.Config = etcd.Config{
		Username: user,
		Password: pass,
	}
	var err error

	err = FillConfigFromDNS(&config, "etcd-client", domain)
	if err != nil {
		return nil, err
	}

	return etcd.New(config)
}

/*
NewTLSFromDNS creates a new etcd client configured from a DNS domain name, with
a TLS configuration and an optional user name and password.
*/
func NewTLSFromDNS(tc *tls.Config,
	domain, user, pass string) (*etcd.Client, error) {
	var config etcd.Config = etcd.Config{
		Username: user,
		Password: pass,
		TLS:      tc,
	}
	var err error

	err = FillConfigFromDNS(&config, "etcd-client-ssl", domain)
	if err != nil {
		return nil, err
	}

	return etcd.New(config)
}
