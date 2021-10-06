package main

import (
	"crypto/tls"
	"io"
	"log"
	"net"
	"net/http"
)

const (
	port      = ":8080"
	pemFile = "gen_cert/ca.crt"
	keyFile = "gen_cert/ca.key"
)


type Server struct {
	Srv  http.Server
	TLSNextProto map[string]func(*http.Server, *tls.Conn, http.Handler)
}

func newServer(port string) *Server {
	return &Server{
		Srv: http.Server{
			Addr: port,
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
				if r.Method == http.MethodConnect {
					log.Println(1)
					handleHTTPS(w, r)
				} else {
					handleHTTP(w, r)
				}
			}),
		},
		TLSNextProto: make(map[string]func(*http.Server, *tls.Conn, http.Handler)),
	}
}

func handleHTTPS(w http.ResponseWriter, r *http.Request) {
	log.Println("connect")
	dest_conn, err := net.Dial("tcp", r.Host)
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	w.WriteHeader(http.StatusOK)

	hijacker, ok := w.(http.Hijacker)
	if !ok {
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}
	client_conn, _, err := hijacker.Hijack()
	if err != nil {
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}
	go transfer(dest_conn, client_conn)
	go transfer(client_conn, dest_conn)

}
func transfer(destination io.WriteCloser, source io.ReadCloser) {
	defer destination.Close()
	defer source.Close()
	io.Copy(destination, source)
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

