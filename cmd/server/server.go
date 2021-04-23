package server

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io"
	"io/ioutil"
	"log"
	"net"
	"sync"

	"github.com/spf13/cobra"

	"github.com/feeltheajf/ztunnel/cmd/util"
	"github.com/feeltheajf/ztunnel/config"
)

var Cmd = &cobra.Command{
	Use:     "server",
	Aliases: []string{"s"},
	Short:   "Run server",
	Run:     util.Wrap(run),
}

// TODO defaults
var flags = struct {
	config string
}{}

func init() {
	Cmd.Flags().StringVarP(&flags.config, "config", "c", config.DefaultPath, "path to config file")
}

var cfg *config.Config

func run() error {
	log.SetFlags(log.Lshortfile)

	cer, err := tls.LoadX509KeyPair("testdata/pki/ca.crt", "testdata/pki/ca.key")
	if err != nil {
		return err
	}

	roots, _ := x509.SystemCertPool()
	if roots == nil {
		roots = x509.NewCertPool()
	}

	certs, err := ioutil.ReadFile("testdata/pki/ca.crt")
	if err != nil {
		return err
	}

	if ok := roots.AppendCertsFromPEM(certs); !ok {
		return errors.New("failed to append client certificate")
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{cer},
		ClientAuth:   tls.VerifyClientCertIfGiven, // TODO configurable per host
		ClientCAs:    roots,
		MinVersion:   tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		},
		PreferServerCipherSuites: true,
		VerifyConnection: func(conn tls.ConnectionState) error {
			// use listener.Accept with net.Conn instead?

			// TODO check if connection is truly aborted
			// TODO extract leaf from conn.VerifiedChains
			// println("req server:", conn.ServerName)
			// for _, crt := range conn.PeerCertificates {
			// 	println("req client:", crt.Subject.CommonName)
			// 	if crt.Subject.CommonName != "autotest" {
			// 		return errors.New("invalid subject")
			// 	}
			// }
			return nil
		},
	}

	cfg, err = config.Load(flags.config)
	if err != nil {
		return err
	}

	ln, err := tls.Listen("tcp", cfg.Address, tlsConfig)
	if err != nil {
		return err
	}
	defer ln.Close()

	println("starting server")
	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}

		go handleConnection(conn)
	}

	return nil
}

func handleConnection(conn net.Conn) {
	defer conn.Close()
	// TODO read max N bytes from connection without losing them
	// check for CONNECT, see ghostunnel for example as well

	// TODO implement conn writer

	tlsConn, ok := conn.(*tls.Conn)
	if !ok {
		return // errors.New("not a TLS connection")
	}

	// TODO set deadlines

	if err := tlsConn.Handshake(); err != nil {
		return // fmt.Errorf("handshake failed: %s", err)
	}

	cs := tlsConn.ConnectionState()
	println("req server:", cs.ServerName)

	for _, crt := range cs.PeerCertificates {
		// TODO peek leaf certificate
		println("req client:", crt.Subject.CommonName)
		if crt.Subject.CommonName != "autotest" {
			conn.Write([]byte("invalid subject"))
			continue
		}
	}

	// TODO move certificate validation here

	srv, ok := cfg.Servers[cs.ServerName]
	if !ok {
		return
	}

	// TODO need to (pre-)parse servers to extract port/scheme

	conn.Write([]byte("sni: " + cs.ServerName + "\n" + "srv: " + srv))

	// fmt.Println(conn.LocalAddr())
	// go func() {
	// 	defer conn.Close()
	// 	backend, err := tls.Dial("tcp", srv, config)
	// 	if err != nil {
	// 		fmt.Printf("%s", err)
	// 		return
	// 	}

	// 	fuse(conn, backend)
	// }()
}

func fuse(client, backend net.Conn) {
	// Copy from client -> backend, and from backend -> client
	// defer p.logConnectionMessage("closed", client, backend)
	// p.logConnectionMessage("opening", client, backend)

	wg := &sync.WaitGroup{}
	wg.Add(1)
	go func() { copyData(client, backend); wg.Done() }()
	copyData(backend, client)
	wg.Wait()
}

func copyData(dst net.Conn, src net.Conn) {
	defer dst.Close()
	defer src.Close()

	// _, err := io.Copy(dst, src)
	io.Copy(dst, src)

	// if err != nil && !isClosedConnectionError(err) {
	// 	// We don't log individual "read from closed connection" errors, because
	// 	// we already have a log statement showing that a pipe has been closed.
	// 	p.logConditional(LogConnectionErrors, "error during copy: %s", err)
	// }
}
