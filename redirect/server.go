package redirect

import (
	"context"
	"fmt"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Config struct {
	Port     int    `yaml:"port"`
	Location string `yaml:"location"`
}

var server *http.Server

func Start(conf Config) error {

	if server != nil {
		_ = server.Shutdown(context.TODO())
	}
	g := gin.Default()
	g.Any("/", func(c *gin.Context) {
		c.Redirect(http.StatusFound, conf.Location)
	})
	server = &http.Server{Handler: g, Addr: fmt.Sprintf(":%d", conf.Port)}
	return server.ListenAndServe()
}

func Stop() error {
	if server != nil {
		return server.Shutdown(context.TODO())
	}
	return nil
}
