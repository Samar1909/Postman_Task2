package middleware

import (
	"context"
	"home/initializers"
	"home/products"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
)

func AuthRequired(c *gin.Context) {
	ctx := context.Background()
	tokenString, err := c.Cookie("Authorization")
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
	}
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("SECRET")), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
	}

	if claims, ok := token.Claims.(jwt.MapClaims); ok {
		if float64(time.Now().Unix()) > claims["exp"].(float64) {
			c.AbortWithStatus(http.StatusUnauthorized)
		} else {
			queries := products.New(initializers.DB)
			req_user, err := queries.GetUserByID(ctx, int32(claims["sub"].(float64)))
			if err != nil {
				c.AbortWithStatus(http.StatusUnauthorized)
			}
			c.Set("user", req_user)
			c.Next()
		}
	} else {
		c.AbortWithStatus(http.StatusUnauthorized)
	}

}
