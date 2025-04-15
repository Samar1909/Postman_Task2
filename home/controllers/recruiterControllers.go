package controllers

import (
	"context"
	"database/sql"
	"fmt"
	"home/initializers"
	"home/products"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
)

func RecruiterHome(c *gin.Context) {
	user, exists := c.Get("user")
	if !exists {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	req_user := user.(products.User)

	c.HTML(http.StatusOK, "recruiter_home.html", gin.H{
		"username":    req_user.Username,
		"home_active": "active",
		"title":       "Home",
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
			"profile_active":      "active",
			"title":               "Update Profile",
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
				"profile_active":      "active",
				"title":               "Update Profile",
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
				"profile_active":      "active",
				"title":               "Update Profile",
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
			UserID:             req_user.UserID,
		})
		if err != nil {
			c.HTML(http.StatusFound, "recruiter_updateProfile.html", gin.H{
				"csrf_token":          formToken,
				"profile_active":      "active",
				"title":               "Update Profile",
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

func RecruiterAddSkill(c *gin.Context) {
	ctx := context.Background()
	queries := products.New(initializers.DB)

	posting_id := c.PostForm("posting_id")
	fmt.Println(posting_id)
	posting_idInt, err := strconv.Atoi(posting_id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}

	skill := c.PostForm("search-input")
	skill_obj, err := queries.GetSkill(ctx, sql.NullString{Valid: true, String: skill})
	if err != nil {
		skill_newObj, err := queries.CreateSkill(ctx, sql.NullString{Valid: true, String: skill})
		if err != nil {
			fmt.Println("2")
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		queries.CreateSkill_Req(ctx, products.CreateSkill_ReqParams{
			PostingID: int32(posting_idInt),
			SkillID:   skill_newObj.SkillID,
		})

	} else {
		_, err := queries.GetApplicantSkill(ctx, skill_obj.SkillID)
		if err != nil {
			queries.CreateSkill_Req(ctx, products.CreateSkill_ReqParams{
				PostingID: int32(posting_idInt),
				SkillID:   skill_obj.SkillID,
			})
		}
	}
	c.SetCookie("posting_id", posting_id, 5, "/", "", false, true)
	c.Redirect(http.StatusFound, "/recruiter/jobPosting/create")
}

func RecruiterDeleteSkill(c *gin.Context) {
	ctx := context.Background()
	queries := products.New(initializers.DB)

	posting_id := c.Param("posting_id")
	skill_id := c.Param("skill_id")

	skill_id_int, err := strconv.Atoi(skill_id)
	if err != nil {
		fmt.Println("2")
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": err.Error(),
		})
		return
	}
	posting_id_int, err := strconv.Atoi(posting_id)
	if err != nil {
		fmt.Println("2")
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": err.Error(),
		})
		return
	}

	err = queries.DeleteRequiredSkill(ctx, products.DeleteRequiredSkillParams{
		SkillID:   int32(skill_id_int),
		PostingID: int32(posting_id_int),
	})

	if err != nil {
		fmt.Println("3")
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": err.Error(),
		})
		return
	}
	c.SetCookie("posting_id", posting_id, 5, "/", "", false, true)
	c.Redirect(http.StatusFound, "/recruiter/jobPosting/create")
}

func RecruiterNewJobPosting(c *gin.Context) {
	ctx := context.Background()
	queries := products.New(initializers.DB)

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

	posting_id := c.PostForm("posting_id")
	JobTitle := c.PostForm("job_title")
	JobDescription := c.PostForm("job_description")
	fmt.Println(JobTitle)
	fmt.Println(JobDescription)
	fmt.Println(posting_id)
	posting_idInt, err := strconv.Atoi(posting_id)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
			"error": err.Error(),
		})
		return
	}

	new_jobPosting, err := queries.UpdateJobPosting(ctx, products.UpdateJobPostingParams{
		PostingID:      int32(posting_idInt),
		JobTitle:       sql.NullString{Valid: true, String: JobTitle},
		JobDescription: sql.NullString{Valid: true, String: JobDescription},
	})
	fmt.Println(new_jobPosting.JobTitle.String)
	fmt.Println(new_jobPosting.JobDescription.String)

	if err != nil {
		log.Fatal(err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	err = queries.DeleteJobPosting(ctx)
	if err != nil {
		log.Fatal(err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
	}

	c.Redirect(http.StatusFound, "/recruiter/home/")

}
