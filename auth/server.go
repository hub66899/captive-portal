package auth

import (
	"bufio"
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	cache2 "github.com/patrickmn/go-cache"
	"net/http"
	"os/exec"
	"strings"
	"sync"
	"time"
)

type Config struct {
	Port          int `yaml:"port"`
	AssertionPort int `yaml:"assertion-port"`
}

var (
	server          *http.Server
	assertionServer *http.Server
	cache           = cache2.New(time.Minute, time.Minute*10)
)

const location = "https://filbet.cloudflareaccess.com/cdn-cgi/access/sso/saml/e2800b161176ab232098537103ea53fd74977309f48b3909867e7ca4b0685f60?RelayState=aaaa11111"

func Start(conf Config) error {
	_ = Stop()
	{
		g := gin.Default()
		g.GET("/", func(c *gin.Context) {
			c.Redirect(http.StatusFound, location)
			return
		})
		server = &http.Server{Handler: g, Addr: fmt.Sprintf(":%d", conf.Port)}
	}
	{
		g := gin.Default()
		g.GET("/", func(c *gin.Context) {
			c.String(200, "hello world")
		})
		assertionServer = &http.Server{Handler: g, Addr: fmt.Sprintf(":%d", conf.AssertionPort)}
	}
	var err error
	w := sync.WaitGroup{}
	w.Add(2)
	go func() {
		if e := server.ListenAndServe(); e != nil {
			err = e
		}
		w.Done()
	}()
	go func() {
		if e := assertionServer.ListenAndServe(); e != nil {
			err = e
		}
		w.Done()
	}()
	w.Wait()
	return err
}

func Stop() error {
	var err error
	if server != nil {
		if e := server.Shutdown(context.TODO()); e != nil {
			err = e
		}
	}
	if assertionServer != nil {
		if e := assertionServer.Shutdown(context.TODO()); e != nil {
			err = e
		}
	}
	return err
}

func getMACAddress(ip string) (string, error) {
	// 执行 arp 命令
	cmd := exec.Command("arp", "-n", ip)
	output, err := cmd.Output()
	if err != nil {
		return "", err
	}

	// 解析命令输出以找到 MAC 地址
	scanner := bufio.NewScanner(strings.NewReader(string(output)))
	for scanner.Scan() {
		line := scanner.Text()
		if strings.Contains(line, ip) {
			fields := strings.Fields(line)
			if len(fields) >= 3 {
				return fields[2], nil // MAC 地址通常位于第三个字段
			}
		}
	}

	return "", fmt.Errorf("MAC address for IP %s not found", ip)
}
