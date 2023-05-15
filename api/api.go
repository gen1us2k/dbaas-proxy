package api

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	vault "github.com/hashicorp/vault/api"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

type Service struct {
	router *gin.Engine
	v      *vault.Client
}
type KubeCluster struct {
	Name       string `json:"name"`
	Kubeconfig string `json:"kubeconfig"`
}

func New() (*Service, error) {
	s := &Service{
		router: gin.Default(),
	}
	config := vault.DefaultConfig()
	config.Address = "http://127.0.0.1:4321"
	client, err := vault.NewClient(config)
	if err != nil {
		return nil, err
	}
	client.SetToken("myroot")
	s.v = client
	s.init()
	return s, nil
}
func (s *Service) init() {
	s.router.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"PUT", "PATCH", "GET", "OPTIONS", "POST"},
		AllowHeaders:     []string{"Origin"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
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
	_, err := clientcmd.BuildConfigFromKubeconfigGetter("", NewConfigGetter(k.Kubeconfig).loadFromString)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"message": err})
		return
	}
	m := map[string]interface{}{
		"kubeconfig": k.Kubeconfig,
	}

	_, err = s.v.KVv2("secret").Put(context.TODO(), k.Name, m)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"message": err})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "success"})
}
func (s *Service) deleteK8s(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		return
	}
	err := s.v.KVv2("secret").Delete(context.TODO(), name)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"message": err})
		return
	}

	c.JSON(http.StatusOK, gin.H{"message": "deleted"})
}
func (s *Service) proxyK8s(c *gin.Context) {
	name := c.Param("name")
	if name == "" {
		return
	}
	kConfig, err := s.v.KVv2("secret").Get(context.TODO(), name)
	kubeconfig, ok := kConfig.Data["kubeconfig"].(string)
	if !ok {
		return
	}
	config, err := clientcmd.BuildConfigFromKubeconfigGetter("", NewConfigGetter(kubeconfig).loadFromString)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"message": err})
		return
	}
	data, err := json.Marshal(kConfig)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"message": err})
		return
	}
	err = json.Unmarshal(data, config)
	if err != nil {
		log.Println(err)
		c.JSON(http.StatusBadRequest, gin.H{"message": err})
		return
	}
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
