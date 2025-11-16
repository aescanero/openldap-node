package apiserver

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"
	"sync"
	"time"

	"github.com/aescanero/openldap-node/config"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

//go:embed dashboard/build
var dashboard embed.FS

const INDEX = "index.html"

func Server(apiconfig config.Config) {
	var wgServer sync.WaitGroup
	stateError := make(chan error)
	wgServer.Add(1)

	go poolMonitor(apiconfig, stateError)

	// Initialize OAuth2 server only if OAuth2 is enabled
	if apiconfig.SrvConfig.OAuth2.Enabled {
		log.Println("OAuth2 authentication: ENABLED")
		InitOAuth2Server(apiconfig)
	} else {
		log.Println("OAuth2 authentication: DISABLED - API endpoints will be accessible without authentication")
	}

	router := gin.Default()

	// Improved CORS configuration
	router.Use(cors.New(cors.Config{
		AllowOrigins: []string{
			"http://localhost:3000",
			"http://localhost:9090",
			"http://127.0.0.1:3000",
			"http://127.0.0.1:9090",
			"https://localhost:9443",
			"https://localhost:3000",
			"https://127.0.0.1:9443",
			"https://127.0.0.1:3000",
		},
		AllowMethods:     []string{"GET", "POST", "PUT", "DELETE", "PATCH", "OPTIONS"},
		AllowHeaders:     []string{"Origin", "Content-Type", "Accept", "Authorization", "X-Requested-With", "Range"},
		ExposeHeaders:    []string{"Content-Length", "Content-Range"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	fsEmbed := EmbedFolder(dashboard, "dashboard/build", true)
	router.Use(Serve("/", fsEmbed))

	serverRoot, err := fs.Sub(dashboard, "dashboard/build")
	if err != nil {
		log.Fatal(err)
	}

	// OAuth2 endpoints (only functional when OAuth2 is enabled)
	router.POST("/oauth/token", TokenHandler)
	router.POST("/oauth/register", GetAuthMiddleware(apiconfig), RegisterOAuth2Client)

	// Public endpoints
	router.GET("/api/hello", hello)
	router.GET("/api/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "healthy", "version": "0.1.3"})
	})

	// API endpoints - Protected when OAuth2 is enabled, public when disabled
	api := router.Group("/api")
	api.Use(GetAuthMiddleware(apiconfig))
	{
		// User management
		api.GET("/users", ListUsers)
		api.GET("/users/search", SearchUsers)
		api.GET("/users/:username", GetUser)
		api.POST("/users", CreateUser)
		api.PUT("/users/:username", UpdateUser)
		api.DELETE("/users/:username", DeleteUser)

		// Group management
		api.GET("/groups", ListGroups)
		api.GET("/groups/search", SearchGroups)
		api.GET("/groups/:groupname", GetGroup)
		api.POST("/groups", CreateGroup)
		api.PUT("/groups/:groupname", UpdateGroup)
		api.DELETE("/groups/:groupname", DeleteGroup)

		// Group membership
		api.POST("/groups/:groupname/members", AddMemberToGroup)
		api.DELETE("/groups/:groupname/members/:username", RemoveMemberFromGroup)

		// Monitoring (legacy endpoint)
		api.GET("/monitor/0", func(ctx *gin.Context) {
			monitor(ctx, apiconfig)
		})
	}

	// Legacy auth endpoint (deprecated, use /oauth/token instead)
	router.POST("/auth", func(c *gin.Context) {
		err = basicAuth(c, apiconfig)
		if err != nil {
			auth(c)
		} else {
			auth(c)
		}
	})

	router.GET("/", func(c *gin.Context) {
		fmt.Printf("URL: %s\n", c.Request.URL.Path)
		c.FileFromFS("/index.html", http.FileSystem(http.FS(serverRoot)))
		c.AbortWithStatus(200)
	})

	router.NoRoute(func(c *gin.Context) {
		c.JSON(404, gin.H{
			"code": "PAGE_NOT_FOUND", "message": "Page not found",
		})
	})

	/* router.NoRoute(func (c *gin.Context) {
		fmt.Println("%s doesn't exists, redirect on /", c.Request.URL.Path)
		c.Redirect(http.StatusMovedPermanently, "/")
	}) */

	/* 	staticRoot, err := fs.Sub(dashboard, "dashboard/build")
	   	if err != nil {
	   		log.Fatal(err)
	   	} */

	//router.Use(AuthMiddleware())

	// Determine ports
	httpPort := ":9090"
	httpsPort := ":9443"
	if apiconfig.SrvConfig.ApiTls.Port != "" {
		httpsPort = ":" + apiconfig.SrvConfig.ApiTls.Port
	}

	// Start servers based on TLS configuration
	if apiconfig.SrvConfig.ApiTls.Enabled {
		// HTTPS is enabled
		if apiconfig.SrvConfig.ApiTls.CertFile == "" || apiconfig.SrvConfig.ApiTls.KeyFile == "" {
			log.Fatal("TLS enabled but certificate or key file not specified")
		}

		log.Printf("Starting HTTPS server on %s", httpsPort)

		if apiconfig.SrvConfig.ApiTls.AutoRedirect {
			// Start HTTP server that redirects to HTTPS
			go func() {
				redirectRouter := gin.New()
				redirectRouter.Use(gin.Logger())
				redirectRouter.Use(func(c *gin.Context) {
					httpsURL := "https://" + c.Request.Host + c.Request.RequestURI
					c.Redirect(http.StatusMovedPermanently, httpsURL)
				})
				log.Printf("Starting HTTP redirect server on %s", httpPort)
				if err := redirectRouter.Run(httpPort); err != nil {
					log.Printf("HTTP redirect server error: %v", err)
				}
			}()
		}

		// Start HTTPS server
		if err := router.RunTLS(httpsPort, apiconfig.SrvConfig.ApiTls.CertFile, apiconfig.SrvConfig.ApiTls.KeyFile); err != nil {
			log.Fatalf("Failed to start HTTPS server: %v", err)
		}
	} else {
		// HTTP only
		log.Printf("Starting HTTP server on %s (TLS disabled)", httpPort)
		router.Run(httpPort)
	}

}

func hello(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "world"})
}

type embedFileSystem struct {
	http.FileSystem
	indexes bool
}

func EmbedFolder(fsEmbed embed.FS, targetPath string, index bool) embedFileSystem {
	fmt.Printf("TargetPath: %s\n", targetPath)
	fmt.Printf("Index: %t\n", index)
	subFS, err := fs.Sub(fsEmbed, targetPath)
	if err != nil {
		panic(err)
	}
	return embedFileSystem{
		FileSystem: http.FS(subFS),
		indexes:    index,
	}
}

func (e embedFileSystem) Exists(prefix string, path string) bool {
	f, err := e.Open(path)
	if err != nil {
		return false
	}

	// check if indexing is allowed
	s, _ := f.Stat()
	if s.IsDir() && !e.indexes {
		return false
	}

	return true
}

func Serve(urlPrefix string, fs embedFileSystem) gin.HandlerFunc {
	fmt.Printf("URL: %s\n", urlPrefix)
	fmt.Printf("FS: %s\n", fs.FileSystem)
	fileserver := http.FileServer(fs)
	if urlPrefix != "" {
		fileserver = http.StripPrefix(urlPrefix, fileserver)
	}
	return func(c *gin.Context) {
		if fs.Exists(urlPrefix, c.Request.URL.Path) {
			fileserver.ServeHTTP(c.Writer, c.Request)
			c.Abort()
		}
	}
}
