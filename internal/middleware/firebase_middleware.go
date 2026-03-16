package middleware

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	firebase "firebase.google.com/go"
	"firebase.google.com/go/auth"
	"github.com/gin-gonic/gin"
	"google.golang.org/api/option"
)

var FirebaseAuth *auth.Client

// Init FirebaseAuth
func InitFirebaseAuth() error {
	opt := option.WithAuthCredentialsFile(option.ServiceAccount, "internal/middleware/serviceAccountKey.json")
	app, err := firebase.NewApp(context.Background(), nil, opt)
	if err != nil {
		return fmt.Errorf("error initializing app: %v", err)
	}

	client, err := app.Auth(context.Background())
	if err != nil {
		return fmt.Errorf("error getting auth client %v", err)
	}

	FirebaseAuth = client
	fmt.Println("✅ Firebase Auth initialized")
	return nil
}

// FirebaseAuthMiddleware
func FirebaseAuthMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Authorization header is reqiored"})
			c.Abort()
			return
		}

		// retrived token from Bearer
		parts := strings.Split(authHeader, " ")
		if len(parts) != 2 || parts[0] != "Bearer" {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid authorization header format"})
			c.Abort()
			return
		}

		token := parts[1]
		// check token and Firebase
		tokenData, err := FirebaseAuth.VerifyIDToken(context.Background(), token)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "Invalid token"})
			c.Abort()
			return
		}

		// keep token data in the context for handler
		c.Set("user_id", tokenData.UID)
		c.Set("email", tokenData.Claims["email"])

		if name, ok := tokenData.Claims["name"]; ok {
			c.Set("name", name)
		}
		if picture, ok := tokenData.Claims["picture"]; ok {
			c.Set("picture", picture)
		}

		c.Next()
	}
}
