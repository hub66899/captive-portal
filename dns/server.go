package dns

import (
	"errors"
	"fmt"
	"github.com/miekg/dns"
	cache2 "github.com/patrickmn/go-cache"
	"log"
	"net"
	"os/exec"
	"regexp"
	"strings"
	"time"
)

type Config struct {
	Port                 int      `yaml:"port"`
	AllowDomains         []string `yaml:"allow-domains"`
	WhiteListSet         string   `yaml:"white-list-set"`
	WhiteListDataFile    string   `yaml:"white-list-data-file"`
	WhiteListTimeoutHour int      `yaml:"white-list-timeout-hour"`
	DefaultAnswerIP      string   `yaml:"default-answer-ip"`
}

var (
	reg          *regexp.Regexp
	srv          *dns.Server
	cache        *cache2.Cache
	expiration   time.Duration
	whitelistSet string
	dataFile     string
	defaultA     net.IP
)

func Start(conf Config) error {
	//expiration
	{
		expiration = time.Hour * 24 * 2
		if conf.WhiteListTimeoutHour != 0 {
			expiration = time.Duration(conf.WhiteListTimeoutHour) * time.Hour
		}
	}
	//whitelistSet
	{
		whitelistSet = conf.WhiteListSet
		if err := flushIpSet(); err != nil {
			log.Printf("flush ip set error: %v", err)
		}
	}
	//regexp
	{
		reg = nil
		keywords := conf.AllowDomains
		if len(keywords) > 0 {
			r, err := regexp.Compile(fmt.Sprintf("(%s)", strings.Join(keywords, "|")))
			if err != nil {
				return err
			}
			reg = r
			log.Printf("read allow-domains %s \n", strings.Join(keywords, ","))
		} else {
			log.Printf("没有配置allow-domains")
		}
	}
	//cache
	dataFile = conf.WhiteListDataFile
	if err := initCache(); err != nil {
		return err
	}
	//serverIP
	{
		ip := net.ParseIP(conf.DefaultAnswerIP)
		if ip == nil {
			return fmt.Errorf("default answer ip invalid: %s", conf.DefaultAnswerIP)
		}
		defaultA = ip
	}
	//server
	srv = &dns.Server{Addr: fmt.Sprintf(":%d", conf.Port), Net: "udp"}
	dns.HandleFunc(".", handleDNSQuery)
	return srv.ListenAndServe()
}

func initCache() error {
	cache = cache2.New(time.Hour*24*2, time.Hour)
	cache.OnEvicted(func(s string, i interface{}) {
		if err := removeIpSet(s); err != nil {
			log.Printf("%v\n", err)
		}
	})
	if err := cache.LoadFile(dataFile); err != nil {
		log.Printf("加载缓存文件失败 %v", err)
		return nil
	}
	var ips []string
	for ip := range cache.Items() {
		ips = append(ips, ip)
	}
	if l := len(ips); l > 0 {
		cmd := exec.Command("nft", "add", "element", "inet", "fw4", whitelistSet, "{"+strings.Join(ips, ",")+"}")
		log.Printf("初始化%d个ip", l)
		return cmd.Run()
	}
	return nil
}

func Stop() error {
	if srv != nil {
		return errors.New("Uninitialized")
	}
	if err := flushIpSet(); err != nil {
		log.Printf("flush ip set error %v", err)
	}
	if err := cache.SaveFile(dataFile); err != nil {
		log.Printf("save cache error %v", err)
	}
	return srv.Shutdown()
}

const upstream = "127.0.0.1:53"

func handleDNSQuery(w dns.ResponseWriter, r *dns.Msg) {
	//白名單 請求上游並加白ip
	if reg != nil {
		name := r.Question[0].Name
		if reg.MatchString(name) {
			client := &dns.Client{}
			msg, _, err := client.Exchange(r, upstream)
			if err != nil {
				log.Printf("Failed to forward query to upstream %s: %v", upstream, err)
				m := new(dns.Msg)
				m.SetRcode(r, dns.RcodeServerFailure)
				_ = w.WriteMsg(m)
				return
			}
			log.Printf("Failed to forward query to upstream %s: %v", upstream, err)
			for _, ans := range msg.Answer {
				if a, ok := ans.(*dns.A); ok {
					if err = addIp(a.A); err != nil {
						log.Printf("add ip %s failed:%v", a.A.String(), err)
					}
				}
			}
			if err = w.WriteMsg(msg); err != nil {
				log.Printf("Failed to write response: %v", err)
			}
			return
		}
	}
	m := new(dns.Msg)
	m.SetReply(r)
	m.Compress = false

	switch r.Opcode {
	case dns.OpcodeQuery:
		for _, q := range m.Question {
			switch q.Qtype {
			case dns.TypeA:
				rr := &dns.A{
					Hdr: dns.RR_Header{Name: q.Name, Rrtype: dns.TypeA, Class: dns.ClassINET, Ttl: 0},
					A:   defaultA,
				}
				m.Answer = append(m.Answer, rr)
			}
		}
	}
	if err := w.WriteMsg(m); err != nil {
		log.Printf("Failed to write response: %v", err)
	}
}

func addIp(ip net.IP) error {
	if ip == nil || ip.To4() == nil {
		return fmt.Errorf("ip %s invalid", ip)
	}
	ipStr := ip.To4().String()
	_, exist := cache.Get(ipStr)
	defer func() {
		cache.Set(ipStr, "", expiration)
	}()
	if exist {
		return nil
	}
	log.Printf("added ip %s\n", ipStr)
	return addIpSet(ipStr)
}

func addIpSet(ipStr string) error {
	cmd := exec.Command("nft", "add", "element", "inet", "fw4", whitelistSet, "{"+ipStr+"}")
	return cmd.Run()
}

func removeIpSet(ipStr string) error {
	cmd := exec.Command("nft", "delete", "element", "inet", "fw4", whitelistSet, "{"+ipStr+"}")
	return cmd.Run()
}

func flushIpSet() error {
	cmd := exec.Command("nft", "flush", "set", "inet", "fw4", whitelistSet)
	return cmd.Run()
}
