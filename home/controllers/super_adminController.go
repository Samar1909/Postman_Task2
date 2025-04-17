package controllers

import (
	"context"
	"fmt"
	"home/initializers"
	"home/products"
	"log"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

func Super_admin_home(c *gin.Context) {
	ctx := context.Background()
	queries := products.New(initializers.DB)
	user, exists := c.Get("user")
	if !exists {
		c.AbortWithStatus(http.StatusUnauthorized)
	}
	req_user := user.(products.User)

	results, err := queries.GetRestrictedUsers(ctx)
	if err != nil {
		log.Fatal(err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
	}
	fmt.Println(req_user.Username)
	c.HTML(http.StatusOK, "super_admin_home.html", gin.H{
		"username": "hello",
		"results":  results,
	})
}

func SuperAdmin_recruiterProfile(c *gin.Context) {
	ctx := context.Background()
	queries := products.New(initializers.DB)

	user_id := c.Param("user_id")
	user_idInt, err := strconv.Atoi(user_id)
	if err != nil {
		log.Fatal(err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
	}
	user, err := queries.GetUserByID(ctx, int32(user_idInt))
	if err != nil {
		log.Fatal(err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	userProfile, err := queries.GetRecruiterProfile(ctx, user.UserID)
	if err != nil {
		log.Fatal(err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	c.HTML(http.StatusOK, "SuperAdmin_recruiterProfile.html", gin.H{
		"user":        user,
		"userProfile": userProfile,
		"title":       fmt.Sprintf("%s Profile", user.Username),
	})
}

func SuperAdmin_recruiterApprove(c *gin.Context) {
	ctx := context.Background()
	queries := products.New(initializers.DB)

	user_id := c.Param("user_id")
	user_idInt, err := strconv.Atoi(user_id)
	if err != nil {
		log.Fatal(err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
	}
	if err = queries.ApproveRecruiter(ctx, int32(user_idInt)); err != nil {
		log.Fatal(err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
	}
	c.Redirect(http.StatusFound, "/super_admin/home")
}

func SuperAdmin_recruiterDecline(c *gin.Context) {
	ctx := context.Background()
	queries := products.New(initializers.DB)

	user_id := c.Param("user_id")
	user_idInt, err := strconv.Atoi(user_id)
	if err != nil {
		log.Fatal(err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
	}
	if err = queries.DeclineRecruiter(ctx, int32(user_idInt)); err != nil {
		log.Fatal(err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
	}
	c.Redirect(http.StatusFound, "/super_admin/home")
}
