package main

import (
	"log"
	"github.com/miekg/dns"
	"net"
)

var records = map[string]string{
	"test.service.": "192.168.0.2",
}

func main () {
	server := &dns.Server{Addr: "127.0.0.1:5555", Net: "udp" }
	server.Handler = &dnsServer{}
	log.Println("Booting DNS Server......")
	err := server.ListenAndServe()
	if err != nil {
		log.Println("Something going wrong", err)
	}
	defer server.Shutdown()
}

type dnsServer struct {
	
}

func (this *dnsServer) ServeDNS (w dns.ResponseWriter, r *dns.Msg) {
	m := dns.Msg{}
	m.SetReply(r)
	switch r.Question[0].Qtype {
	case dns.TypeA:
		domain := m.Question[0].Name
		address, ok := records[domain]
		if ok {
			m.Authoritative = true
			m.Answer = append(m.Answer, &dns.A{
				Hdr: dns.RR_Header{ Name: domain, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 60},
				A: net.ParseIP(address),
			})
			w.WriteMsg(&m)
		} else {
			// c := dns.Client{Net: "udp", ReadTimeout: 2*time.Second, WriteTimeout: 2*time.Second,}
			c := new(dns.Client)
			res, _, err := c.Exchange(r, "8.8.8.8:53")
			log.Println(res)
			if err != nil {
				log.Println(err, res)
			}
			w.WriteMsg(res)
		}
	}
	
}