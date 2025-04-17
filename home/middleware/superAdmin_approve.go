package middleware

import (
	"context"
	"fmt"
	"home/initializers"
	"home/products"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
)

func SuperAdmin_approve(c *gin.Context) {
	fmt.Println("in the superAdmin_approve middleware")

	ctx := context.Background()
	queries := products.New(initializers.DB)
	user, exists := c.Get("user")
	if !exists {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	req_user, ok := user.(products.User)
	if !ok {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	req_userProfile, err := queries.GetRecruiterProfile(ctx, req_user.UserID)
	if err != nil {
		log.Fatal(err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if req_userProfile.Approved {
		c.Next()
		return
	}

	c.Redirect(http.StatusFound, "/recruiter/super_admin/accessRestricted")
}
