package controllers

import (
	"context"
	"database/sql"
	"home/initializers"
	"home/products"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"golang.org/x/crypto/bcrypt"
)

func SignUp(c *gin.Context) {
	ctx := context.Background()

	var body struct {
		Email     string
		Username  string
		Password1 string
		Password2 string
		Role      string
	}

	if c.Bind(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to read request body",
		})
		return
	}

	if body.Password1 != body.Password2 {
		c.JSON(http.StatusBadRequest, gin.H{
			"error":     "The 2 passwords do not match",
			"email":     body.Email,
			"username":  body.Username,
			"Password1": body.Password1,
			"password2": body.Password2,
		})
		return
	}

	hash, err := bcrypt.GenerateFromPassword([]byte(body.Password1), bcrypt.DefaultCost)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to hash password",
		})
		return
	}

	var role_id int
	switch body.Role {
	case "super_admin":
		role_id = 1
	case "recruiter":
		role_id = 2
	case "applicant":
		role_id = 3
	}

	queries := products.New(initializers.DB)

	err = queries.CreateNewUser(ctx, products.CreateNewUserParams{
		Email:        body.Email,
		Username:     body.Username,
		PasswordHash: sql.NullString{String: string(hash), Valid: true},
		RoleID:       sql.NullInt32{Int32: int32(role_id), Valid: true},
	})
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "User with this email already exists",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "User created successfully",
	})
}

func Login(c *gin.Context) {
	ctx := context.Background()

	var body struct {
		Email    string
		Password string
	}

	if c.Bind(&body) != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed reading request body",
		})
		return
	}

	queries := products.New(initializers.DB)

	req_user, err := queries.GetUserByEmail(ctx, body.Email)
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "The user with requested email does not exist",
		})
		return
	}

	err = bcrypt.CompareHashAndPassword([]byte(req_user.PasswordHash.String), []byte(body.Password))
	if err != nil {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "Invalid email or password",
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
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "Failed to create JWT Token",
		})
		return
	}

	//sending cookie back to client in form of cookie
	c.SetSameSite(http.SameSiteLaxMode)
	c.SetCookie("Authorization", tokenString, 3600*24*30, "", "", false, true)
	c.JSON(http.StatusOK, gin.H{})

}

func Validate(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "I am logged in",
	})
}
