package main

import (
	"context"
	"fmt"
	"home/controllers"
	"home/initializers"
	"home/middleware"
	"home/products"
	"net/http"

	"github.com/gin-gonic/gin"
)

func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectToDb()
}

func main() {
	r := gin.Default()
	r.LoadHTMLGlob("templates/*")
	r.GET("/", middleware.UnauthenticatedUser, middleware.CSRFMiddleware, func(c *gin.Context) {
		formToken, exists := c.Get("csrf_token")
		fmt.Println(formToken)
		if !exists {
			c.AbortWithStatus(http.StatusInternalServerError)
		}
		c.HTML(http.StatusOK, "signup.html", gin.H{
			"csrf_token": formToken,
		})
	})
	r.POST("/", controllers.SignUp)

	r.GET("/login", middleware.UnauthenticatedUser, middleware.CSRFMiddleware, func(c *gin.Context) {
		formToken, exists := c.Get("csrf_token")
		fmt.Println(formToken)
		if !exists {
			fmt.Println("Got a server error")
			c.AbortWithStatus(http.StatusInternalServerError)
		}
		signup_success, err := c.Cookie("signup_success")
		if err == nil && signup_success != "" {
			// Cookie exists and has a value
			c.HTML(http.StatusOK, "login.html", gin.H{
				"message":     signup_success,
				"messageType": "success",
				"csrf_token":  formToken,
			})

			// Clear the cookie after using it
			c.SetCookie("signup_success", "", -1, "/", "", false, false)
		} else {
			// No cookie or empty cookie - normal login page
			c.HTML(http.StatusOK, "login.html", gin.H{
				"csrf_token": formToken,
			})
		}
	})
	r.POST("/login", controllers.Login)
	r.GET("/super_admin/home", middleware.AuthRequired, middleware.AllowedGroups(1), controllers.Super_admin_home)

	r.GET("/recruiter/home/", middleware.AuthRequired, middleware.AllowedGroups(2), controllers.RecruiterHome)
	r.GET("/recruiter/updateProfile/", middleware.AuthRequired, middleware.AllowedGroups(2), middleware.CSRFMiddleware, func(c *gin.Context) {
		formToken, exists := c.Get("csrf_token")
		if !exists {
			c.AbortWithStatus(http.StatusInternalServerError)
		}
		user, exists := c.Get("user")
		if !exists {
			fmt.Println("1")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		req_user, ok := user.(products.User)
		if !ok {
			fmt.Println("2")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		userProfile, exists := c.Get("userProfile")
		if !exists {
			fmt.Println("3")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		req_userProfile, ok := userProfile.(products.RecruiterProfile)
		if !ok {
			fmt.Println("4")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		c.HTML(http.StatusOK, "recruiter_updateProfile.html", gin.H{
			"csrf_token":          formToken,
			"profile_active":      "active",
			"title":               "Update Profile",
			"email":               req_user.Email,
			"username":            req_user.Username,
			"company_name":        req_userProfile.CompanyName.String,
			"company_description": req_userProfile.CompanyDescription.String,
		})
	})

	r.POST("recruiter/updateProfile/", middleware.AuthRequired, middleware.AllowedGroups(2), controllers.RecruiterUpdateProfile)
	r.GET("/applicant/home", middleware.AuthRequired, middleware.AllowedGroups(3), controllers.ApplicantHome)
	r.GET("/applicant/updateProfile/", middleware.AuthRequired, middleware.AllowedGroups(3), middleware.CSRFMiddleware, func(c *gin.Context) {
		formToken, exists := c.Get("csrf_token")
		if !exists {
			c.AbortWithStatus(http.StatusInternalServerError)
		}
		user, exists := c.Get("user")
		if !exists {
			fmt.Println("1")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		req_user, ok := user.(products.User)
		if !ok {
			fmt.Println("2")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		userProfile, exists := c.Get("userProfile")
		if !exists {
			fmt.Println("3")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		req_userProfile, ok := userProfile.(products.ApplicantProfile)
		if !ok {
			fmt.Println("4")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		queries := products.New(initializers.DB)
		ctx := context.Background()

		req_userSkills, err := queries.GetApplicantSkills(ctx, req_user.UserID)
		if err != nil {
			c.JSON(http.StatusUnauthorized, gin.H{
				"error": "Failed geting user skills",
			})
		}
		fmt.Println(req_userProfile.School.String)

		c.HTML(http.StatusFound, "applicant_updateProfile.html", gin.H{
			"csrf_token":     formToken,
			"profile_active": "active",
			"title":          "Update Profile",
			"email":          req_user.Email,
			"username":       req_user.Username,
			"first_name":     req_userProfile.FirstName.String,
			"last_name":      req_userProfile.LastName.String,
			"school":         req_userProfile.School.String,
			"college":        req_userProfile.College.String,
			"age":            req_userProfile.Age.Int32,
			"UserSkills":     req_userSkills,
		})
	})

	r.POST("/applicant/updateProfile/", middleware.AuthRequired, middleware.AllowedGroups(3), controllers.ApplicantUpdateProfile)

	r.GET("/applicant/updateProfile/searchSkills", middleware.AuthRequired, middleware.AllowedGroups(3), controllers.ApplicantSearchSkills)
	r.POST("/applicant/updateProfile/addSkill", middleware.AuthRequired, middleware.AllowedGroups(3), controllers.ApplicantAddSkill)
	r.GET("/applicant/updateProfile/deleteSkill/:skill_id", middleware.AuthRequired, middleware.AllowedGroups(3), controllers.ApplicantDeleteSkill)
	r.GET("/applicant/resume/", middleware.AuthRequired, middleware.AllowedGroups(3), middleware.CSRFMiddleware, func(c *gin.Context) {
		formToken, exists := c.Get("csrf_token")
		if !exists {
			c.AbortWithStatus(http.StatusInternalServerError)
		}
		user, exists := c.Get("user")
		if !exists {
			fmt.Println("1")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		req_user, ok := user.(products.User)
		if !ok {
			fmt.Println("2")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}

		userProfile, exists := c.Get("userProfile")
		if !exists {
			fmt.Println("3")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		req_userProfile, ok := userProfile.(products.ApplicantProfile)
		if !ok {
			fmt.Println("4")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		var message string
		var messageType string
		resume_validationSuccess, err := c.Cookie("resume_validationSuccess")
		if err == nil && resume_validationSuccess != "" {
			message = resume_validationSuccess
			messageType = "success"
		}
		resume_validationFail, err := c.Cookie("resume_validationFail")
		if err == nil && resume_validationFail != "" {
			message = resume_validationFail
			messageType = "danger"
		}
		c.HTML(http.StatusOK, "applicant_resume.html", gin.H{
			"csrf_token":     formToken,
			"profile_active": "active",
			"title":          "Your Resume",
			"User":           req_user,
			"UserProfile":    req_userProfile,
			"message":        message,
			"messageType":    messageType,
		})
		c.SetCookie("resume_validationSuccess", "", -1, "/", "", false, false)
		c.SetCookie("resume_validationFail", "", -1, "/", "", false, false)

	})

	r.POST("/applicant/resume/upload/", middleware.AuthRequired, middleware.AllowedGroups(3), controllers.ApplicantResumeUpload)

	r.GET("logout/", middleware.AuthRequired, middleware.CSRFMiddleware, controllers.LogOut)

	r.Run()
}
