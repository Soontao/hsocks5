package hsocks5

import (
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/patrickmn/go-cache"
	"github.com/pmezard/adblock/adblock"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/net/proxy"
)

// ProxyServer class
type ProxyServer struct {
	m         *adblock.RuleMatcher
	priIPList *IPList
	cnIPList  *IPList
	c         *cache.Cache
	prom      http.Handler
	socksAddr string
	metric    *ProxyServerMetrics
}

// NewProxyServer object
func NewProxyServer(socksAddr string) (*ProxyServer, error) {
	c := cache.New(30*24*time.Hour, 1*time.Minute)
	pIPList := LoadIPListFrom("assets/private_ip_list.txt")
	cnIPList := LoadIPListFrom("assets/china_ip_list.txt")
	prom := promhttp.Handler()

	return &ProxyServer{
		m:         LoadGFWList(),
		priIPList: pIPList,
		cnIPList:  cnIPList,
		c:         c,
		socksAddr: socksAddr,
		prom:      prom,
		metric:    NewProxyServerMetrics(),
	}, nil

}

func (s *ProxyServer) createProxy() (proxy.Dialer, error) {
	return proxy.SOCKS5("tcp", s.socksAddr, nil, proxy.Direct)
}

func (s *ProxyServer) ServeHTTP(res http.ResponseWriter, req *http.Request) {

	if req.Method == "CONNECT" {

		s.metric.connTotal.WithLabelValues("CONNECT").Inc()

		s.handleConnect(res, req)

	} else {

		if req.RequestURI == "/hsocks5/__/metric" {
			s.prom.ServeHTTP(res, req)
			return
		}

		if req.RequestURI == "/favicon.ico" {
			res.WriteHeader(http.StatusNotFound)
			return
		}

		s.metric.connTotal.WithLabelValues("REQUEST").Inc()

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

	s.metric.cacheHitTotal.WithLabelValues("check").Inc()

	routineType := "FALLBACK"

	rt = true

	normalizeURI := hostnameOrURI

	if !strings.HasPrefix(normalizeURI, "http") {
		normalizeURI = fmt.Sprintf("https://%v", normalizeURI) // gfwlist must have protocol
	}

	url, err := url.Parse(normalizeURI)

	hostname := url.Hostname()

	defer func() {
		s.metric.routineResultTotal.WithLabelValues(hostname, strconv.FormatBool(rt), routineType).Inc()
	}()

	if cachedValue, exist := s.c.Get(hostname); exist { // use hostname as cache key
		s.metric.cacheHitTotal.WithLabelValues("with_cache").Inc()
		return cachedValue.(bool)
	}

	if err != nil {
		log.Printf("parse url '%v' failed.", normalizeURI)
		return
	}

	if s.priIPList.Contains(hostname) {
		routineType = "IN_PRIVATE_NETWORK"
		rt = true // internal network
	} else if s.isInGFWList(normalizeURI) {
		routineType = "IN_GFW_LIST"
		rt = false // banned by gfw
	} else if !s.cnIPList.Contains(hostname) {
		routineType = "NOT_IN_CN_IP_LIST"
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
		s.sendError(w, fmt.Errorf("Proxt Server Internal Error"))
		return // error break
	}

	conn, bufrw, err := hj.Hijack() // get TCp connection

	if err != nil {
		log.Println(err)
		s.sendError(w, err)
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

	if s.isDirectAccess(req.RequestURI) {
		client = http.Client{}
	} else {
		dialer, err := s.createProxy()
		if err != nil {
			s.sendError(w, err)
			return
		}
		client = http.Client{Transport: &http.Transport{Dial: dialer.Dial}}
	}

	// create a new http request from original inbound request
	newReq, err := http.NewRequest(req.Method, req.URL.String(), req.Body)

	newReq.Header = req.Header.Clone()

	if err != nil {
		log.Println(err)
		s.sendError(w, err)
		return
	}

	proxyResponse, err := client.Do(newReq)

	s.metric.requestStatusTotal.WithLabelValues(newReq.URL.Hostname(), string(proxyResponse.StatusCode)).Inc()

	if err != nil {

		log.Println(err)
		s.sendError(w, err)

	} else {

		s.pipeResponse(proxyResponse, w)

	}

}

func (s *ProxyServer) sendError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(fmt.Sprintf("http agent error happened, %v", err)))
}

func (s *ProxyServer) pipeResponse(from *http.Response, to http.ResponseWriter) {
	for k, vs := range from.Header {
		h := to.Header()
		for _, v := range vs {
			h.Set(k, v)
		}
	}

	to.WriteHeader(from.StatusCode)

	defer from.Body.Close()

	io.Copy(to, from.Body)
}

// Start server
func (s *ProxyServer) Start(addr string) error {
	hs := http.Server{Addr: addr, Handler: s}
	log.Printf("start server at %v", addr)
	return hs.ListenAndServe()
}
