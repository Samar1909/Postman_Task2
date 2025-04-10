package middleware

import (
	"crypto/rand"
	"encoding/base64"
	"fmt"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func generateRandomToken(length int) (string, error) {
	bytes := make([]byte, length)
	_, err := rand.Read(bytes)
	if err != nil {
		log.Println("Unable to generate csrf token: ", err)
	}
	return base64.URLEncoding.EncodeToString(bytes), nil
}

func CSRFMiddleware(c *gin.Context) {
	fmt.Println("in the csrf middleware")
	csrf_token, err := c.Cookie("CSRF_Token")
	fmt.Println("middleware wala = ", csrf_token)
	if err != nil {
		fmt.Println("New cookie")
		csrfToken, err := generateRandomToken(32)
		if err != nil {
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     "CSRF_Token",
			Value:    csrfToken,
			Quoted:   true,
			Path:     "/",
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteStrictMode,
		})
		fmt.Println(csrfToken)
		c.Set("csrf_token", csrfToken)
		c.Next()
	}
	c.Set("csrf_token", csrf_token)
	c.Next()
}
