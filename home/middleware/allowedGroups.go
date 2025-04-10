package middleware

import (
	"home/products"
	"net/http"

	"github.com/gin-gonic/gin"
)

func AllowedGroups(reqRoleID int) gin.HandlerFunc {
	return func(c *gin.Context) {
		user, exists := c.Get("user")
		if !exists {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		authUser, ok := user.(products.User)
		if !ok {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		if authUser.RoleID.Valid && authUser.RoleID.Int32 == 1 {
			c.Next() //since super admin has full control over the site
		}
		if !authUser.RoleID.Valid || authUser.RoleID.Int32 != int32(reqRoleID) {
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.Next()
	}

}
