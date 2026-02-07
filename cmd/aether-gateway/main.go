package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
	"time"

	"aether-rea/internal/core"

	"github.com/quic-go/quic-go"
	"github.com/quic-go/quic-go/http3"
	"github.com/quic-go/webtransport-go"
)

var (
	listenAddr = flag.String("listen", ":4433", "Listen address")
	certFile   = flag.String("cert", "cert.pem", "TLS certificate file")
	keyFile    = flag.String("key", "key.pem", "TLS key file")
	psk        = flag.String("psk", "", "Pre-shared key")
	secretPath = flag.String("path", "/v1/api/sync", "Secret path for WebTransport")
)

func main() {
	flag.Parse()

	if *psk == "" {
		log.Fatal("PSK is required")
	}

	// Load TLS certs
	certs, err := tls.LoadX509KeyPair(*certFile, *keyFile)
	if err != nil {
		log.Fatalf("Failed to load certs: %v", err)
	}

	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{certs},
		NextProtos:   []string{http3.NextProtoH3},
	}

	quicConfig := &quic.Config{
		EnableDatagrams: true,
	}

	server := webtransport.Server{
		H3: http3.Server{
			Addr:       *listenAddr,
			TLSConfig:  tlsConfig,
			QUICConfig: quicConfig,
		},
		CheckOrigin: func(r *http.Request) bool { return true },
	}

	http.HandleFunc(*secretPath, func(w http.ResponseWriter, r *http.Request) {
		session, err := server.Upgrade(w, r)
		if err != nil {
			log.Printf("Upgrade failed: %v", err)
			w.WriteHeader(500)
			return
		}
		handleSession(session, *psk)
	})

	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.Write([]byte(`<!doctype html>
<html lang="en">
  <head>
    <meta charset="utf-8" />
    <title>Aether Edge Relay</title>
  </head>
  <body>
    <h1>Aether Edge Relay</h1>
    <p>Operational</p>
  </body>
</html>`))
	})

	log.Printf("Listening on %s", *listenAddr)
	if err := server.ListenAndServe(); err != nil {
		log.Fatal(err)
	}
}

func handleSession(session *webtransport.Session, psk string) {
	log.Println("New session established")
	
	// Create a stream counter
	var streamID uint64 = 0

	for {
		stream, err := session.AcceptStream(context.Background())
		if err != nil {
			log.Printf("AcceptStream failed: %v", err)
			break
		}
		streamID++
		go handleStream(stream, psk, streamID)
	}
}

func handleStream(stream webtransport.Stream, psk string, streamID uint64) {
	defer stream.Close()

	reader := core.NewRecordReader(stream)
	
	// Read Metadata
	record, err := reader.ReadNextRecord()
	if err != nil {
		log.Printf("[Stream %d] Failed to read metadata record: %v", streamID, err)
		writeError(stream, 0x0001, "metadata required")
		return
	}

	if record.Type != core.TypeMetadata {
		log.Printf("[Stream %d] Invalid record type: %d", streamID, record.Type)
		writeError(stream, 0x0001, "metadata required")
		return
	}

	meta, err := core.DecryptMetadata(record, psk, streamID)
	if err != nil {
		log.Printf("[Stream %d] Decrypt failed: %v", streamID, err)
		writeError(stream, 0x0002, "metadata decrypt failed")
		return
	}

	targetAddr := fmt.Sprintf("%s:%d", meta.Host, meta.Port)
	log.Printf("[Stream %d] Connecting to %s", streamID, targetAddr)

	conn, err := net.DialTimeout("tcp", targetAddr, 10*time.Second)
	if err != nil {
		log.Printf("[Stream %d] Connect failed: %v", streamID, err)
		writeError(stream, 0x0004, "connect failed")
		return
	}
	defer conn.Close()

	// Bidirectional pipe
	errCh := make(chan error, 2)

	// WebTransport -> TCP
	go func() {
		buf := make([]byte, 32*1024)
		for {
			n, err := reader.Read(buf)
			if n > 0 {
				if _, wErr := conn.Write(buf[:n]); wErr != nil {
					errCh <- wErr
					return
				}
			}
			if err != nil {
				if err != io.EOF {
					errCh <- err
				} else {
					errCh <- nil
				}
				return
			}
		}
	}()

	// TCP -> WebTransport
	go func() {
		buf := make([]byte, 32*1024)
		for {
			n, err := conn.Read(buf)
			if n > 0 {
				// Encrypt/Wrap in Data Record
				recordBytes, err := core.BuildDataRecord(buf[:n], meta.Options.MaxPadding)
				if err != nil {
					errCh <- err
					return
				}
				if _, wErr := stream.Write(recordBytes); wErr != nil {
					errCh <- wErr
					return
				}
			}
			if err != nil {
				if err != io.EOF {
					// Ignore "use of closed network connection" if caused by other side closing
					if !strings.Contains(err.Error(), "closed network connection") {
						errCh <- err
					} else {
						errCh <- nil
					}
				} else {
					errCh <- nil
				}
				return
			}
		}
	}()

	select {
	case err := <-errCh:
		if err != nil {
			log.Printf("[Stream %d] Stream error: %v", streamID, err)
		}
	}
	// Cleanup happens via defer stream.Close() and defer conn.Close()
}

func writeError(w io.Writer, code uint16, msg string) {
	record, _ := core.BuildErrorRecord(code, msg)
	w.Write(record)
}
