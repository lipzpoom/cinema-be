package handlers

import (
	"context"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)

type AuthHandler struct {
	DB *mongo.Database
}

type UserJWT struct {
	UserID     string `json:"user_id"`
	ImgProfile string `json:"img_profile"`
	Email      string `json:"email"`
	Name       string `json:"name"`
	Role       string `json:"role"`
}

func generateJWT(userJWT UserJWT) (string, error) {
	secret := os.Getenv("JWT_SECRET")
	if secret == "" {
		return "", errors.New("JWT_SECRET not set in environment variables")
	}

	claims := jwt.MapClaims{
		"user_id":     userJWT.UserID,
		"img_profile": userJWT.ImgProfile,
		"email":       userJWT.Email,
		"name":        userJWT.Name,
		"role":        userJWT.Role,
		"exp":         time.Now().Add(time.Hour * 72).Unix(),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err := token.SignedString([]byte(secret))
	if err != nil {
		return "", err
	}

	return signedToken, nil
}

func getUserIDString(id interface{}) string {
	if oid, ok := id.(primitive.ObjectID); ok {
		return oid.Hex()
	}
	if str, ok := id.(string); ok {
		return str
	}
	return ""
}

// GoogleLogin godoc
// @Summary Login with Google Provider
// @Description Login using Google Firebase idToken and returns a custom internal JWT
// @Tags Auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Success 201 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/auth/google [post]
func (h *AuthHandler) GoogleLogin(c *gin.Context) {
	// retrieved from middleware
	googleID := c.GetString("user_id")
	imgProfile := c.GetString("picture")
	email := c.GetString("email")
	name := c.GetString("name")

	if googleID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "User info not found in token"})
		return
	}

	usersColl := h.DB.Collection("users")

	// find User in the database
	var user map[string]interface{}
	err := usersColl.FindOne(context.Background(), bson.M{"google_id": googleID}).Decode(&user)

	if err == mongo.ErrNoDocuments {
		// if not found, create new user
		newUser := bson.M{
			"google_id":   googleID,
			"img_profile": imgProfile,
			"email":       email,
			"name":        name,
			"role":        "USER",
			"created_at":  time.Now(),
			"updated_at":  time.Now(),
		}

		res, insertErr := usersColl.InsertOne(context.Background(), newUser)
		if insertErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to create user"})
			return
		}

		// return user info
		newUser["_id"] = res.InsertedID
		userIDStr := getUserIDString(res.InsertedID)
		userJWT := UserJWT{
			UserID:     userIDStr,
			ImgProfile: imgProfile,
			Email:      email,
			Name:       name,
			Role:       "USER",
		}
		token, jwtErr := generateJWT(userJWT)
		if jwtErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate JWT"})
			return
		}
		c.JSON(http.StatusCreated, gin.H{"message": "User registered successfully", "token": token})
		return
	} else if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	hasChanges := false
	updateUser := bson.M{}

	// If exists user, return user info
	if dbName, ok := user["name"].(string); !ok || dbName != name {
		updateUser["name"] = name
		hasChanges = true
		user["name"] = name
	}
	if dbEmail, ok := user["email"].(string); !ok || dbEmail != email {
		updateUser["email"] = email
		hasChanges = true
		user["email"] = email
	}
	if dbImg, ok := user["img_profile"].(string); !ok || dbImg != imgProfile {
		updateUser["img_profile"] = imgProfile
		hasChanges = true
		user["img_profile"] = imgProfile
	}

	if hasChanges {
		updateUser["updated_at"] = time.Now()
		_, updateErr := usersColl.UpdateOne(context.Background(), bson.M{"google_id": googleID}, bson.M{"$set": updateUser})
		if updateErr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to update user"})
			return
		}
	}

	// create JWT token
	userIDStr := getUserIDString(user["_id"])
	roleStr := user["role"].(string)

	userJWT := UserJWT{
		UserID:     userIDStr,
		ImgProfile: user["img_profile"].(string),
		Email:      user["email"].(string),
		Name:       user["name"].(string),
		Role:       roleStr,
	}

	token, jwtErr := generateJWT(userJWT)
	if jwtErr != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to generate JWT"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Login successfully", "token": token})
}

// GetProfile godoc
// @Summary Get Current User Profile
// @Description Decodes the JWT to retrieve current logged in user information
// @Tags Auth
// @Security BearerAuth
// @Produce json
// @Success 200 {object} map[string]interface{}
// @Failure 400 {object} map[string]interface{}
// @Failure 401 {object} map[string]interface{}
// @Failure 404 {object} map[string]interface{}
// @Failure 500 {object} map[string]interface{}
// @Router /api/auth/profile [get]
func (h *AuthHandler) GetProfile(c *gin.Context) {
	userID := c.GetString("user_id")
	if userID == "" {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "Unauthorized"})
		return
	}

	objectID, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid user ID format"})
		return
	}

	var user bson.M
	err = h.DB.Collection("users").FindOne(c.Request.Context(), bson.M{"_id": objectID}).Decode(&user)
	if err != nil {
		if err == mongo.ErrNoDocuments {
			c.JSON(http.StatusNotFound, gin.H{"error": "User not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Database error"})
		return
	}

	c.JSON(http.StatusOK, user)
}
