package main

import (
	"captive-portal/auth"
	"captive-portal/config"
	"captive-portal/dns"
	"captive-portal/redirect"
	"captive-portal/utils"
	"github.com/pkg/errors"
	"log"
	"net"
	"os"
)

type Config struct {
	InterfaceName string          `yaml:"interface-name"`
	Dns           dns.Config      `yaml:"dns"`
	Redirect      redirect.Config `yaml:"redirect"`
	Auth          auth.Config     `yaml:"auth"`
}

var defaultConfig = Config{
	Auth: auth.Config{
		Port:          8081,
		AssertionPort: 8080,
	},
}

var (
	configFile = "./config.yml"
	conf       Config
)

func init() {
	{
		v := os.Getenv("CONFIG_FILE")
		if v != "" {
			configFile = v
		}
	}
	localConfig, err := config.LocalYamlConfig[Config](configFile, defaultConfig)
	if err != nil {
		log.Fatalf("%+v", err)
	}
	c := localConfig.GetConfig()
	conf = *c
	//ip, err := getIP(conf.InterfaceName)
	//if err != nil {
	//	log.Fatalf("%+v", err)
	//}
	//conf.Dns.DefaultAnswerIP = ip
}

func main() {
	//go func() {
	//	if err := dns.Start(conf.Dns); err != nil {
	//		panic(err)
	//	}
	//}()
	//go func() {
	//	if err := redirect.Start(conf.Redirect); err != nil {
	//		panic(err)
	//	}
	//}()
	_ = auth.Start(conf.Auth)
	utils.OnShutdown(func() {
		//_ = dns.Stop()
		//_ = redirect.Stop()
		auth.Stop()
	})
}

func getIP(interfaceName string) (string, error) {
	// 通过名称获取网卡信息
	iface, err := net.InterfaceByName(interfaceName)
	if err != nil {
		return "", errors.WithStack(err)
	}

	// 获取网卡关联的地址信息
	addrs, err := iface.Addrs()
	if err != nil {
		return "", errors.WithStack(err)
	}
	if len(addrs) == 0 {
		return "", errors.Errorf("network interface %s no ip address", interfaceName)
	}
	switch v := addrs[0].(type) {
	case *net.IPNet:
		return v.IP.String(), nil
	case *net.IPAddr:
		return v.IP.String(), nil
	}
	return "", errors.Errorf("network interface %s has an unsupported address type", interfaceName)
}
