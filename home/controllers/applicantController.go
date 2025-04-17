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

type PartsData struct {
	Text string `json:"text"`
}

type ContentData struct {
	Parts []PartsData `json:"parts"`
}

type RequestData struct {
	Contents []ContentData `json:"contents"`
}

type ContentResponse struct {
	Parts []PartsData `json:"parts"`
	Role  string      `json:"role"`
}

type CandidateResponse struct {
	Content      ContentResponse `json:"content"`
	FinishReason string          `json:"finishReason"`
	AvgLogProbs  float64         `json:"avgLogprobs"`
}

type TokenDetails struct {
	Modality   string `json:"modality"`
	TokenCount int32  `json:"tokenCount"`
}

type UsageMetaData struct {
	PromptTokenCount        int32          `json:"promptTokenCount"`
	CandidatesTokenCount    int32          `json:"candidatesTokenCount"`
	TotalTokenCount         int32          `json:"totalTokenCount"`
	PromptTokensDetails     []TokenDetails `json:"promptTokensDetails"`
	CandidatesTokensDetails []TokenDetails `json:"candidatesTokensDetails"`
}

type ResponseData struct {
	Candidates    []CandidateResponse `json:"candidates"`
	UsageMetaData UsageMetaData       `json:"usageMetadata"`
	ModelVersion  string              `json:"modelVersion"`
}

func ApplicantHome(c *gin.Context) {
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

	userSkills, err := queries.GetApplicantSkills(ctx, req_user.UserID)
	if err != nil {
		log.Fatal(err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	userAppliedjobs, err := queries.GetJobPosting_applicants(ctx, req_user.UserID)
	if err != nil {
		log.Fatal(err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	fmt.Println(req_user.Username)
	c.HTML(http.StatusOK, "applicant_home.html", gin.H{
		"user":            req_user,
		"userProfile":     req_userProfile,
		"userSKills":      userSkills,
		"userAppliedjobs": userAppliedjobs,
	})
}

func ApplicantSearchSkills(c *gin.Context) {
	query := c.DefaultQuery("q", "")
	ctx := context.Background()
	queries := products.New(initializers.DB)

	results, err := queries.SearchSkillsFunc(ctx, query)

	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "error fetching results from database",
		})
		log.Println(err.Error())
		return
	}
	fmt.Println(results)
	c.JSON(http.StatusOK, gin.H{
		"results": results,
	})

}

func ApplicantAddSkill(c *gin.Context) {
	ctx := context.Background()
	queries := products.New(initializers.DB)

	user, exists := c.Get("user")
	if !exists {
		fmt.Println("1")
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "Unable to get user details",
		})
		return
	}
	req_user := user.(products.User)

	skill := c.PostForm("search-input")
	skill_obj, err := queries.GetSkill(ctx, sql.NullString{Valid: true, String: skill})
	if err != nil {
		skill_newObj, err := queries.CreateSkill(ctx, sql.NullString{Valid: true, String: skill})
		if err != nil {
			fmt.Println("2")
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": err.Error(),
			})
			return
		}
		queries.NewApplicantSkill(ctx, products.NewApplicantSkillParams{
			UserID:  req_user.UserID,
			SkillID: skill_newObj.SkillID,
		})

	} else {
		_, err := queries.GetApplicantSkill(ctx, products.GetApplicantSkillParams{
			UserID:  req_user.UserID,
			SkillID: skill_obj.SkillID,
		})
		if err != nil {
			queries.NewApplicantSkill(ctx, products.NewApplicantSkillParams{
				UserID:  req_user.UserID,
				SkillID: skill_obj.SkillID,
			})
		}
	}
	c.Redirect(http.StatusFound, "/applicant/updateProfile/")
}

func ApplicantDeleteSkill(c *gin.Context) {
	ctx := context.Background()
	queries := products.New(initializers.DB)

	user, exists := c.Get("user")
	if !exists {
		fmt.Println("1")
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "Unable to get user details",
		})
		return
	}
	req_user := user.(products.User)

	skill_id := c.Param("skill_id")

	skill_id_int, err := strconv.Atoi(skill_id)
	if err != nil {
		fmt.Println("2")
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": err.Error(),
		})
		return
	}

	err = queries.DeleteApplicantSkill(ctx, products.DeleteApplicantSkillParams{
		UserID:  req_user.UserID,
		SkillID: int32(skill_id_int),
	})

	if err != nil {
		fmt.Println("3")
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": err.Error(),
		})
		return
	}

	c.Redirect(http.StatusFound, "/applicant/updateProfile/")
}

func ApplicantResumeUpload(c *gin.Context) {
	ctx := context.Background()
	queries := products.New(initializers.DB)

	user, exists := c.Get("user")
	if !exists {
		fmt.Println("1")
		c.JSON(http.StatusUnauthorized, gin.H{
			"message": "Unable to get user details",
		})
		return
	}
	req_user := user.(products.User)

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
	req_userSkills, err := queries.GetApplicantSkills(ctx, req_user.UserID)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Failed geting user skills",
		})
	}
	skill_str := ""
	for _, skill := range req_userSkills {
		skill_str += skill.Name.String
	}

	file, err := c.FormFile("resume")
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed getting file import",
		})
		return
	}

	destination := fmt.Sprintf("resume/%d_resume.pdf", req_user.UserID)

	//Validating resume now
	f, r, err := pdf.Open(destination)

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
	errorString := GetDataFromResume(data, req_user.Email, req_userProfile.FirstName.String, req_userProfile.LastName.String, req_userProfile.School.String, req_userProfile.College.String, req_userProfile.Age.Int32, skill_str)
	if errorString == "" {
		//Data validated successfully
		err = c.SaveUploadedFile(file, destination)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "Failed to save file to database",
			})
			return
		}
		queries.AddApplicantProfile(ctx, products.AddApplicantProfileParams{
			ResumeFilename: sql.NullString{Valid: true, String: fmt.Sprintf("%d_resume.pdf", req_user.UserID)},
			UserID:         req_user.UserID,
		})
		c.SetCookie("resume_validationSuccess", "Resume successfully Validated & saved to database", 5, "/", "", false, true)
		c.Redirect(http.StatusFound, "/applicant/resume/")
		return
	} else {
		c.SetCookie("resume_validationFail", fmt.Sprintf("Failed to save resume to Database. %s", errorString), 5, "/", "", false, true)
		c.Redirect(http.StatusFound, "/applicant/resume/")
		return
	}

}

func GetDataFromResume(data string, email string, first_name string, last_name string, school string, college string, age int32, skill_str string) string {
	api_key := os.Getenv("GEMINI_API_KEY")

	api_url := fmt.Sprintf("https://generativelanguage.googleapis.com/v1beta/models/gemini-2.0-flash:generateContent?key=%s", api_key)

	prompt := fmt.Sprintf(
		"Match the following information from the given data :- First Name - %s, Last Name - %s , School - %s, College - %s, Age - %d, Skills - %s. Given Data - %s. Return the output string in the following format:- first_nameMatchT(True if it matches correctly) or first_nameMatchF(False if it doesn't matches) and similarly for the rest of the fields. Also at the end of the string, return an error in case if a field doesn't match in the following format error:(Write the error here like if a field is not found then say This field not found in resume, and similarly handle the cases of partial match(state which value was not present) or not matching at all).", first_name, last_name, school, college, age, skill_str, data)

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

	errorString := ""
	index := strings.Index(strings.ToLower(responseText), "error")

	if index != -1 {
		for i := index; i < len(responseText) && responseText[i] != '\n'; i++ {
			errorString += string(responseText[i])
		}
	}

	return errorString
}
func ApplicantExportResume(c *gin.Context) {
	ctx := context.Background()
	queries := products.New(initializers.DB)

	user, exists := c.Get("user")
	if !exists {
		fmt.Println("1")
		c.AbortWithStatus(http.StatusUnauthorized)
	}
	req_user := user.(products.User)

	fileName, err := queries.GetApplicantResume(ctx, req_user.UserID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "Failed to fetch file from database",
		})
		return
	}

	filePath := "resume/" + fileName.String
	c.FileAttachment(filePath, fileName.String)
	c.Redirect(http.StatusFound, "/applicant/resume/")
}

func ApplicantUpdateProfile(c *gin.Context) {
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
	FirstName := c.PostForm("first_name")
	LastName := c.PostForm("last_name")
	School := c.PostForm("school")
	College := c.PostForm("college")
	Age := c.PostForm("age")

	if Email == "" || Username == "" || FirstName == "" || LastName == "" || School == "" || College == "" || Age == "" {
		c.HTML(http.StatusFound, "applicant_updateProfile.html", gin.H{
			"csrf_token":     formToken,
			"profile_active": "active",
			"title":          "Update Profile",
			"email":          Email,
			"username":       Username,
			"first_name":     FirstName,
			"last_name":      LastName,
			"school":         School,
			"college":        College,
			"age":            Age,
			"message":        "No field can be left blank",
			"messageType":    "danger",
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
			c.HTML(http.StatusFound, "applicant_updateProfile.html", gin.H{
				"csrf_token":     formToken,
				"profile_active": "active",
				"title":          "Update Profile",
				"email":          Email,
				"username":       Username,
				"first_name":     FirstName,
				"last_name":      LastName,
				"school":         School,
				"college":        College,
				"age":            Age,
				"message":        "A user with this email already exists",
				"messageType":    "danger",
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
			c.HTML(http.StatusFound, "applicant_updateProfile.html", gin.H{
				"csrf_token":     formToken,
				"profile_active": "active",
				"title":          "Update Profile",
				"email":          Email,
				"username":       Username,
				"first_name":     FirstName,
				"last_name":      LastName,
				"school":         School,
				"college":        College,
				"age":            Age,
				"message":        err.Error(),
				"messageType":    "danger",
			})
			return
		}

		age_int, err := strconv.Atoi(Age)
		if err != nil {
			c.HTML(http.StatusFound, "applicant_updateProfile.html", gin.H{
				"csrf_token":     formToken,
				"profile_active": "active",
				"title":          "Update Profile",
				"email":          Email,
				"username":       Username,
				"first_name":     FirstName,
				"last_name":      LastName,
				"school":         School,
				"college":        College,
				"age":            Age,
				"message":        err.Error(),
				"messageType":    "danger",
			})
			return
		}
		err = queries.UpdateApplicantProfile(ctx, products.UpdateApplicantProfileParams{
			FirstName: sql.NullString{Valid: true, String: FirstName},
			LastName:  sql.NullString{Valid: true, String: LastName},
			School:    sql.NullString{Valid: true, String: School},
			College:   sql.NullString{Valid: true, String: College},
			Age:       sql.NullInt32{Valid: true, Int32: int32(age_int)},
			UserID:    req_user.UserID,
		})

		if err != nil {
			c.HTML(http.StatusFound, "applicant_updateProfile.html", gin.H{
				"csrf_token":     formToken,
				"profile_active": "active",
				"title":          "Update Profile",
				"email":          Email,
				"username":       Username,
				"first_name":     FirstName,
				"last_name":      LastName,
				"school":         School,
				"college":        College,
				"age":            Age,
				"message":        err.Error(),
				"messageType":    "danger",
			})
			return
		}
		fmt.Println("Profile Updated")
		c.Redirect(http.StatusFound, "/applicant/home")
		return
	}
}

func ApplicantRequestAnotherDate(c *gin.Context) {
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

	current_interview, err := queries.GetInterview(ctx, products.GetInterviewParams{
		UserID:    int32(user_idInt),
		PostingID: int32(posting_idInt),
	})

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
		c.HTML(http.StatusOK, "applicant_interviewAnotherDate.html", gin.H{
			"csrf_token":  formToken,
			"title":       "Rescheduling Interview",
			"message":     "Invalid date or time",
			"messageType": "danger",
		})
		return
	}
	interview_date, err := time.Parse("2006-01-02", dateStr)
	if err != nil {
		log.Fatal(err.Error())
		c.HTML(http.StatusOK, "applicant_interviewAnotherDate.html", gin.H{
			"csrf_token":  formToken,
			"title":       "Rescheduling Interview",
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
		current_interview.InterviewDatetime.Time.Second(),
		current_interview.InterviewDatetime.Time.Nanosecond(),
		time.Local,
	)

	//The entered date and time should not be before the current time
	if interview_dateTime.Before(time.Now()) {
		c.HTML(http.StatusOK, "applicant_interviewAnotherDate.html", gin.H{
			"csrf_token":  formToken,
			"title":       "Rescheduling Interview",
			"message":     "The entered date and time should not be before the current time",
			"messageType": "danger",
		})
		return
	}

	if err = queries.UpdateInterviewAnotherDateReq(ctx, products.UpdateInterviewAnotherDateReqParams{
		PostingID:       int32(posting_idInt),
		UserID:          int32(user_idInt),
		AnotherDatetime: sql.NullTime{Valid: true, Time: interview_dateTime},
	}); err != nil {
		log.Fatal(err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	redirect_url := fmt.Sprintf("/applicant/interview/%s/%s", posting_id, user_id)
	c.Redirect(http.StatusFound, redirect_url)
}
