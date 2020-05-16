package hsocks5

import (
	"fmt"
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
	p             proxy.Dialer
	m             *adblock.RuleMatcher
	privateIPList *IPList
	cnIPList      *IPList
	c             *cache.Cache
}

// NewProxyServer object
func NewProxyServer(socksProxyAddr string) (*ProxyServer, error) {
	dialer, err := proxy.SOCKS5("tcp", socksProxyAddr, nil, proxy.Direct)
	if err != nil {
		return nil, err
	}
	c := cache.New(30*24*time.Hour, 1*time.Minute)
	pIPList := LoadIPListFrom("assets/private_ip_list.txt")
	cnIPList := LoadIPListFrom("assets/china_ip_list.txt")
	return &ProxyServer{p: dialer, m: LoadGFWList(), privateIPList: pIPList, cnIPList: cnIPList, c: c}, nil
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

	if strings.HasPrefix(normalizeURI, "http") {
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

func (s *ProxyServer) handleConnect(res http.ResponseWriter, req *http.Request) {

	host := req.Host // host & port
	hostname := req.URL.Hostname()

	log.Printf("CONNECT %v", host)

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

	var remote net.Conn

	if s.isDirectAccess(hostname) {
		remote, err = net.Dial("tcp", host)
	} else {
		remote, err = s.p.Dial("tcp", host)
	}

	if err != nil {
		log.Println(err)
		conn.Close()
		return // error break
	}

	bufrw.WriteString("HTTP/1.1 200 OK\r\n\r\n") // connect accept

	if err := bufrw.Flush(); err != nil {
		log.Println(err)
		remote.Close()
		conn.Close()
		return // error break
	}

	errChan := make(chan error, 2)

	go pipe(conn, remote, errChan)
	go pipe(remote, conn, errChan)

	<-errChan // ignore error
	<-errChan

	// all connection closed

}

func (s *ProxyServer) handleRequest(res http.ResponseWriter, req *http.Request) {
	host := req.Host // host & port
	log.Printf("HTTP %v %v", req.Method, host)

	var client http.Client

	if s.isDirectAccess(req.URL.RequestURI()) {
		client = http.Client{}
	} else {
		client = http.Client{Transport: &http.Transport{Dial: s.p.Dial}}
	}

	newReq, err := http.NewRequest(req.Method, req.URL.String(), req.Body)

	if err != nil {
		log.Println(err)
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte(fmt.Sprintf("http agent error happened, %v", err)))
		return
	}

	result, err := client.Do(newReq)

	if err != nil {
		log.Println(err)
		res.WriteHeader(http.StatusInternalServerError)
		res.Write([]byte(fmt.Sprintf("http agent error happened, %v", err)))
	} else {
		result.Write(res)
	}

}

// Start server
func (s *ProxyServer) Start(addr string) error {
	hs := http.Server{Addr: addr, Handler: s}
	log.Printf("start server at %v", addr)
	return hs.ListenAndServe()
}
