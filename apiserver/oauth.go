package apiserver

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/aescanero/openldap-node/config"
	"github.com/aescanero/openldap-node/ldaputils"
	"github.com/gin-gonic/gin"
	"github.com/go-redis/redis/v8"
	oauth2 "github.com/go-oauth2/oauth2/v4"
	"github.com/go-oauth2/oauth2/v4/errors"
	"github.com/go-oauth2/oauth2/v4/generates"
	"github.com/go-oauth2/oauth2/v4/manage"
	"github.com/go-oauth2/oauth2/v4/models"
	"github.com/go-oauth2/oauth2/v4/server"
	"github.com/go-oauth2/oauth2/v4/store"
	oredis "github.com/go-oauth2/redis/v4"
	"github.com/golang-jwt/jwt/v5"
)

var oauth2Server *server.Server
var apiConfig config.Config

// InitOAuth2Server initializes the OAuth2 server with LDAP authentication
func InitOAuth2Server(cfg config.Config) *server.Server {
	apiConfig = cfg

	manager := manage.NewDefaultManager()

	// Token store - Use Redis if enabled, otherwise fallback to memory
	if cfg.SrvConfig.Redis.Enabled {
		// Initialize Redis client
		redisHost := cfg.SrvConfig.Redis.Host
		if redisHost == "" {
			redisHost = "localhost"
		}
		redisPort := cfg.SrvConfig.Redis.Port
		if redisPort == "" {
			redisPort = "6379"
		}

		redisClient := redis.NewClient(&redis.Options{
			Addr:     fmt.Sprintf("%s:%s", redisHost, redisPort),
			Password: cfg.SrvConfig.Redis.Password,
			DB:       cfg.SrvConfig.Redis.DB,
		})

		// Test Redis connection
		ctx := context.Background()
		if err := redisClient.Ping(ctx).Err(); err != nil {
			log.Printf("Redis connection failed: %v. Falling back to memory store", err)
			manager.MustTokenStorage(store.NewMemoryTokenStore())
		} else {
			log.Printf("Redis token store connected successfully at %s:%s", redisHost, redisPort)
			tokenStore := oredis.NewRedisStore(&redis.Options{
				Addr:     fmt.Sprintf("%s:%s", redisHost, redisPort),
				Password: cfg.SrvConfig.Redis.Password,
				DB:       cfg.SrvConfig.Redis.DB,
			})
			manager.MapTokenStorage(tokenStore)
		}
	} else {
		log.Println("Using in-memory token store (Redis disabled)")
		manager.MustTokenStorage(store.NewMemoryTokenStore())
	}

	// JWT access token generator
	manager.MapAccessGenerate(generates.NewJWTAccessGenerate("", []byte("openldap-secret-key"), jwt.SigningMethodHS256))

	// Client store - Add default client for machine-to-machine
	clientStore := store.NewClientStore()
	clientStore.Set("openldap-client", &models.Client{
		ID:     "openldap-client",
		Secret: "openldap-secret",
		Domain: "http://localhost:9090",
	})
	manager.MapClientStorage(clientStore)

	// Configure token expiration
	manager.SetClientTokenCfg(&manage.Config{
		AccessTokenExp:    time.Hour * 2,
		RefreshTokenExp:   time.Hour * 24 * 7,
		IsGenerateRefresh: true,
	})

	// Password token expiration for user authentication
	manager.SetPasswordTokenCfg(&manage.Config{
		AccessTokenExp:    time.Hour * 2,
		RefreshTokenExp:   time.Hour * 24 * 7,
		IsGenerateRefresh: true,
	})

	srv := server.NewDefaultServer(manager)
	srv.SetAllowGetAccessRequest(true)
	srv.SetClientInfoHandler(server.ClientFormHandler)

	// LDAP-based password authentication
	srv.SetPasswordAuthorizationHandler(func(ctx context.Context, clientID, username, password string) (userID string, err error) {
		// Build user DN from username
		userDN := fmt.Sprintf("uid=%s,ou=users,%s", username, cfg.Database[0].Base)

		// Try to bind with LDAP
		conn, err := ldaputils.Connect(cfg, userDN, password)
		if err != nil {
			log.Printf("LDAP authentication failed for user %s: %v", username, err)
			return "", errors.ErrInvalidGrant
		}
		defer conn.Close()

		log.Printf("LDAP authentication successful for user: %s (clientID: %s)", username, clientID)
		return username, nil
	})

	// Client credentials authentication (machine-to-machine)
	srv.SetClientAuthorizedHandler(func(clientID string, grant oauth2.GrantType) (allowed bool, err error) {
		// Allow client_credentials grant type
		if grant == oauth2.ClientCredentials {
			return true, nil
		}
		return true, nil
	})

	srv.SetInternalErrorHandler(func(err error) (re *errors.Response) {
		log.Println("Internal Error:", err.Error())
		return
	})

	srv.SetResponseErrorHandler(func(re *errors.Response) {
		log.Println("Response Error:", re.Error.Error())
	})

	oauth2Server = srv
	return srv
}

// TokenHandler handles OAuth2 token requests
func TokenHandler(c *gin.Context) {
	if oauth2Server == nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "OAuth2 server not initialized"})
		return
	}

	err := oauth2Server.HandleTokenRequest(c.Writer, c.Request)
	if err != nil {
		log.Printf("Token request error: %v", err)
	}
}

// RegisterOAuth2Client registers a new OAuth2 client dynamically
func RegisterOAuth2Client(c *gin.Context) {
	var clientData struct {
		ClientID     string `json:"client_id" binding:"required"`
		ClientSecret string `json:"client_secret" binding:"required"`
		Domain       string `json:"domain" binding:"required"`
	}

	if err := c.ShouldBindJSON(&clientData); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	// This is a simplified version - in production, you'd want to store this in a database
	c.JSON(http.StatusCreated, gin.H{
		"message":   "Client registered successfully",
		"client_id": clientData.ClientID,
	})
}

// GetOAuth2Server returns the initialized OAuth2 server
func GetOAuth2Server() *server.Server {
	return oauth2Server
}
