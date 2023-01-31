package main

import (
	"github.com/golang/glog"
	"golang.org/x/net/context"
	"net/http"
	"os"
	"os/signal"
	"syscall"
)

func main() {
	webhookServer := &WebhookServer{
		server: &http.Server{
			Addr: ":8080",
			//TLSConfig: &tls.Config{Certificates: []tls.Certificate{pair}},
		},
	}
	// 注册handler
	mux := http.NewServeMux()
	mux.HandleFunc("/mutate", webhookServer.dispatch)
	mux.HandleFunc("/validate", webhookServer.dispatch)
	webhookServer.server.Handler = mux
	go func() {
		if error := webhookServer.server.ListenAndServe(); error != nil {
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
