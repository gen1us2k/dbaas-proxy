package api

import (
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/gen1us2k/dbaas-proxy/storage"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Service struct {
	router  *gin.Engine
	storage *storage.Storage
}
type KubeCluster struct {
	Name       string `json:"name"`
	Kubeconfig string `json:"kubeconfig"`
}

func New() *Service {
	s := &Service{
		router:  gin.Default(),
		storage: storage.New(),
	}
	s.init()
	return s
}
func (s *Service) init() {
	s.router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"PUT", "PATCH", "GET", "OPTIONS", "POST"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return origin == "https://github.com"
		},
		MaxAge: 12 * time.Hour,
	}))
	s.router.POST("/k8s", s.addK8s)
	s.router.DELETE("/k8s/:name", s.deleteK8s)
	s.router.Any("/proxy/:name/*proxyPath", s.proxyK8s)
}
func (s *Service) addK8s(c *gin.Context) {
	var k KubeCluster
	if err := c.BindJSON(&k); err != nil {
		log.Println(err)
		return
	}
	log.Println(k.Kubeconfig)
	config, err := clientcmd.BuildConfigFromKubeconfigGetter("", NewConfigGetter(k.Kubeconfig).loadFromString)
	if err != nil {
		log.Println(err)
		return
	}
	s.storage.Add(k.Name, config)
	c.JSON(http.StatusOK, gin.H{"message": "success"})
}
func (s *Service) deleteK8s(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		return
	}
	s.storage.Delete(name)
	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
func (s *Service) proxyK8s(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		return
	}
	config := s.storage.Get(name)
	reverseProxy := httputil.NewSingleHostReverseProxy(&url.URL{
		Host:   strings.TrimPrefix(config.Host, "https://"),
		Scheme: "https",
	})
	transport, err := rest.TransportFor(config)
	if err != nil {
		return
	}
	reverseProxy.Transport = transport
	req := c.Request
	req.URL.Path = strings.TrimLeft(req.URL.Path, fmt.Sprintf("/proxy/%s", name))
	reverseProxy.ServeHTTP(c.Writer, req)
}
func (s *Service) Run() error {
	return s.router.Run("localhost:8080")
}
