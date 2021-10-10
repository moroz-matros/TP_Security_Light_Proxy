package main

import (
	"bufio"
	"crypto/tls"
	"io"
	"log"
	"math/rand"
	"net/http"
	"net/http/httputil"
	"os"
	"os/exec"
	"strconv"
	"strings"
)

const (
	port      = ":8080"
	pemFile = "gen_cert/ca.crt"
	keyCAFile = "gen_cert/ca.key"
	keyCertFile = "gen_cert/cert.key"
	certDir = "gen_cert/certs/"
)


type Server struct {
	Srv  http.Server
}

func newServer(port string) *Server {
	return &Server{
		Srv: http.Server{
			Addr: port,
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request){
				if r.Method == http.MethodConnect {
					handleHTTPS(w, r)
				} else {
					handleHTTP(w, r)
				}
			}),
		},
	}
}
func getCert(host string) (tls.Certificate, error){
	_, err := os.Stat(certDir + host + ".crt")
	if os.IsNotExist(err) {
		genCommand := exec.Command("gen_cert/gen_cert.sh", host, strconv.Itoa(rand.Intn(1000000)))

		_, err = genCommand.CombinedOutput()
		if err != nil {
			log.Println("error executing command getting cert")
			log.Println(err)
			return tls.Certificate{}, err
		}
	}

	file, err := tls.LoadX509KeyPair(certDir + host + ".crt", keyCertFile)
	if err != nil {
		log.Println("error loading pair", err)
		return tls.Certificate{}, err
	}

	return file, nil
}

func handleHTTPS(w http.ResponseWriter, r *http.Request) {
	// initialize tcp client
	// take over connection
	hijacker, ok := w.(http.Hijacker)
	if !ok {
		log.Println("Hijacking not supported")
		http.Error(w, "Hijacking not supported", http.StatusInternalServerError)
		return
	}

	clientConn, _, err := hijacker.Hijack()
	if err != nil {
		log.Println("hijacking error", err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
	}

	_, err = clientConn.Write([]byte("HTTP/1.1 200 Connection established\r\n\r\n"))
	if err != nil {
		log.Println("handshaking failed", err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		clientConn.Close()
		return
	}
	defer clientConn.Close()

	// get host
	host := strings.Split(r.Host, ":")[0]

	// get certs
	caCert, err := getCert(host)
	if err != nil {
		log.Println("error getting cert")
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}

	// create config
	tlsConfig := &tls.Config{
		Certificates: []tls.Certificate{caCert},
		ServerName: r.URL.Scheme,
	}

	// create conn with config
	tcpClient := tls.Server(clientConn, tlsConfig)
	err = tcpClient.Handshake()
	if err != nil {
		tcpClient.Close()
		log.Println("handshaking failed", err)
		return
	}
	defer tcpClient.Close()

	// create tcp server
	tcpServer, err := tls.Dial("tcp", r.URL.Host, tlsConfig)
	if err != nil {
		log.Println(err)
		http.Error(w, err.Error(), http.StatusServiceUnavailable)
		return
	}
	defer tcpServer.Close()

	// get request
	clientReader := bufio.NewReader(tcpClient)
	request, err := http.ReadRequest(clientReader)
	if err != nil {
		log.Println("error getting request", err)
		return
	}

	//serve request
	dumpRequest, err := httputil.DumpRequest(request, true)
	if err != nil {
		log.Println("failed to dump request", err)
		return
	}
	_, err = tcpServer.Write(dumpRequest)
	if err != nil {
		log.Println("failed to write request", err)
		return
	}

	serverReader := bufio.NewReader(tcpServer)
	response, err := http.ReadResponse(serverReader, request)
	if err != nil {
		log.Println("failed to read response", err)
		return
	}

	rawResponse, err := httputil.DumpResponse(response, true)
	if err != nil {
		log.Println("failed to dump response", err)
		return
	}

	_, err = tcpClient.Write(rawResponse)
	if err != nil {
		log.Println("fail to write response: ", err)
		return
	}
}

func handleHTTP(w http.ResponseWriter, r *http.Request) {
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

