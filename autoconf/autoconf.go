/*
Package autoconf provides functions for determining an automatic configuration
for etcd from program flags.
*/
package autoconf

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"flag"
	"io/ioutil"
	"sync"

	clientbuilder "github.com/caoimhechaos/go-etcd-clientbuilder"
	etcd "github.com/coreos/etcd/clientv3"
)

var endpoint = flag.String(
	"etcd-endpoint", "",
	"Explicitly set etcd endpoint to this value. Overrides --etcd-domain.")
var domain = flag.String(
	"etcd-domain", "", "Domain name to retrieve etcd configuration from")
var username = flag.String(
	"etcd-user", "", "User name to specify to etcd (not recommended)")
var password = flag.String(
	"etcd-password", "", "Password to specify to etcd (not recommended)")

var serverCAPath = flag.String(
	"etcd-root-ca", "", "Path to an etcd root CA certificate")
var clientCertFile = flag.String(
	"etcd-client-cert", "", "Path to a TLS client certificate")
var clientKeyFile = flag.String(
	"etcd-client-key", "", "Path to a TLS client key")
var useTLS = flag.Bool("etcd-use-tls", true,
	"Whether to use TLS for the etcd client")

var etcdClient *etcd.Client
var etcdOnce *sync.Once
var etcdOnceError error

/*
DefaultEtcdClient returns an etcd instance configured from the default etcd
flags. A new instance is constructed on demand.
*/
func DefaultEtcdClient() (*etcd.Client, error) {
	etcdOnce.Do(func() { etcdClient, etcdOnceError = buildDefaultEtcdClient() })
	return etcdClient, etcdOnceError
}

/*
buildDefaultEtcdClient builds a new etcd client from the default etcd flags.
*/
func buildDefaultEtcdClient() (*etcd.Client, error) {
	var tc = new(tls.Config)
	var err error

	tc.RootCAs, err = x509.SystemCertPool()
	if err != nil {
		return nil, err
	}

	if serverCAPath != nil && len(*serverCAPath) > 0 {
		var cert *x509.Certificate
		var certPEMBlock []byte
		var certDERBlock *pem.Block

		certPEMBlock, err = ioutil.ReadFile(*serverCAPath)
		if err != nil {
			return nil, err
		}

		certDERBlock, _ = pem.Decode(certPEMBlock)
		if certDERBlock == nil {
			return nil, errors.New(
				"Error decoding certificate " + *serverCAPath)
		}

		cert, err = x509.ParseCertificate(certDERBlock.Bytes)
		if err != nil {
			return nil, err
		}

		tc.RootCAs.AddCert(cert)
	}

	if clientCertFile != nil && len(*clientCertFile) > 0 &&
		clientKeyFile != nil && len(*clientKeyFile) > 0 {
		var cert tls.Certificate
		var err error

		cert, err = tls.LoadX509KeyPair(*clientCertFile, *clientKeyFile)
		if err != nil {
			return nil, err
		}

		tc.Certificates = append(tc.Certificates, cert)
	}

	if *useTLS {
		return clientbuilder.NewTLSFromDNS(tc, *domain, *username, *password)
	}
	return clientbuilder.NewFromDNS(*domain, *username, *password)
}
