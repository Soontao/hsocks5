package hsocks5

import (
	"io"
	"log"
	"net"
	"net/http"

	"golang.org/x/net/proxy"
)

// ProxyServer class
type ProxyServer struct {
	p proxy.Dialer
}

// NewProxyServer object
func NewProxyServer(socksProxyAddr string) (*ProxyServer, error) {
	dialer, err := proxy.SOCKS5("tcp", socksProxyAddr, nil, proxy.Direct)
	if err != nil {
		return nil, err
	}
	return &ProxyServer{p: dialer}, nil
}

func pipe(c1, c2 net.Conn, errChan chan error) {
	_, err := io.Copy(c1, c2)
	if err != nil {
		log.Println(err)
	} else {
		log.Printf("Connection End")
	}
	c1.Close()
	c2.Close()
	errChan <- err
}

func (s *ProxyServer) ServeHTTP(res http.ResponseWriter, req *http.Request) {

	if req.Method == "CONNECT" {
		hj, ok := res.(http.Hijacker)

		if !ok {
			res.WriteHeader(500)
			res.Write([]byte("Proxt Server Internal Error"))
			return // error break
		}

		conn, bufrw, err := hj.Hijack()

		if err != nil {
			log.Println(err)
			res.WriteHeader(500)
			res.Write([]byte("Proxt Server Internal Error"))
			return // error break
		}

		proxyConn, err := s.p.Dial("tcp", req.Host)
		if err != nil {
			log.Println(err)
			conn.Close()
			return // error break
		}

		bufrw.WriteString("HTTP/1.1 200 OK\r\n\r\n") // connect accept

		if err := bufrw.Flush(); err != nil {
			log.Println(err)
			proxyConn.Close()
			conn.Close()
			return // error break
		}

		errChan := make(chan error, 2)

		go pipe(conn, proxyConn, errChan)
		go pipe(proxyConn, conn, errChan)

		e1, e2 := <-errChan, <-errChan

		if e1 != nil {
			log.Println(e1)
		}

		if e2 != nil {
			log.Println(e2)
		}

	}

}

// Start server
func (s *ProxyServer) Start(addr string) error {
	hs := http.Server{Addr: addr, Handler: s}
	return hs.ListenAndServe()
}
