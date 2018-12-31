package main

import (
	"net/http"
	"time"
)

// Gomitmproxy create a mitm proxy and start it
func Gomitmproxy(addr string, ch chan bool) {
	tlsConfig := NewTLSConfig("ca-pk.pem", "ca-cert.pem", "", "")
	handler := InitConfig(tlsConfig)
	server := &http.Server{
		Addr:         addr,
		ReadTimeout:  1 * time.Hour,
		WriteTimeout: 1 * time.Hour,
		Handler:      handler,
	}

	go func() {
		server.ListenAndServe()
		ch <- true
	}()

	return
}
