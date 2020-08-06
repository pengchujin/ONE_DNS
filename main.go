package main

import (
	"log"
	"github.com/miekg/dns"
	"net"
	"context"
	"github.com/go-redis/redis/v8"
	"encoding/json"
)

var ctx = context.Background()

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

	opt, err := redis.ParseURL("redis://localhost:6379/0")
	if err != nil {
			panic(err)
	}
	rdb := redis.NewClient(opt)

	//获取客户端IP
	var realIP net.IP
	if addr, ok := w.RemoteAddr().(*net.UDPAddr); ok {
		realIP = make(net.IP, len(addr.IP))
		copy(realIP, addr.IP)
	} else if addr, ok := w.RemoteAddr().(*net.TCPAddr); ok {
		realIP = make(net.IP, len(addr.IP))
		copy(realIP, addr.IP)
	}
	log.Println(realIP)
	m := dns.Msg{}
	m.SetReply(r)

	domain := m.Question[0].Name

	val, err := rdb.Get(ctx, domain).Result()

	log.Println("++++++++++++ ",r.Question[0].Qtype, " ++++++++++++")

	if err != nil {
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
				
				if err != nil {
					log.Println(err, res)
				}
				
				answerString, err := json.Marshal(res.Answer)
				if err != nil {
					log.Println(nil)
				}
	
				error := rdb.Set(ctx, m.Question[0].Name, string(answerString), 0).Err()
	
				if(err != nil) {
					panic(error)
				}
	
				w.WriteMsg(res)
			}
		
		default:
			// c := dns.Client{Net: "udp", ReadTimeout: 2*time.Second, WriteTimeout: 2*time.Second,}
			c := new(dns.Client)
			res, _, err := c.Exchange(r, "8.8.8.8:53")
			if err != nil {
				log.Println(err, res)
			}
	
			answerString, err := json.Marshal(res.Answer)
				if err != nil {
					log.Println(nil)
				}
	
			error := rdb.Set(ctx, m.Question[0].Name, string(answerString), 0).Err()
	
			if(err != nil) {
				panic(error)
			}
	
			w.WriteMsg(res)
		}
	} 
	
	m.Authoritative = true
	dnsA := []dns.A{}
	err = json.Unmarshal([]byte(val), &dnsA)
	log.Println(dnsA)

	for _, value := range dnsA {
		m.Answer = append(m.Answer, &value)
	}

	if err != nil {
		log.Println(err)
	}
	m.Authoritative = true
	w.WriteMsg(&m)

}
