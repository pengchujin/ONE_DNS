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
	server := &dns.Server{Addr: "0.0.0.0:5555", Net: "udp" }
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
	var realIP net.IP
	if addr, ok := w.RemoteAddr().(*net.UDPAddr); ok {
		realIP = make(net.IP, len(addr.IP))
		copy(realIP, addr.IP)
	} else if addr, ok := w.RemoteAddr().(*net.TCPAddr); ok {
		realIP = make(net.IP, len(addr.IP))
		copy(realIP, addr.IP)
	}
	m := dns.Msg{}

	o := new(dns.OPT)
	o.Hdr.Name = "."
	o.Hdr.Rrtype = dns.TypeOPT
	e := new(dns.EDNS0_SUBNET)
	e.Code = dns.EDNS0SUBNET
	e.Family = 1	// 1 for IPv4 source address, 2 for IPv6
	e.SourceNetmask = 32	// 32 for IPV4, 128 for IPv6
	e.SourceScope = 0
	e.Address = realIP	// for IPv4
	o.Option = append(o.Option, e)
	r.Extra = append(r.Extra, o)
	// e.Address = net.ParseIP("2001:7b8:32a::2")	// for IPV6

	m.SetReply(r)
	log.Println(realIP)
	log.Println(r)
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