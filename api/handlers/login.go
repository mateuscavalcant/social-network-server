package handlers

import (
	"log"
	"net/http"
	"os"
	"social-network-server/pkg/database"
	"social-network-server/pkg/models"
	"social-network-server/pkg/models/errs"

	"strings"

	"github.com/dgrijalva/jwt-go"

	"github.com/gin-gonic/gin"
	"golang.org/x/crypto/bcrypt"
)

// UserLogin handles user login requests.
func UserLogin(c *gin.Context) {
	var user models.User

	// Extract identifier (username or email) and password from the request
	identifier := strings.TrimSpace(c.PostForm("identifier"))
	password := strings.TrimSpace(c.PostForm("password"))

	// Prepare error response object
	resp := errs.ErrorResponse{
		Error: make(map[string]string),
	}

	// Determine if the identifier is an email or username
	isEmail := strings.Contains(identifier, "@")
	var queryField string
	if isEmail {
		queryField = "email"
	} else {
		queryField = "username"
	}

	// Connect to the database
	db := database.GetDB()

	// Query the user's ID, email, and password from the database based on the identifier
	err := db.QueryRow("SELECT id, email, password FROM user WHERE "+queryField+"=?", identifier).Scan(&user.ID, &user.Email, &user.Password)
	if err != nil {
		log.Println("Error executing SQL statement:", err)
		resp.Error["credentials"] = "Invalid credentials"
		c.JSON(400, resp)
		return
	}

	// Print the user's ID
	log.Println("User ID:", user.ID)

	// Compare the hashed password from the database with the provided password
	encErr := bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(password))
	if encErr != nil {
		resp.Error["password"] = "Invalid password"
		c.JSON(400, resp)
		return
	}

	// Set the user ID in the context of the request
	c.Set("id", user.ID)

	// Create JWT token
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, jwt.MapClaims{
		"id":    user.ID,
		"email": user.Email,
	})

	// Sign and get the complete encoded token as a string
	tokenString, err := token.SignedString([]byte(os.Getenv("SESSION_SECRET")))
	if err != nil {
		log.Println("Error signing token:", err)
		c.JSON(500, gin.H{"error": "Failed to generate token"})
		return
	}

	// Create a cookie with the token
	cookie := http.Cookie{
		Name:     "token",
		Value:    tokenString,
		HttpOnly: true,
	}

	// Set the cookie in the response
	http.SetCookie(c.Writer, &cookie)

	// Return success message to the client
	c.JSON(200, gin.H{
		"message": "User logged in successfully",
		"token":   tokenString,
	})
}
