package main

import (
	"crypto/tls"
	"flag"
	"fmt"
	"github.com/golang/glog"
	"golang.org/x/net/context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

type ServerParameters struct {
	port     string
	certFile string
	keyFile  string
}

func main() {
	var parameters ServerParameters
	flag.StringVar(&parameters.port, "port", "443", "server port")
	flag.StringVar(&parameters.keyFile, "tlsKeyFile", "/etc/webhook/certs/key.pem", "File containing the x509 private key to --tlsCertFile.")
	flag.StringVar(&parameters.certFile, "tlsCertFile", "/etc/webhook/certs/cert.pem", "File containing the x509 Certificate for HTTPS.")

	pair, err := tls.LoadX509KeyPair(parameters.certFile, parameters.keyFile)
	if err != nil {
		glog.Errorf("Failed to load key pair: %v", err)
	}

	webhookServer := &WebhookServer{
		server: &http.Server{
			Addr:      fmt.Sprintf(":%v", parameters.port),
			TLSConfig: &tls.Config{Certificates: []tls.Certificate{pair}},
		},
	}
	// 注册handler
	mux := http.NewServeMux()
	mux.HandleFunc("/mutate", webhookServer.dispatch)
	mux.HandleFunc("/validate", webhookServer.dispatch)
	webhookServer.server.Handler = mux
	go func() {
		if error := webhookServer.server.ListenAndServeTLS(parameters.certFile, parameters.keyFile); error != nil {
			glog.Error("Failed to listen and serve webhook server: %v", error)
		}
	}()
	glog.Info("Server started")

	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	glog.Infof("Got OS shutdown signal, shutting down webhook server gracefully...")
	webhookServer.server.Shutdown(context.Background())
}
