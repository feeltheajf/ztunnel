package client

import (
	"crypto/tls"
	"fmt"
	"io"
	"log"
	"net"
	"sync"

	"github.com/spf13/cobra"

	"github.com/feeltheajf/ztunnel/cmd/util"
)

var Cmd = &cobra.Command{
	Use:     "client",
	Aliases: []string{"c"},
	Short:   "Run client",
	Run:     util.Wrap(run),
}

// TODO defaults
var flags = struct {
	config string
}{}

func init() {
	Cmd.Flags().StringVarP(&flags.config, "config", "c", "", "path to config file")
}

func run() error {
	log.SetFlags(log.Lshortfile)

	// cfg, err := config.Load(flags.config)
	// if err != nil {
	// 	return err
	// }

	cer, err := tls.LoadX509KeyPair("testdata/pki/ca.crt", "testdata/pki/ca.key")
	if err != nil {
		return err
	}

	config := &tls.Config{
		Certificates: []tls.Certificate{cer},
		MinVersion:   tls.VersionTLS12,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
			tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305,
			tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305,
		},
		InsecureSkipVerify: true, // TODO remove
		ServerName:         "localhost",
	}

	ln, err := tls.Listen("tcp", "127.0.0.1:8443", config)
	if err != nil {
		return err
	}
	defer ln.Close()

	println("starting client")
	for {
		conn, err := ln.Accept()
		if err != nil {
			return err
		}
		// tcpConn, ok := conn.(*net.TCPConn)
		// if !ok {
		// 	return fmt.Errorf("not a TCP connection")
		// }

		// fmt.Println(tcpConn.)

		go func() {
			defer conn.Close()
			backend, err := tls.Dial("tcp", "localhost:443", config)
			if err != nil {
				fmt.Printf("%s", err)
				return
			}

			fuse(conn, backend)
		}()
	}
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
