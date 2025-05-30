package controllers

import (
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
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/google"
)

var GoogleOauthConfig *oauth2.Config

func InitGoogleOauth() {
	GoogleOauthConfig = &oauth2.Config{
		ClientID:     os.Getenv("OAUTH_CLIENT_ID"),
		ClientSecret: os.Getenv("OAUTH_CLIENT_SECRET"),
		RedirectURL:  "http://127.0.0.1:8000/auth/google/callback",
		Scopes:       []string{"https://www.googleapis.com/auth/userinfo.email", "https://www.googleapis.com/auth/userinfo.profile"},
		Endpoint:     google.Endpoint,
	}
}

func SignUp(c *gin.Context) {
	ctx := context.Background()

	csrf_token, err := c.Cookie("CSRF_Token")

	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
	}
	formToken := c.PostForm("csrf_token")

	if csrf_token != formToken {
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

	Email := c.PostForm("email")
	Username := c.PostForm("username")
	Password1 := c.PostForm("password1")
	Password2 := c.PostForm("password2")
	Role := c.PostForm("role")

	if Password1 != Password2 {
		c.HTML(http.StatusOK, "signup.html", gin.H{
			"email":       Email,
			"username":    Username,
			"password1":   Password1,
			"password2":   Password2,
			"csrf_token":  csrf_token,
			"message":     "The 2 passwords do not match",
			"messageType": "danger",
		})
		return
	}

	queries := products.New(initializers.DB)
	_, err = queries.GetUserByEmail(ctx, Email)
	if err != nil {
		hash, err := bcrypt.GenerateFromPassword([]byte(Password1), bcrypt.DefaultCost)
		if err != nil {
			c.HTML(http.StatusOK, "signup.html", gin.H{
				"email":       Email,
				"username":    Username,
				"password1":   Password1,
				"password2":   Password2,
				"csrf_token":  csrf_token,
				"message":     "Failed to hash password",
				"messageType": "danger",
			})
			return
		}

		var role_id int
		switch Role {
		case "Super Admin":
			role_id = 1
		case "Recruiter":
			role_id = 2
		case "Applicant":
			role_id = 3
		}

		newUser, err := queries.CreateNewUser(ctx, products.CreateNewUserParams{
			Email:        Email,
			Username:     Username,
			PasswordHash: sql.NullString{String: string(hash), Valid: true},
			RoleID:       sql.NullInt32{Int32: int32(role_id), Valid: true},
		})
		if err != nil {
			c.HTML(http.StatusOK, "signup.html", gin.H{
				"email":       Email,
				"username":    Username,
				"password1":   Password1,
				"password2":   Password2,
				"csrf_token":  csrf_token,
				"message":     err.Error(),
				"messageType": "danger",
			})
			return
		}
		//making user profile
		if role_id == 2 {
			err := queries.CreateRecruiterProfile(ctx, int32(newUser.UserID))
			if err != nil {
				c.HTML(http.StatusOK, "signup.html", gin.H{
					"email":       Email,
					"username":    Username,
					"password1":   Password1,
					"password2":   Password2,
					"csrf_token":  csrf_token,
					"message":     err.Error(),
					"messageType": "danger",
				})
				return
			}
		} else if role_id == 3 {
			err := queries.CreateApplicantProfile(ctx, int32(newUser.UserID))
			if err != nil {
				c.HTML(http.StatusOK, "signup.html", gin.H{
					"email":       Email,
					"username":    Username,
					"password1":   Password1,
					"password2":   Password2,
					"csrf_token":  csrf_token,
					"message":     err.Error(),
					"messageType": "danger",
				})
				return
			}
		}
		c.SetCookie("signup_success", "Your account was created successfully! You can now Log In", 10, "/", "", false, false)
		c.Redirect(http.StatusFound, "login")
	} else {
		c.HTML(http.StatusOK, "signup.html", gin.H{
			"email":       Email,
			"username":    Username,
			"password1":   Password1,
			"password2":   Password2,
			"csrf_token":  csrf_token,
			"message":     "The user with this email already exists",
			"messageType": "danger",
		})
		return
	}
}

func Login(c *gin.Context) {
	ctx := context.Background()

	csrf_token, err := c.Cookie("CSRF_Token")
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
	}
	formToken := c.PostForm("csrf_token")

	if csrf_token != formToken {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	//expiring the cookie after verification
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

	Email := c.PostForm("email")
	Password := c.PostForm("password")

	queries := products.New(initializers.DB)

	req_user, err := queries.GetUserByEmail(ctx, Email)
	if err != nil {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"email":       Email,
			"password":    Password,
			"message":     "Invalid email or password",
			"messageType": "danger",
		})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(req_user.PasswordHash.String), []byte(Password))
	if err != nil {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"email":       Email,
			"password":    Password,
			"message":     "Invalid email or password",
			"messageType": "danger",
		})
		return
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": req_user.UserID,
		"exp": time.Now().Add(time.Hour * 24 * 30).Unix(),
	})

	// Sign and get the complete encoded token as a string using the secret
	tokenString, err := token.SignedString([]byte(os.Getenv("SECRET")))

	if err != nil {
		c.HTML(http.StatusOK, "login.html", gin.H{
			"email":       Email,
			"password":    Password,
			"message":     "Failed to create JWT Token",
			"messageType": "danger",
		})
		return
	}

	//sending cookie back to client in form of cookie
	fmt.Println("I am here")
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("Authorization", tokenString, 3600*24*30, "", "", false, true)
	c.SetCookie("login_success", fmt.Sprintf("Successfully logged in as %s", req_user.Username), 10, "/", "", false, false)
	c.Redirect(http.StatusFound, "/")

}

func LogOut(c *gin.Context) {
	for _, cookie := range c.Request.Cookies() {
		http.SetCookie(c.Writer, &http.Cookie{
			Name:     cookie.Name,
			Value:    "",
			Path:     "/",
			HttpOnly: true,
			Secure:   false,
			SameSite: http.SameSiteStrictMode,
			Expires:  time.Unix(0, 0),
			MaxAge:   -1,
		})
	}
	c.Redirect(http.StatusFound, "/login")
}

func HandleGoogleLogin(c *gin.Context) {
	url := GoogleOauthConfig.AuthCodeURL("random-state")
	c.Redirect(http.StatusTemporaryRedirect, url)
}

func HandleGoogleCallback(c *gin.Context) {
	ctx := context.Background()
	queries := products.New(initializers.DB)

	code := c.Query("code")
	if code == "" {
		c.AbortWithStatusJSON(http.StatusBadRequest, gin.H{"error": "code not found"})
		return
	}
	token, err := GoogleOauthConfig.Exchange(ctx, code)
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	client := GoogleOauthConfig.Client(ctx, token)
	resp, err := client.Get("https://www.googleapis.com/oauth2/v2/userinfo")
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	defer resp.Body.Close()
	userData, _ := io.ReadAll(resp.Body)

	var userInfo struct {
		Email         string `json:"email"`
		VerifiedEmail bool   `json:"verified_email"`
		Name          string `json:"name"`
		Picture       string `json:"picture"`
		Id            string `json:"id"`
	}
	if err := json.Unmarshal(userData, &userInfo); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to parse user info"})
		return
	}
	exists, err := queries.CheckUserByEmail(ctx, userInfo.Email)
	if err != nil {
		log.Fatal(err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	var req_user products.User
	if !exists {
		req_user, err = queries.CreateNewUser(ctx, products.CreateNewUserParams{
			Username: userInfo.Name,
			Email:    userInfo.Email,
		})
		if err != nil {
			log.Fatal(err.Error())
			c.AbortWithStatus(http.StatusInternalServerError)
		}
	} else {
		req_user, err = queries.GetUserByEmail(ctx, userInfo.Email)
		if err != nil {
			log.Fatal(err.Error())
			c.AbortWithStatus(http.StatusInternalServerError)
		}
	}

	jwt_token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"sub": req_user.UserID,
		"exp": time.Now().Add(time.Hour * 24 * 30).Unix(),
	})

	tokenString, err := jwt_token.SignedString([]byte(os.Getenv("SECRET")))

	if err != nil {
		log.Fatal(err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
	}
	c.SetCookie("Authorization", tokenString, 3600*24*30, "", "", false, true)
	c.SetCookie("login_success", fmt.Sprintf("Successfully logged in as %s", req_user.Username), 10, "/", "", false, false)
	c.Redirect(http.StatusFound, "/")
}

func UserGroups(c *gin.Context) {
	ctx := context.Background()
	queries := products.New(initializers.DB)

	csrf_token, err := c.Cookie("CSRF_Token")
	if err != nil {
		c.AbortWithStatus(http.StatusUnauthorized)
	}
	formToken := c.PostForm("csrf_token")

	if csrf_token != formToken {
		c.AbortWithStatus(http.StatusUnauthorized)
		return
	}

	//expiring the cookie after verification
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

	Role := c.PostForm("role")
	var role_id int
	switch Role {
	case "Super Admin":
		role_id = 1
	case "Recruiter":
		role_id = 2
	case "Applicant":
		role_id = 3
	}

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

	err = queries.UpdateUserRoleID(ctx, products.UpdateUserRoleIDParams{
		UserID: user.UserID,
		RoleID: sql.NullInt32{Valid: true, Int32: int32(role_id)},
	})

	if err != nil {
		log.Fatal(err.Error())
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}
	c.Redirect(http.StatusFound, "/")
}
