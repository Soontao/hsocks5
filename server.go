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

	"github.com/pmezard/adblock/adblock"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"golang.org/x/net/proxy"
)

// ProxyServer class
type ProxyServer struct {
	m         *adblock.RuleMatcher
	priIPList *IPList
	cnIPList  *IPList
	kvCache   *KVCache
	prom      http.Handler
	metric    *ProxyServerMetrics
	kl        *KeyLock
	option    *ProxyServerOption
}

// ProxyServerOption parameter
type ProxyServerOption struct {
	ListenAddr  string
	RedisAddr   string
	SocksAddr   string
	ChinaSwitch bool
}

// NewProxyServer object
func NewProxyServer(option *ProxyServerOption) (*ProxyServer, error) {
	pIPList := LoadIPListFrom("assets/private_ip_list.txt")
	cnIPList := LoadIPListFrom("assets/china_ip_list.txt")
	prom := promhttp.Handler()

	if option.ChinaSwitch {
		log.Println("enable smart traffic transfer for china")
	}

	return &ProxyServer{
		option:    option,
		kvCache:   NewKVCache(option.RedisAddr),
		m:         LoadGFWList(),
		priIPList: pIPList,
		cnIPList:  cnIPList,
		prom:      prom,
		metric:    NewProxyServerMetrics(),
		kl:        NewKeyLock(),
	}, nil

}

func (s *ProxyServer) createProxy() (proxy.Dialer, error) {
	return proxy.SOCKS5("tcp", s.option.SocksAddr, nil, proxy.Direct)
}

func (s *ProxyServer) ServeHTTP(res http.ResponseWriter, req *http.Request) {

	if req.Method == "CONNECT" {

		s.metric.connTotal.WithLabelValues("CONNECT").Inc()

		s.handleConnect(res, req)

		return

	}

	// handle direct http request
	if len(req.URL.Host) == 0 {

		s.handleRequest(res, req)
		return

	}

	// handle proxy http request

	s.metric.connTotal.WithLabelValues("REQUEST").Inc()

	s.handleProxyRequest(res, req)

	return

}

// isInGFWList url
func (s *ProxyServer) isInGFWList(url string) bool {
	b, _, _ := s.m.Match(&adblock.Request{URL: url})
	return b
}

// isWithoutProxy for the target
// 'true' means not require proxy
// 'false' means require proxy
func (s *ProxyServer) isWithoutProxy(hostnameOrURI string) (rt bool) {

	// for other user, proxy all requests to socks5 proxy
	if !s.option.ChinaSwitch {
		return false
	}

	s.metric.cacheHitTotal.WithLabelValues("check").Inc()

	reason := "FALLBACK"

	rt = true

	normalizeURI := hostnameOrURI

	if !strings.HasPrefix(normalizeURI, "http") {
		normalizeURI = fmt.Sprintf("https://%v", normalizeURI) // gfwlist must have protocol
	}

	url, err := url.Parse(normalizeURI)

	hostname := url.Hostname()

	// avoid the in-consistence for single hostname
	s.kl.Lock(hostname)
	defer s.kl.Unlock(hostname)

	// metric log
	defer func() {
		s.metric.routineResultTotal.WithLabelValues(hostname, strconv.FormatBool(rt), reason).Inc()
	}()

	if cachedValue, exist := s.kvCache.Get(hostname); exist { // use hostname as cache key
		reason = "CACHE"
		s.metric.cacheHitTotal.WithLabelValues("with_cache").Inc()
		if rt, err = strconv.ParseBool(cachedValue); err != nil {
			log.Printf("parse bool failed for '%v', please check your cahce", hostnameOrURI)
		}
		return
	}

	if err != nil {
		log.Printf("parse url '%v' failed.", normalizeURI)
		return
	}

	if s.priIPList.Contains(hostname) {
		reason = "IN_PRIVATE_NETWORK"
		rt = true // internal network
	} else if s.isInGFWList(normalizeURI) {
		reason = "IN_GFW_LIST"
		rt = false // banned by gfw
	} else if !s.cnIPList.Contains(hostname) {
		reason = "NOT_IN_CN_IP_LIST"
		rt = false // service server is not located in china
	}

	s.kvCache.Set(hostname, strconv.FormatBool(rt))

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

	if conn != nil {
		defer conn.Close()
	}

	var remote net.Conn

	if s.isWithoutProxy(hostname) {
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

	if remote != nil {
		defer remote.Close()
	}

	bufrw.WriteString("HTTP/1.1 200 Connection established\r\n\r\n") // connect accept

	if err := bufrw.Flush(); err != nil {
		log.Println(err)
		return // error break
	}

	if remote != nil && conn != nil {

		errChans := make(chan error, 2)

		go pipe(remote, conn, errChans)
		go pipe(conn, remote, errChans)

		<-errChans
		<-errChans

	}

	// all transfer finished
	// close connections with 'defer'

}

func (s *ProxyServer) handleRequest(w http.ResponseWriter, req *http.Request) {
	// >> Prometheus
	if req.RequestURI == "/hsocks5/__/metric" {
		s.prom.ServeHTTP(w, req)
		return
	}
	// << Prometheus

	// default 404 not exist
	w.WriteHeader(http.StatusNotFound)
	return
}

func (s *ProxyServer) handleProxyRequest(w http.ResponseWriter, req *http.Request) {
	host := req.Host // host & port
	log.Printf("HTTP %v %v", req.Method, host)

	var client http.Client

	if s.isWithoutProxy(req.RequestURI) {
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

	proxyResponse, err := client.Do(newReq) // if err happened, dont access proxyResponse

	if err != nil {

		log.Println(err)
		s.sendError(w, err)

	} else {

		s.metric.requestStatusTotal.WithLabelValues(newReq.URL.Hostname(), fmt.Sprint(proxyResponse.StatusCode)).Inc()
		s.pipeResponse(proxyResponse, w)

	}

}

// sendError alias
func (s *ProxyServer) sendError(w http.ResponseWriter, err error) {
	w.WriteHeader(http.StatusInternalServerError)
	w.Write([]byte(fmt.Sprintf("http proxy internal error happened, %v", err)))
}

// pipeResponse for http request
func (s *ProxyServer) pipeResponse(from *http.Response, to http.ResponseWriter) {
	h := to.Header()

	for k, vs := range from.Header {
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
	log.Printf("start server at %v", s.option.ListenAddr)
	return hs.ListenAndServe()
}
