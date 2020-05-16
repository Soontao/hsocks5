package hsocks5

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/pmezard/adblock/adblock"
	"golang.org/x/net/proxy"
)

// ProxyServer class
type ProxyServer struct {
	m             *adblock.RuleMatcher
	privateIPList *IPList
	cnIPList      *IPList
	c             *cache.Cache
	socksAddr     string
}

// NewProxyServer object
func NewProxyServer(socksProxyAddr string) (*ProxyServer, error) {
	c := cache.New(30*24*time.Hour, 1*time.Minute)
	pIPList := LoadIPListFrom("assets/private_ip_list.txt")
	cnIPList := LoadIPListFrom("assets/china_ip_list.txt")
	return &ProxyServer{m: LoadGFWList(), privateIPList: pIPList, cnIPList: cnIPList, c: c, socksAddr: socksProxyAddr}, nil
}

func (s *ProxyServer) createProxy() (proxy.Dialer, error) {
	return proxy.SOCKS5("tcp", s.socksAddr, nil, proxy.Direct)
}

func (s *ProxyServer) ServeHTTP(res http.ResponseWriter, req *http.Request) {

	if req.Method == "CONNECT" {

		s.handleConnect(res, req)

	} else {

		s.handleRequest(res, req)

	}

}

// isInGFWList url
func (s *ProxyServer) isInGFWList(url string) bool {

	b, _, _ := s.m.Match(&adblock.Request{URL: url})
	return b
}

// isDirectAccess the target
func (s *ProxyServer) isDirectAccess(hostnameOrURI string) (rt bool) {
	rt = true

	normalizeURI := hostnameOrURI

	if !strings.HasPrefix(normalizeURI, "http") {
		normalizeURI = fmt.Sprintf("https://%v", normalizeURI) // gfwlist must have protocol
	}

	url, err := url.Parse(normalizeURI)

	hostname := url.Hostname()

	if cachedValue, exist := s.c.Get(hostname); exist {
		return cachedValue.(bool)
	}

	if err != nil {
		log.Printf("parse url '%v' failed.", normalizeURI)
		return
	}

	if s.privateIPList.Contains(hostname) {
		rt = true // internal network
	} else if s.isInGFWList(normalizeURI) {
		rt = false // banned by gfw
	} else if !s.cnIPList.Contains(hostname) {
		rt = false // service server is not located in china
	}

	s.c.SetDefault(hostname, rt)

	return

}

func (s *ProxyServer) handleConnect(w http.ResponseWriter, req *http.Request) {

	host := req.Host // host & port
	hostname := req.URL.Hostname()

	log.Printf("CONNECT %v", host)

	hj, ok := w.(http.Hijacker)

	if !ok {
		w.WriteHeader(500)
		w.Write([]byte("Proxt Server Internal Error"))
		return // error break
	}

	conn, bufrw, err := hj.Hijack()

	if err != nil {
		log.Println(err)
		w.WriteHeader(500)
		w.Write([]byte("Proxt Server Internal Error"))
		return // error break
	}

	defer conn.Close()

	var remote net.Conn

	if s.isDirectAccess(hostname) {
		remote, err = net.Dial("tcp", host)
	} else {
		if dial, err := s.createProxy(); err == nil {
			remote, err = dial.Dial("tcp", host)
		}

	}

	if err != nil {
		log.Println(err)
		conn.Close()
		return // error break
	}

	defer remote.Close()

	bufrw.WriteString("HTTP/1.1 200 Connection established\r\n\r\n") // connect accept

	if err := bufrw.Flush(); err != nil {
		log.Println(err)
		remote.Close()
		conn.Close()
		return // error break
	}

	errChans := make(chan error, 2)

	go pipe(remote, conn, errChans)
	go pipe(conn, remote, errChans)

	<-errChans
	<-errChans

	// all transfer finished

}

func (s *ProxyServer) handleRequest(w http.ResponseWriter, req *http.Request) {
	host := req.Host // host & port
	log.Printf("HTTP %v %v", req.Method, host)

	var client http.Client

	if s.isDirectAccess(req.URL.RequestURI()) {
		client = http.Client{}
	} else {
		dialer, err := s.createProxy()
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			w.Write([]byte(fmt.Sprintf("http agent create failed, %v", err)))
			return
		}
		client = http.Client{Transport: &http.Transport{Dial: dialer.Dial}}
	}

	// create a new http request from original inbound request
	newReq, err := http.NewRequest(req.Method, req.URL.String(), req.Body)

	newReq.Header = req.Header

	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("http agent error happened, %v", err)))
		return
	}

	proxyResponse, err := client.Do(newReq)

	if err != nil {

		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		w.Write([]byte(fmt.Sprintf("http agent error happened, %v", err)))

	} else {

		for k, vs := range proxyResponse.Header {
			for _, v := range vs {
				w.Header().Set(k, v)
			}
		}

		w.WriteHeader(proxyResponse.StatusCode)

		defer proxyResponse.Body.Close()

		io.Copy(w, proxyResponse.Body)

	}

}

// Start server
func (s *ProxyServer) Start(addr string) error {
	hs := http.Server{Addr: addr, Handler: s}
	log.Printf("start server at %v", addr)
	return hs.ListenAndServe()
}
