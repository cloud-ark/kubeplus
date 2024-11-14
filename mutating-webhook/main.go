package main

import (
	"context"
	"crypto/tls"
	"flag"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
)

var (
	kubeClient *kubernetes.Clientset
)

func main() {
	var parameters WhSvrParameters
	// get command line parameters
	flag.IntVar(&parameters.port, "port", 443, "Webhook server port.")
	flag.BoolVar(&parameters.alsoLogToStderr, "alsologtostderr", true, "Flag that controls sending logs to stderr")
	flag.StringVar(&parameters.certFile, "tlsCertFile", "/etc/webhook/certs/cert.pem", "File containing the x509 Certificate for HTTPS.")
	flag.StringVar(&parameters.keyFile, "tlsKeyFile", "/etc/webhook/certs/key.pem", "File containing the x509 private key to --tlsCertFile.")
	flag.Parse()

	// creates the in-cluster config
	cfg, err := rest.InClusterConfig()
	if err != nil {
		panic(err)
	}
	kubeClient = kubernetes.NewForConfigOrDie(cfg)

	// fmt.Println(kubeClient.CoreV1().Pods("Default").Get("cluster1-mysql-0", metav1.GetOptions{}))

	pair, err := tls.LoadX509KeyPair(parameters.certFile, parameters.keyFile)
	if err != nil {
		panic(fmt.Errorf("Failed to load key pair: %v", err))
	}
	whsvr := &WebhookServer{
		server: &http.Server{
			Addr:      fmt.Sprintf(":%v", parameters.port),
			TLSConfig: &tls.Config{MinVersion: tls.VersionTLS12,
					       MaxVersion: tls.VersionTLS12,
					       Certificates: []tls.Certificate{pair},
			                       CipherSuites: []uint16{
								// TLS 1.0 - 1.2 cipher suites.
								tls.TLS_RSA_WITH_RC4_128_SHA,
								tls.TLS_RSA_WITH_3DES_EDE_CBC_SHA,
								tls.TLS_RSA_WITH_AES_128_CBC_SHA,
								tls.TLS_RSA_WITH_AES_256_CBC_SHA,
								tls.TLS_RSA_WITH_AES_128_CBC_SHA256,
								tls.TLS_RSA_WITH_AES_128_GCM_SHA256,
								tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
								tls.TLS_ECDHE_ECDSA_WITH_RC4_128_SHA,
								tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA,
								tls.TLS_ECDHE_ECDSA_WITH_AES_256_CBC_SHA,
								tls.TLS_ECDHE_RSA_WITH_RC4_128_SHA,
								tls.TLS_ECDHE_RSA_WITH_3DES_EDE_CBC_SHA,
								tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA,
								tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
								tls.TLS_ECDHE_ECDSA_WITH_AES_128_CBC_SHA256,
								tls.TLS_ECDHE_RSA_WITH_AES_128_CBC_SHA256,
								tls.TLS_ECDHE_RSA_WITH_AES_128_GCM_SHA256,
								tls.TLS_ECDHE_ECDSA_WITH_AES_128_GCM_SHA256,
								tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
								tls.TLS_ECDHE_ECDSA_WITH_AES_256_GCM_SHA384,
								tls.TLS_ECDHE_RSA_WITH_CHACHA20_POLY1305_SHA256,
								tls.TLS_ECDHE_ECDSA_WITH_CHACHA20_POLY1305_SHA256,
								// TLS 1.3 cipher suites.
								tls.TLS_AES_128_GCM_SHA256,
								tls.TLS_AES_256_GCM_SHA384,
								tls.TLS_CHACHA20_POLY1305_SHA256,
								},
							},
				},
	}

	// define http server and server handler
	mux := http.NewServeMux()
	mux.HandleFunc("/mutate", whsvr.serve)
	whsvr.server.Handler = mux

	// start webhook server in new routine
	go func() {
		if err := whsvr.server.ListenAndServeTLS("", ""); err != nil {
			panic(fmt.Errorf("Failed to listen and serve webhook server: %v", err))
		}
	}()

	fmt.Println("Server started")

	// listening OS shutdown singal
	signalChan := make(chan os.Signal, 1)
	signal.Notify(signalChan, syscall.SIGINT, syscall.SIGTERM)
	<-signalChan

	fmt.Println("Got OS shutdown signal, shutting down webhook server gracefully...")
	whsvr.server.Shutdown(context.Background())
}
