package controllers

import (
	"bytes"
	"context"
	"database/sql"
	"encoding/json"
	"fmt"
	"home/initializers"
	"home/products"
	"io"
	"log"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/ledongthuc/pdf"
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
	job_title := c.PostForm("job_title")
	job_description := c.PostForm("job_description")
	fmt.Println("posting_id in addSKill view:- ", posting_id)
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
		_, err := queries.GetSkill_req(ctx, products.GetSkill_reqParams{
			PostingID: int32(posting_idInt),
			SkillID:   skill_obj.SkillID,
		})
		if err != nil {
			queries.CreateSkill_Req(ctx, products.CreateSkill_ReqParams{
				PostingID: int32(posting_idInt),
				SkillID:   skill_obj.SkillID,
			})
		}
	}
	c.SetCookie("posting_id", posting_id, 5, "/", "", false, true)
	c.SetCookie("job_title", job_title, 5, "/", "", false, true)
	c.SetCookie("job_description", job_description, 5, "/", "", false, true)
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

	if JobTitle != "" || JobDescription != "" {
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

		c.Redirect(http.StatusFound, "/recruiter/jobPostings/all")
	} else {
		skills_req, err := queries.GetRequiredSkills(ctx, int32(posting_idInt))
		if err != nil {
			log.Fatal(err.Error())
			c.AbortWithStatus(http.StatusInternalServerError)
		}
		c.HTML(http.StatusFound, "recruiter_createJobPosting.html", gin.H{
			"csrf_token":      formToken,
			"add_job_active":  "active",
			"title":           "New Job Posting",
			"posting_id":      posting_id,
			"job_title":       JobTitle,
			"job_description": JobDescription,
			"skills_req":      skills_req,
			"message":         "No field can be left blank",
			"messageType":     "danger",
		})
		return
	}

}

func RecruiterResumeImport(c *gin.Context) {
	ctx := context.Background()
	queries := products.New(initializers.DB)
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

	f, r, err := pdf.Open(filePath)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	defer f.Close()

	var buf bytes.Buffer
	b, err := r.GetPlainText()
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": err.Error(),
		})
		return
	}
	buf.ReadFrom(b)
	data := buf.String()
	api_key := os.Getenv("GEMINI_API_KEY")

	api_url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent?key=%s", api_key)

	prompt := fmt.Sprintf(
		"Extract key information from the following data :- First Name, Last Name, School, College, Age , Skills, Key Projects, Companies in which the applicant has worked.Handle the case of student or 0 companies. Given Data - %s. In the final output there should be no special charecter like '*'(asterisk strictly not allowed) etc. except new-line charecter. Brackets and spaces are allowed", data)

	request := RequestData{
		Contents: []ContentData{
			{
				Parts: []PartsData{
					{
						Text: prompt,
					},
				},
			},
		},
	}

	json_data, err := json.Marshal(request)
	if err != nil {
		log.Fatal(err)
	}
	resp, err := http.Post(api_url, "application/json", bytes.NewBuffer(json_data))
	if err != nil {
		log.Fatal(err)
	}

	defer resp.Body.Close()
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		log.Fatal(err)
	}

	var dat ResponseData

	if err := json.Unmarshal(body, &dat); err != nil {
		log.Fatal(err)
	}
	var responseText string

	if len(dat.Candidates) > 0 && len(dat.Candidates[0].Content.Parts) > 0 {
		responseText = dat.Candidates[0].Content.Parts[0].Text
	} else {
		log.Fatal("Error geting response from Gemini")
	}

	myHTMLResponse := strings.ReplaceAll(responseText, "\n", "<br>")

	fmt.Println(myHTMLResponse)

	c.SetCookie("resume_parse", myHTMLResponse, 5, "/", "", false, true)
	redirect_url := fmt.Sprintf("/recruiter/jobPostings/%s/applicants/%s", posting_id, user_id)
	c.Redirect(http.StatusFound, redirect_url)

}

func RecruiterSubmitInterviewForm(c *gin.Context) {
	ctx := context.Background()
	queries := products.New(initializers.DB)

	// CSRF Verification
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

	user, err := queries.GetUserByID(ctx, int32(user_idInt))
	if err != nil {
		log.Fatal(err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	timeStr := c.PostForm("interview_time")
	dateStr := c.PostForm("interview_date")

	interview_time, err := time.Parse("15:04", timeStr)
	if err != nil {
		log.Fatal(err.Error())
		c.HTML(http.StatusOK, "recruiter_interviewForm.html", gin.H{
			"csrf_token":  formToken,
			"title":       "Interview",
			"username":    user.Username,
			"message":     "Invalid date or time",
			"messageType": "danger",
		})
		return
	}
	interview_date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		log.Fatal(err.Error())
		c.HTML(http.StatusOK, "recruiter_interviewForm.html", gin.H{
			"csrf_token":  formToken,
			"title":       "Interview",
			"username":    user.Username,
			"message":     "Invalid date or time",
			"messageType": "danger",
		})
		return
	}
	interview_dateTime := time.Date(
		interview_date.Year(),
		interview_date.Month(),
		interview_date.Day(),
		interview_time.Hour(),
		interview_time.Minute(),
		interview_date.Second(),
		interview_date.Nanosecond(),
		time.Local,
	)

	//The entered date and time should not be before the current time
	if interview_dateTime.Before(time.Now()) {
		c.HTML(http.StatusOK, "recruiter_interviewForm.html", gin.H{
			"csrf_token":  formToken,
			"title":       "Interview",
			"username":    user.Username,
			"message":     "The entered date and time should not be before the current time",
			"messageType": "danger",
		})
		return
	}
	newInterview, err := queries.CreateInterview(ctx, products.CreateInterviewParams{
		UserID:            int32(user_idInt),
		PostingID:         int32(posting_idInt),
		InterviewDatetime: sql.NullTime{Valid: true, Time: interview_dateTime},
	})

	if err != nil {
		log.Fatal(err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	fmt.Println(newInterview.InterviewDatetime.Time)
	redirect_url := fmt.Sprintf("/recruiter/jobPostings/%s/applicants/%s/", posting_id, user_id)
	c.Redirect(http.StatusFound, redirect_url)
}
