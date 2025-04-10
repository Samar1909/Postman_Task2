package middleware

import (
	"context"
	"fmt"
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
	fmt.Println("in auth required middleware")
	// 1. Read JWT cookie
	tokenString, err := c.Cookie("Authorization")
	if err != nil {
		c.Redirect(http.StatusFound, "/")
		c.Abort()
		return
	}

	// 2. Parse JWT token
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (interface{}, error) {
		return []byte(os.Getenv("SECRET")), nil
	}, jwt.WithValidMethods([]string{jwt.SigningMethodHS256.Alg()}))
	if err != nil || token == nil {
		fmt.Println("1")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// 3. Extract claims
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		fmt.Println("2")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// 4. Check expiry
	if float64(time.Now().Unix()) > claims["exp"].(float64) {
		fmt.Println("3")

		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	// 5. Get user from DB
	userID := int32(claims["sub"].(float64)) // assuming sub is user ID
	queries := products.New(initializers.DB)
	reqUser, err := queries.GetUserByID(ctx, userID)
	if err != nil {
		fmt.Println("4")

		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}
	c.Set("user", reqUser)

	// 6. Load profile depending on role
	switch reqUser.RoleID.Int32 {
	case 2:
		profile, err := queries.GetRecruiterProfile(ctx, reqUser.UserID)
		if err != nil {
			fmt.Println(err.Error())

			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Set("userProfile", profile)

	case 3:
		profile, err := queries.GetApplicantProfile(ctx, reqUser.UserID)
		if err != nil {
			fmt.Println("6")

			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Set("userProfile", profile)
	}

	// 7. Let request continue
	c.Next()
}
