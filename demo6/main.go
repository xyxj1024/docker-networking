package main

import (
	"flag"
	"math/rand"
	"net"
	"os"
	"os/signal"
	"strings"
	"sync"
	"syscall"
	"time"

	"github.com/miekg/dns" // https://www.rfc-editor.org/rfc/rfc1035.html
	logrus "github.com/sirupsen/logrus"
)

type Addresses []string

func (a *Addresses) String() string {
	return "my string representation"
}

func (a *Addresses) Set(val string) error {
	*a = append(*a, val)
	return nil
}

var (
	client       Client
	nodes        []SwarmNode
	swarmDomains Addresses
	rateLimit    int64
	mu           = &sync.Mutex{}
)

const (
	NodeRefreshInterval = 60
	TTL                 = NodeRefreshInterval
)

func init() {
	flag.Var(&swarmDomains, "domain", "(Required) Domain to resolve addresses for (can be specified multiple times)")
	flag.Int64Var(&rateLimit, "ratelimit", 0, "(Optional) Number of concurrent requests")
}

func main() {
	flag.Parse()

	if len(swarmDomains) == 0 {
		flag.Usage()
		logrus.Fatalf("Aborting: No domains given.")
	}
	logrus.Infof("Using domains: %v", swarmDomains)

	var err error
	client, err = NewClient()
	if err != nil {
		logrus.Error(err)
	}

	refreshNodes()

	ticker := time.NewTicker(time.Second * NodeRefreshInterval)
	go func() {
		for range ticker.C {
			refreshNodes()
		}
	}()

	var handler dns.HandlerFunc
	if rateLimit > 0 {
		logrus.Infof("Limiting the number of concurrent requests to %d", rateLimit)
		l := make(chan struct{}, rateLimit)
		handler = dns.HandlerFunc(func(w dns.ResponseWriter, r *dns.Msg) {
			l <- struct{}{}
			defer func() { <-l }()
			handleRequest(w, r)
		})
	} else {
		handler = dns.HandlerFunc(handleRequest)
	}

	go func() {
		srv := &dns.Server{Addr: ":53", Net: "udp", Handler: handler}
		err := srv.ListenAndServe()
		if err != nil {
			logrus.Fatalf("Failed to set udp listener %s\n", err.Error())
		}
	}()

	sig := make(chan os.Signal)
	signal.Notify(sig, syscall.SIGINT, syscall.SIGTERM)
	for {
		select {
		case s := <-sig:
			logrus.Fatalf("Signal (%d) received, stopping\n", s)
			ticker.Stop()
		}
	}
}

func answerForNodes(domain string) []dns.RR {
	mu.Lock()
	var rrs []dns.RR // resource records
	for _, node := range nodes {
		if node.IsManager {
			rr := new(dns.A) // a host address
			rr.Hdr = dns.RR_Header{Name: domain, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: uint32(TTL)}
			rr.A = net.ParseIP(node.Ip)
			rrs = append(rrs, rr)
		}
	}
	mu.Unlock()

	return shuffleRRs(rrs)
}

func matchingDomain(domain string) *string {
	normalizedDomain := strings.ToLower(domain)

	for _, name := range swarmDomains {
		if (normalizedDomain == name+".") || strings.HasSuffix(normalizedDomain, "."+name+".") {
			return &name
		}
	}
	return nil
}

func shuffleRRs(src []dns.RR) []dns.RR {
	dest := make([]dns.RR, len(src))
	perm := rand.Perm(len(src))
	for i, v := range perm {
		dest[v] = src[i]
	}
	return dest
}

func handleRequest(w dns.ResponseWriter, r *dns.Msg) {
	m := new(dns.Msg)
	m.SetReply(r)
	m.Authoritative = true

	domain := r.Question[0].Name
	swarmDomain := matchingDomain(domain)

	if swarmDomain != nil && r.Question[0].Qtype == 1 { // Only answer questions for A records on supported domains
		ip, _, _ := net.SplitHostPort(w.RemoteAddr().String())
		logrus.Infof("Request: %15s %s", ip, domain)
		m.Answer = answerForNodes(domain)
	} else {
		m.Answer = []dns.RR{}
	}

	w.WriteMsg(m)
}

func refreshNodes() {
	var err error
	mu.Lock()
	nodes, err = client.ListActiveNodes()
	logrus.Infof("Refreshed nodes: %v\n", nodes)
	mu.Unlock()
	if err != nil {
		panic(err)
	}
}
