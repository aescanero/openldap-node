package apiserver

import (
	"errors"
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/aescanero/openldap-node/config"
	"github.com/aescanero/openldap-node/ldaputils"
	"github.com/gin-gonic/gin"
	"github.com/go-ldap/ldap/v3"
	"github.com/golang-jwt/jwt"
)

type ldapConn struct {
	conn     *ldap.Conn
	username string
	ts       int64
	token    string
}

type poolConn struct {
	pool []*ldapConn
}

var pc poolConn

func poolMonitor(apiconfig config.Config, stateError chan error) {
	now := time.Now().Unix()
	time.Sleep(1000 * time.Millisecond)
	var newpc poolConn
	var needChange = false
	for _, conn := range pc.pool {
		ts := conn.ts
		if now-ts > 300 {
			needChange = true
		} else {
			conn.conn.Close()
			newpc.pool = append(newpc.pool, conn)
		}
	}

	if needChange {
		pc = newpc
	}
}

func auth(ctx *gin.Context) error {

	username, _, _ := ctx.Request.BasicAuth()
	//username := ctx.PostForm("username")
	//password := c.PostForm("password")
	fmt.Printf("func auth: username: %s\n", username)
	claims := jwt.MapClaims{
		"username": username,
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, _ := token.SignedString([]byte("world"))
	fmt.Printf("func auth: token: %s\n", tokenString)
	usernameFound := false
	for _, conn := range pc.pool {
		if conn.username == username {
			usernameFound = true
			fmt.Printf("func auth: username found: %s\n", username)
			conn.ts = time.Now().Unix()
			conn.token = tokenString
		}
	}

	if !usernameFound {
		ctx.Abort()
		ctx.Writer.WriteHeader(http.StatusUnauthorized)
		return errors.New("WWW-Authenticate")
	}

	ctx.JSON(http.StatusOK, gin.H{"token": tokenString})
	return nil
}

func basicAuth(c *gin.Context, apiconfig config.Config) error {
	now := time.Now().Unix()

	user, password, ok := c.Request.BasicAuth()
	fmt.Printf("func basicAuth: user: %s\n", user)
	if !ok {
		c.Abort()
		c.Writer.WriteHeader(http.StatusUnauthorized)
		return errors.New("WWW-Authenticate")
	} else {
		lc, err := ldaputils.Connect(apiconfig, user, password)
		if err != nil {
			c.Abort()
			c.Writer.WriteHeader(http.StatusUnauthorized)
			return errors.New("WWW-Authenticate")
		}

		usernameFound := false
		for _, conn := range pc.pool {
			if conn.username == user {
				fmt.Printf("func basicAuth: username found: %s\n", user)
				usernameFound = true
				conn.ts = time.Now().Unix()
				conn.conn.Close()
				conn.conn = lc
			}
		}

		if !usernameFound {
			fmt.Printf("func basicAuth: username not found: %s\n", user)
			pc.pool = append(pc.pool, &ldapConn{
				conn:     lc,
				username: user,
				ts:       now,
				token:    "",
			})
		}
	}

	return nil
}

func AuthMiddleware() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		authString := ctx.GetHeader("Authorization")
		splitToken := strings.Split(authString, "Bearer ")
		tokenString := splitToken[1]
		fmt.Printf("Token: %s\n", tokenString)
		if tokenString == "" {
			ctx.JSON(http.StatusUnauthorized, gin.H{"message": "Authorization token not found"})
			ctx.Abort()
			return
		}

		token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
			return []byte("world"), nil
		})

		if err != nil || !token.Valid {
			ctx.JSON(http.StatusUnauthorized, gin.H{"message": "Invalid token"})
			ctx.Abort()
			return
		}

		tokenFound := false
		for _, conn := range pc.pool {
			if conn.token == tokenString {
				tokenFound = true
				conn.ts = time.Now().Unix()
			}
		}

		if !tokenFound {
			ctx.JSON(http.StatusUnauthorized, gin.H{"message": "Authorization token not found"})
			ctx.Abort()
			return
		}

		ctx.Next()
	}
}

func ProtectedHandler(ctx *gin.Context) {
	ctx.JSON(http.StatusOK, gin.H{"message": "Protected endpoint accessed"})
}
