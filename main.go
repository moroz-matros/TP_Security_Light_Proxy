package main

import (
	"io"
	"log"
	"net/http"
)

const (
	port      = ":8080"
)

type Server struct {
	Srv  http.Server
}

func newServer(port string) *Server {
	return &Server{
		Srv: http.Server{
			Addr: port,
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
				handleHTTP(w, r)
			}),
		},

	}
}

func handleHTTP(w http.ResponseWriter, r *http.Request) {
	if r.Proto != "HTTP/1.1" {
		http.Error(w, "Http support only", http.StatusForbidden)
	}

	resp, err := http.DefaultTransport.RoundTrip(r)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer resp.Body.Close()
	copyHeader(w.Header(), resp.Header)
	w.WriteHeader(resp.StatusCode)
	_, err = io.Copy(w, resp.Body)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
}

func copyHeader(dst, src http.Header) {
	for key, values := range src {
		for _, value := range values {
			dst.Add(key, value)
		}
	}
}

func main() {
	server := newServer(port)
	log.Fatal(server.Srv.ListenAndServe())
}

