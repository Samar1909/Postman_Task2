package controllers

import (
	"fmt"
	"home/products"
	"net/http"

	"github.com/gin-gonic/gin"
)

func Super_admin_home(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.AbortWithStatus(http.StatusUnauthorized)
	}
	req_user := user.(products.User)
	fmt.Println(req_user.Username)
	c.HTML(http.StatusOK, "super_admin_home.html", gin.H{
		"username": "hello",
	})
}
