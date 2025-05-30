package main

import (
	"context"
	"database/sql"
	"fmt"
	"home/controllers"
	"home/initializers"
	"home/middleware"
	"home/products"
	"html/template"
	"log"
	"net/http"
	"os"
	"strconv"

	"github.com/gin-gonic/gin"
)

func init() {
	initializers.LoadEnvVariables()
	initializers.ConnectToDb()
}

func main() {
	controllers.InitGoogleOauth()
	r := gin.Default()
	fmt.Println("Client ID:", os.Getenv("OAUTH_CLIENT_ID"))
	fmt.Println("Client Secret:", os.Getenv("OAUTH_CLIENT_SECRET"))
	r.LoadHTMLGlob("templates/*")
	r.GET("/", middleware.UnauthenticatedUser, middleware.CSRFMiddleware, func(c *gin.Context) {
		fmt.Println("in  login now")
		fmt.Println("Client ID:", os.Getenv("OAUTH_CLIENT_ID"))
		fmt.Println("Client Secret:", os.Getenv("OAUTH_CLIENT_SECRET"))
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
	r.GET("/auth/google/login", controllers.HandleGoogleLogin)
	r.GET("/auth/google/callback", controllers.HandleGoogleCallback)
	r.GET("/users/:user_id/group", middleware.AuthRequired, middleware.CSRFMiddleware, func(c *gin.Context) {
		formToken, exists := c.Get("csrf_token")
		fmt.Println(formToken)
		if !exists {
			fmt.Println("Got a server error")
			c.AbortWithStatus(http.StatusInternalServerError)
		}
		c.HTML(http.StatusOK, "groups.html", gin.H{
			"csrf_token": formToken,
		})
	})
	r.POST("/users/:user_id/group", middleware.AuthRequired, controllers.UserGroups)

	//superAdmin Endpoints
	r.GET("/super_admin/home", middleware.AuthRequired, middleware.AllowedGroups(1), controllers.Super_admin_home)
	r.GET("/super_admin/:user_id/profile", middleware.AuthRequired, middleware.AllowedGroups(1), controllers.SuperAdmin_recruiterProfile)
	r.GET("/super_admin/:user_id/approve", middleware.AuthRequired, middleware.AllowedGroups(1), controllers.SuperAdmin_recruiterApprove)
	r.GET("/super_admin/:user_id/decline", middleware.AuthRequired, middleware.AllowedGroups(1), controllers.SuperAdmin_recruiterDecline)

	// Recruiters Endpoints
	r.GET("/recruiter/super_admin/accessRestricted", middleware.AuthRequired, middleware.AllowedGroups(2), func(c *gin.Context) {
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

		results, err := queries.GetUsersByRoleID(ctx, 1)
		if err != nil {
			log.Fatal(err.Error())
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		c.HTML(http.StatusOK, "recruiter_approve.html", gin.H{
			"results":  results,
			"title":    "Not Approved User",
			"username": req_user.Username,
		})

	})
	r.GET("/recruiter/home/", middleware.AuthRequired, middleware.AllowedGroups(2), middleware.SuperAdmin_approve, controllers.RecruiterHome)
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

	r.POST("recruiter/updateProfile/", middleware.AuthRequired, middleware.AllowedGroups(2), middleware.SuperAdmin_approve, controllers.RecruiterUpdateProfile)
	r.GET("/recruiter/jobPosting/create", middleware.AuthRequired, middleware.AllowedGroups(2), middleware.SuperAdmin_approve, middleware.CSRFMiddleware, func(c *gin.Context) {
		formToken, exists := c.Get("csrf_token")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "CSRF verification failed",
			})
			return
		}
		user, exists := c.Get("user")
		if !exists {
			fmt.Println("1")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "CSRF verification failed",
			})
			return
		}
		req_user, ok := user.(products.User)
		if !ok {
			fmt.Println("2")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "CSRF verification failed",
			})
			return
		}
		queries := products.New(initializers.DB)
		ctx := context.Background()

		posting_id, err := c.Cookie("posting_id")
		fmt.Println("post_id", posting_id)

		var job_title string
		var job_description string

		var job_posting products.JobPosting

		if err == nil && posting_id != "" {
			// Cookie exists and has a value
			posting_idInt, err := strconv.Atoi(posting_id)
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": err.Error(),
				})
				return
			}
			job_title, err = c.Cookie("job_title")
			job_description, err = c.Cookie("job_description")
			job_posting, err = queries.GetJobPosting(ctx, int32(posting_idInt))
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to create new Job Posting. Pls try again leter",
				})
				return
			}

			// Clear the cookie after using it
			c.SetCookie("job_title", "", -1, "/", "", false, false)
			c.SetCookie("job_description", "", -1, "/", "", false, false)
			c.SetCookie("posting_id", "", -1, "/", "", false, false)

		} else {
			// No cookie or empty cookie - normal login page
			job_posting, err = queries.CreateJobPosting(ctx, sql.NullInt32{Valid: true, Int32: req_user.UserID})
			if err != nil {
				c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
					"error": "Failed to create new Job Posting. Pls try again leter",
				})
				return
			}
		}

		skills_req, err := queries.GetRequiredSkills(ctx, job_posting.PostingID)
		if err != nil {
			c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to get required skills. Pls try again leter",
			})
			return
		}

		c.HTML(http.StatusOK, "recruiter_createJobPosting.html", gin.H{
			"csrf_token":      formToken,
			"profile_active":  "add_job_active",
			"title":           "New Job Posting",
			"posting_id":      job_posting.PostingID,
			"job_title":       job_title,
			"job_description": job_description,
			"skills_req":      skills_req,
		})
	})

	r.GET("/recruiter/jobPosting/searchSkills", middleware.AuthRequired, middleware.AllowedGroups(2), middleware.SuperAdmin_approve, controllers.ApplicantSearchSkills)
	r.POST("/recruiter/jobPosting/addSkill/", middleware.AuthRequired, middleware.AllowedGroups(2), middleware.SuperAdmin_approve, controllers.RecruiterAddSkill)
	r.GET("/recruiter/jobPosting/deleteSkill/:skill_id/:posting_id", middleware.AuthRequired, middleware.AllowedGroups(2), middleware.SuperAdmin_approve, controllers.RecruiterDeleteSkill)
	r.POST("/recruiter/jobPosting/create", middleware.AuthRequired, middleware.AllowedGroups(2), middleware.SuperAdmin_approve, controllers.RecruiterNewJobPosting)
	r.GET("/recruiter/jobPostings/all", middleware.AuthRequired, middleware.AllowedGroups(2), middleware.SuperAdmin_approve, func(c *gin.Context) {
		queries := products.New(initializers.DB)
		ctx := context.Background()
		user, exists := c.Get("user")
		if !exists {
			fmt.Println("1")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "CSRF verification failed",
			})
			return
		}
		req_user, ok := user.(products.User)
		if !ok {
			fmt.Println("2")
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "CSRF verification failed",
			})
			return
		}

		jobPostings, err := queries.GetJobPosting_recruiters(ctx, req_user.UserID)
		if err != nil {
			log.Fatal(err.Error())
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		c.HTML(http.StatusOK, "recruiter_jobPostingList.html", gin.H{
			"job_postings":    jobPostings,
			"job_list_active": "active",
			"title":           "Job Postings",
		})
	})

	r.GET("/recruiter/jobPostings/:posting_id", middleware.AuthRequired, middleware.AllowedGroups(2), middleware.SuperAdmin_approve, func(c *gin.Context) {
		queries := products.New(initializers.DB)
		ctx := context.Background()
		posting_id := c.Param("posting_id")
		posting_idInt, err := strconv.Atoi(posting_id)
		if err != nil {
			log.Fatal(err.Error())
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		job_posting, err := queries.GetJobPosting(ctx, int32(posting_idInt))
		if err != nil {
			log.Fatal(err.Error())
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		recruiter_profile, err := queries.GetRecruiterProfile(ctx, job_posting.UserID.Int32)
		if err != nil {
			log.Fatal(err.Error())
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		skills_req, err := queries.GetRequiredSkills(ctx, job_posting.PostingID)
		if err != nil {
			log.Fatal(err.Error())
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		recruiter_user, err := queries.GetUserByID(ctx, job_posting.UserID.Int32)
		if err != nil {
			log.Fatal(err.Error())
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		var postingDate string

		if job_posting.PostingDate.Valid {
			postingDate = job_posting.PostingDate.Time.Format("2006-01-02")
			fmt.Println("Posting Date:", postingDate)
		} else {
			log.Fatal("Posting date is null")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.HTML(http.StatusOK, "recruiter_jobPostingDetail.html", gin.H{
			"title":             job_posting.JobTitle.String,
			"job_posting":       job_posting,
			"recruiter_profile": recruiter_profile,
			"skills_req":        skills_req,
			"recruiter_user":    recruiter_user,
			"posting_date":      postingDate,
		})

	})
	r.GET("/recruiter/jobPostings/:posting_id/applicants", middleware.AuthRequired, middleware.AllowedGroups(2), middleware.SuperAdmin_approve, func(c *gin.Context) {
		queries := products.New(initializers.DB)
		ctx := context.Background()
		posting_id := c.Param("posting_id")
		posting_idInt, err := strconv.Atoi(posting_id)
		if err != nil {
			log.Fatal(err.Error())
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		job_posting, err := queries.GetJobPosting(ctx, int32(posting_idInt))
		if err != nil {
			log.Fatal(err.Error())
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		results, err := queries.GetJobApplicants(ctx, int32(posting_idInt))
		if err != nil {
			log.Fatal(err.Error())
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		c.HTML(http.StatusOK, "recruiter_jobApplicantsList.html", gin.H{
			"title":     fmt.Sprintf("Job Applicants for %s", job_posting.JobTitle.String),
			"results":   results,
			"Job_Title": job_posting.JobTitle,
		})
	})
	r.GET("/recruiter/jobPostings/:posting_id/applicants/:user_id", middleware.AuthRequired, middleware.AllowedGroups(2), middleware.SuperAdmin_approve, func(c *gin.Context) {
		queries := products.New(initializers.DB)
		ctx := context.Background()
		posting_id := c.Param("posting_id")
		user_id := c.Param("user_id")
		user_idInt, err := strconv.Atoi(user_id)
		if err != nil {
			log.Fatal(err.Error())
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		user, err := queries.GetUserByID(ctx, int32(user_idInt))
		if err != nil {
			log.Fatal(err.Error())
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		userProfile, err := queries.GetApplicantProfile(ctx, int32(user_idInt))
		if err != nil {
			log.Fatal(err.Error())
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		userSkills, err := queries.GetApplicantSkills(ctx, user.UserID)
		if err != nil {
			log.Fatal(err.Error())
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		resume_parse, err := c.Cookie("resume_parse")
		var resume_parseText template.HTML

		if err == nil && resume_parse != "" {
			// Cookie exists and has a value
			resume_parseText = template.HTML(resume_parse)
			// Clear the cookie after using it
			c.SetCookie("resume_parse", "", -1, "/", "", false, false)

		}

		interview, err := c.Cookie("interview")

		if err == nil && interview != "" {
			// Cookie exists and has a value
			c.HTML(http.StatusOK, "recruiter_jobApplicantsDetail.html", gin.H{
				"title":        user.Username,
				"user":         user,
				"userProfile":  userProfile,
				"posting_id":   posting_id,
				"userSkills":   userSkills,
				"resume_parse": resume_parseText,
				"message":      interview,
				"messageType":  "warning",
			})
			// Clear the cookie after using it
			c.SetCookie("resume_parse", "", -1, "/", "", false, false)
			return
		}
		c.HTML(http.StatusOK, "recruiter_jobApplicantsDetail.html", gin.H{
			"title":        user.Username,
			"user":         user,
			"userProfile":  userProfile,
			"posting_id":   posting_id,
			"userSkills":   userSkills,
			"resume_parse": resume_parseText,
		})
	})
	r.GET("/recruiter/jobPostings/:posting_id/applicants/:user_id/resume/download", middleware.AuthRequired, middleware.AllowedGroups(2), middleware.SuperAdmin_approve, func(c *gin.Context) {
		queries := products.New(initializers.DB)
		ctx := context.Background()
		posting_id := c.Param("posting_id")
		user_id := c.Param("user_id")
		user_idInt, err := strconv.Atoi(user_id)
		if err != nil {
			log.Fatal(err.Error())
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		fileName, err := queries.GetApplicantResume(ctx, int32(user_idInt))
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to fetch file from database",
			})
			return
		}

		filePath := "resume/" + fileName.String
		c.FileAttachment(filePath, fileName.String)
		redirect_url := fmt.Sprintf("/recruiter/jobPostings/%s/applicants/%s", posting_id, user_id)
		c.Redirect(http.StatusFound, redirect_url)
	})

	r.GET("/recruiter/jobPostings/:posting_id/applicants/:user_id/resume/extract", middleware.AuthRequired, middleware.AllowedGroups(2), middleware.SuperAdmin_approve, controllers.RecruiterResumeImport)
	r.GET("/recruiter/jobPostings/:posting_id/applicants/:user_id/interview", middleware.AuthRequired, middleware.AllowedGroups(2), middleware.SuperAdmin_approve, middleware.CSRFMiddleware, func(c *gin.Context) {
		queries := products.New(initializers.DB)
		ctx := context.Background()
		formToken, exists := c.Get("csrf_token")
		if !exists {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "CSRF verification failed",
			})
			return
		}
		user_id := c.Param("user_id")
		user_idInt, err := strconv.Atoi(user_id)
		if err != nil {
			log.Fatal(err.Error())
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		posting_id := c.Param("posting_id")
		posting_idInt, err := strconv.Atoi(posting_id)
		if err != nil {
			log.Fatal(err.Error())
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		exists, err = queries.Interview_exists(ctx, products.Interview_existsParams{
			PostingID: int32(posting_idInt),
			UserID:    int32(user_idInt),
		})
		if err != nil {
			log.Fatal(err.Error())
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		if exists {
			c.SetCookie("interview", "You have already sent Interview request to this candidate", 5, "/", "", false, true)
		}
		user, err := queries.GetUserByID(ctx, int32(user_idInt))
		if err != nil {
			log.Fatal(err.Error())
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		c.HTML(http.StatusOK, "recruiter_interviewForm.html", gin.H{
			"csrf_token": formToken,
			"title":      "Interview",
			"username":   user.Username,
		})
	})
	r.POST("/recruiter/jobPostings/:posting_id/applicants/:user_id/interview", middleware.AuthRequired, middleware.AllowedGroups(2), middleware.SuperAdmin_approve, controllers.RecruiterSubmitInterviewForm)

	// Applicant Endpoints
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
	r.GET("/applicant/resume/export/", middleware.AuthRequired, middleware.AllowedGroups(3), controllers.ApplicantExportResume)
	r.GET("/applicant/jobPostings/all", middleware.AuthRequired, middleware.AllowedGroups(3), func(c *gin.Context) {
		queries := products.New(initializers.DB)
		ctx := context.Background()
		job_postings, err := queries.GetJobPostings(ctx)
		if err != nil {
			log.Fatal(err.Error())
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		c.HTML(http.StatusOK, "applicant_jobPostingList.html", gin.H{
			"job_postings": job_postings,
		})
	})

	r.GET("/applicant/jobPostings/:posting_id", middleware.AuthRequired, middleware.AllowedGroups(3), func(c *gin.Context) {
		queries := products.New(initializers.DB)
		ctx := context.Background()
		posting_id := c.Param("posting_id")
		posting_idInt, err := strconv.Atoi(posting_id)
		if err != nil {
			log.Fatal(err.Error())
			c.AbortWithStatus(http.StatusInternalServerError)
			return
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
		job_posting, err := queries.GetJobPosting(ctx, int32(posting_idInt))
		if err != nil {
			log.Fatal(err.Error())
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		recruiter_profile, err := queries.GetRecruiterProfile(ctx, job_posting.UserID.Int32)
		if err != nil {
			log.Fatal(err.Error())
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		skills_req, err := queries.GetRequiredSkills(ctx, job_posting.PostingID)
		if err != nil {
			log.Fatal(err.Error())
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		recruiter_user, err := queries.GetUserByID(ctx, job_posting.UserID.Int32)
		if err != nil {
			log.Fatal(err.Error())
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		var postingDate string

		if job_posting.PostingDate.Valid {
			postingDate = job_posting.PostingDate.Time.Format("2006-01-02")
			fmt.Println("Posting Date:", postingDate)
		} else {
			log.Fatal("Posting date is null")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		exists, err = queries.CheckJobPosting_applicant(ctx, products.CheckJobPosting_applicantParams{
			PostingID: job_posting.PostingID,
			UserID:    req_user.UserID,
		})
		if err != nil {
			log.Fatal("Posting date is null")
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		c.HTML(http.StatusOK, "applicant_jobPostingDetail.html", gin.H{
			"title":             job_posting.JobTitle.String,
			"job_posting":       job_posting,
			"recruiter_profile": recruiter_profile,
			"skills_req":        skills_req,
			"recruiter_user":    recruiter_user,
			"posting_date":      postingDate,
			"applied":           exists,
		})

	})

	r.GET("/applicant/jobPostings/:posting_id/apply", middleware.AuthRequired, middleware.AllowedGroups(3), func(c *gin.Context) {
		queries := products.New(initializers.DB)
		ctx := context.Background()

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
		posting_id := c.Param("posting_id")
		fmt.Println(posting_id)
		posting_id_int, err := strconv.Atoi(posting_id)
		if err != nil {
			log.Fatal(err.Error())
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		queries.CreateJobPosting_applicants(ctx, products.CreateJobPosting_applicantsParams{
			PostingID: int32(posting_id_int),
			UserID:    req_user.UserID,
		})
		c.Redirect(http.StatusFound, "/")
	})
	r.GET("/applicant/interview/all", middleware.AuthRequired, middleware.AllowedGroups(3), func(c *gin.Context) {
		queries := products.New(initializers.DB)
		ctx := context.Background()

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

		results, err := queries.GetInterviews(ctx, req_user.UserID)
		if err != nil {
			log.Fatal(err.Error())
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		c.HTML(http.StatusFound, "applicant_interviewList.html", gin.H{
			"title":            "My Interviews",
			"interview_active": "active",
			"results":          results,
		})
	})

	r.GET("/applicant/interview/:posting_id/:user_id", middleware.AuthRequired, middleware.AllowedGroups(3), func(c *gin.Context) {
		queries := products.New(initializers.DB)
		ctx := context.Background()

		user_id := c.Param("user_id")
		user_idInt, err := strconv.Atoi(user_id)
		if err != nil {
			log.Fatal(err.Error())
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		posting_id := c.Param("posting_id")
		posting_idInt, err := strconv.Atoi(posting_id)
		if err != nil {
			log.Fatal(err.Error())
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		current_interview, err := queries.GetInterviewDetails(ctx, products.GetInterviewDetailsParams{
			UserID:    int32(user_idInt),
			PostingID: int32(posting_idInt),
		})

		if err != nil {
			log.Fatal(err.Error())
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		skills_req, err := queries.GetRequiredSkills(ctx, int32(posting_idInt))
		if err != nil {
			log.Fatal(err.Error())
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		interview_date := fmt.Sprintf("%02d/%02d/%d", current_interview.InterviewDatetime.Time.Day(), current_interview.InterviewDatetime.Time.Month(), current_interview.InterviewDatetime.Time.Year())
		interview_day := current_interview.InterviewDatetime.Time.Weekday().String()
		interview_time := fmt.Sprintf("%02d:%02d", current_interview.InterviewDatetime.Time.Hour(), current_interview.InterviewDatetime.Time.Minute())

		c.HTML(http.StatusOK, "applicant_interviewDetail.html", gin.H{
			"title":             fmt.Sprintf("Interview request for %s", current_interview.JobTitle),
			"current_interview": current_interview,
			"skills_req":        skills_req,
			"interview_date":    interview_date,
			"interview_day":     interview_day,
			"interview_time":    interview_time,
		})

	})

	r.GET("/applicant/interview/:posting_id/:user_id/decline/", middleware.AuthRequired, middleware.AllowedGroups(3), func(c *gin.Context) {
		queries := products.New(initializers.DB)
		ctx := context.Background()

		user_id := c.Param("user_id")
		user_idInt, err := strconv.Atoi(user_id)
		if err != nil {
			log.Fatal(err.Error())
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		posting_id := c.Param("posting_id")
		posting_idInt, err := strconv.Atoi(posting_id)
		if err != nil {
			log.Fatal(err.Error())
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}

		if err = queries.UpdateInterviewDeclineComplete(ctx, products.UpdateInterviewDeclineCompleteParams{
			DeclinedComplete: true,
			UserID:           int32(user_idInt),
			PostingID:        int32(posting_idInt),
		}); err != nil {
			log.Fatal(err.Error())
			c.AbortWithStatus(http.StatusInternalServerError)
			return
		}
		redirect_url := fmt.Sprintf("/applicant/interview/%s/%s", posting_id, user_id)
		c.Redirect(http.StatusFound, redirect_url)

	})

	r.GET("/applicant/interview/:posting_id/:user_id/decline/anotherDate", middleware.AuthRequired, middleware.AllowedGroups(3), middleware.CSRFMiddleware, func(c *gin.Context) {
		formToken, exists := c.Get("csrf_token")
		if !exists {
			log.Fatal("Csrf verification failed")
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.HTML(http.StatusOK, "applicant_interviewAnotherDate.html", gin.H{
			"csrf_token": formToken,
			"title":      "Rescheduling Interview",
		})
	})
	r.POST("/applicant/interview/:posting_id/:user_id/decline/anotherDate", middleware.AuthRequired, middleware.AllowedGroups(3), controllers.ApplicantRequestAnotherDate)

	//applicant views end here
	r.GET("logout/", middleware.AuthRequired, middleware.CSRFMiddleware, controllers.LogOut)

	r.Run()
}
