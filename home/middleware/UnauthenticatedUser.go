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

func UnauthenticatedUser(c *gin.Context) {
	ctx := context.Background()
	tokenString, err := c.Cookie("Authorization")
	if err != nil {
		c.Next()
		return
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
			if req_user.RoleID.Valid {
				switch req_user.RoleID.Int32 {
				case 1:
					c.Redirect(http.StatusFound, "/super_admin/home")
					c.Abort()
					return
				case 2:
					c.Redirect(http.StatusFound, "/recruiter/updateProfile")
					c.Abort()
					return
				case 3:
					c.Redirect(http.StatusFound, "/applicant/home")
					c.Abort()
					return
				default:
					c.AbortWithStatus(http.StatusUnauthorized)
					return

				}
			} else {
				c.AbortWithStatus(http.StatusUnauthorized)
			}

		}
	} else {
		c.AbortWithStatus(http.StatusUnauthorized)
	}
}
