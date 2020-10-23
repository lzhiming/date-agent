package date_agent

import (
	"fmt"
	"github.com/gin-gonic/gin"
	"io"
	"k8s.io/klog"
	"net/http"
	"os"
	"strings"
)

func header() gin.HandlerFunc {
	return func(c *gin.Context) {
		method := c.Request.Method
		origin := c.Request.Header.Get("Origin")
		var headerKeys []string
		for k, v := range c.Request.Header {
			_ = v
			headerKeys = append(headerKeys, k)
		}
		headerStr := strings.Join(headerKeys, ", ")
		if headerStr != "" {
			headerStr = fmt.Sprintf("access-control-allow-origin, access-control-allow-headers, %s", headerStr)
		} else {
			headerStr = "access-control-allow-origin, access-control-allow-headers"
		}
		if origin != "" {
			c.Writer.Header().Set("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Origin", origin)
			c.Header("Access-Control-Allow-Methods", "POST, GET, OPTIONS, PUT, DELETE,UPDATE")
			//  header types
			c.Header("Access-Control-Allow-Headers", "Authorization, Content-Length, X-CSRF-Token, Token,session,X_Requested_With,Accept, Origin, Host, Connection, Accept-Encoding, Accept-Language,DNT, X-CustomHeader, Keep-Alive, User-Agent, X-Requested-With, If-Modified-Since, Cache-Control, Content-Type, Pragma")
			c.Header("Access-Control-Expose-Headers", "Content-Length, Access-Control-Allow-Origin, Access-Control-Allow-Headers,Cache-Control,Content-Language,Content-Type,Expires,Last-Modified,Pragma,FooBar")
			c.Header("Access-Control-Max-Age", "172800")
			c.Header("Access-Control-Allow-Credentials", "true")
			c.Set("content-type", "application/json")
		}
		if method == "OPTIONS" {
			c.JSON(http.StatusOK, "Options Request!")
		}
		c.Next()
	}
}

func InitHttp(addr string, hub *Hub) *http.Server {
	f, _ := os.Create("gin.log")
	gin.DefaultWriter = io.MultiWriter(f, os.Stdout)

	router := gin.Default()
	//router.Use(gin.LoggerWithConfig(gin.LoggerConfig{Output: writer}), gin.RecoveryWithWriter(writer))
	router.Use(header())
	router.LoadHTMLGlob("templates/*")
	router.StaticFS("/statics", http.Dir("./statics"))
	router.GET("/hello", func(c *gin.Context) {
		// todo handle request and return data by hub
		for index, value := range hub.tasks {
			fmt.Println("taskId", hub.taskId)
			fmt.Printf("%+v\n", index)
			fmt.Printf("%+v\n", value)
		}
		fmt.Println("hub.ret", hub.ret)
		c.JSON(http.StatusOK, hub.ret)
	})

	router.GET("/getJobs", func(c *gin.Context) {
		fmt.Printf("%+v\n", hub)
		c.JSON(http.StatusOK, hub.ret)
	})

	router.POST("/changeTime", func(c *gin.Context) {
		//name := c.PostForm("hostname")
		command := c.PostForm("command")
		go func() {
			//<-time.After(time.Second * 10)
			klog.Info("new task")
			hub.NewTask([]string{command})
		}()
		c.JSON(http.StatusOK, hub.ret)
	})

	router.POST("/getHub", func(c *gin.Context) {
		var l int
		if len(hub.tasks) > 5 {
			l = len(hub.tasks) - 5
		}
		c.JSONP(http.StatusOK, gin.H{"tasks": hub.tasks[l:], "ret": hub.ret})
	})

	router.GET("/index", func(c *gin.Context) {
		c.HTML(http.StatusOK, "index.tmpl", gin.H{})
	})

	server := &http.Server{
		Addr:    addr,
		Handler: router,
	}
	go func() {
		if err := server.ListenAndServe(); err != nil {
			if err == http.ErrServerClosed {
				klog.Info("Server closed under request")
			} else {
				klog.V(2).Info("Server closed unexpected err:", err)
			}
		}
	}()
	return server
}
