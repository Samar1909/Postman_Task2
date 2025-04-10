package controllers

import (
	"context"
	"database/sql"
	"fmt"
	"home/initializers"
	"home/products"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

func RecruiterHome(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.AbortWithStatus(http.StatusUnauthorized)
	}
	req_user := user.(products.User)
	fmt.Println(req_user.Username)
	c.HTML(http.StatusOK, "recruiter_home.html", gin.H{
		"username": req_user.Username,
	})
}

func RecruiterUpdateProfile(c *gin.Context) {
	ctx := context.Background()
	queries := products.New(initializers.DB)

	//get the user variable
	user, exists := c.Get("user")
	if !exists {
		fmt.Println("1")
		c.AbortWithStatus(http.StatusUnauthorized)
	}
	req_user := user.(products.User)

	//csrf verification
	csrf_token, err := c.Cookie("CSRF_Token")

	if err != nil {
		fmt.Println("3")
		c.AbortWithStatus(http.StatusUnauthorized)
	}
	formToken := c.PostForm("csrf_token")
	fmt.Println(csrf_token)
	fmt.Println(formToken)
	if csrf_token != formToken {
		fmt.Println("4")
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	http.SetCookie(c.Writer, &http.Cookie{
		Name:     "CSRF_Token",
		Value:    "",
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
		SameSite: http.SameSiteStrictMode,
		Expires:  time.Unix(0, 0),
		MaxAge:   -1,
	})

	//data retrieval and validation
	Email := c.PostForm("email")
	Username := c.PostForm("username")
	CompanyName := c.PostForm("company_name")
	CompanyDescription := c.PostForm("company_description")

	if Email == "" || Username == "" || CompanyName == "" || CompanyDescription == "" {
		c.HTML(http.StatusFound, "recruiter_updateProfile.html", gin.H{
			"csrf_token":          formToken,
			"email":               Email,
			"username":            Username,
			"company_name":        CompanyName,
			"company_description": CompanyDescription,
			"message":             "No field can be left blank",
			"messageType":         "danger",
		})
		return
	}

	var email_valid bool = false

	//validating unique email constraint
	if Email != req_user.Email {
		_, err := queries.GetUserByEmail(ctx, Email)
		if err != nil {
			//No user found with the given email
			email_valid = true
		} else {
			c.HTML(http.StatusFound, "recruiter_updateProfile.html", gin.H{
				"csrf_token":          formToken,
				"email":               Email,
				"username":            Username,
				"company_name":        CompanyName,
				"company_description": CompanyDescription,
				"message":             "A user with this email already exists",
				"messageType":         "danger",
			})
			return
		}
	} else {
		email_valid = true
	}
	if email_valid {
		err = queries.UpdateUser(ctx, products.UpdateUserParams{
			Email:    Email,
			Username: Username,
			UserID:   req_user.UserID,
		})
		if err != nil {
			c.HTML(http.StatusFound, "recruiter_updateProfile.html", gin.H{
				"csrf_token":          formToken,
				"email":               Email,
				"username":            Username,
				"company_name":        CompanyName,
				"company_description": CompanyDescription,
				"message":             err.Error(),
				"messageType":         "danger",
			})
			return
		}

		err = queries.UpdateRecruiterProfile(ctx, products.UpdateRecruiterProfileParams{
			CompanyName:        sql.NullString{String: CompanyName, Valid: CompanyName != ""},
			CompanyDescription: sql.NullString{String: CompanyDescription, Valid: CompanyDescription != ""},
		})
		if err != nil {
			c.HTML(http.StatusFound, "recruiter_updateProfile.html", gin.H{
				"csrf_token":          formToken,
				"email":               Email,
				"username":            Username,
				"company_name":        CompanyName,
				"company_description": CompanyDescription,
				"message":             err.Error(),
				"messageType":         "danger",
			})
			return
		}
		fmt.Println("Profile Updated")
		c.Redirect(http.StatusFound, "/recruiter/home")
		return
	}
}

func RecruiterNewJobPosting(c *gin.Context) {

}
