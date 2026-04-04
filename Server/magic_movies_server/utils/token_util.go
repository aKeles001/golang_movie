package utils

import (
	"context"
	"errors"
	"os"
	"time"

	database "github.com/aKeles001/golang_movie/Server/magic_movies_server/database"
	"github.com/gin-gonic/gin"
	jwt "github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

type SignedDetails struct {
	Email      string
	First_name string
	Last_name  string
	Role       string
	UserId     string
	jwt.RegisteredClaims
}

var SECRET_KEY string = os.Getenv("SECRET_KEY")
var SECRET_REFRESH_KEY string = os.Getenv("SECRET_REFRESH_KEY")

func GenerateAllTokens(email string, firstName string, lastName string, role string, userId string) (signedToken string, signedRefreshToken string, err error) {
	claims := &SignedDetails{
		Email:      email,
		First_name: firstName,
		Last_name:  lastName,
		Role:       role,
		UserId:     userId,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "MagicMovie",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24)),
		},
	}
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	signedToken, err = token.SignedString([]byte(SECRET_KEY))
	if err != nil {
		return "", "", err
	}
	refreshClaims := &SignedDetails{
		Email:      email,
		First_name: firstName,
		Last_name:  lastName,
		Role:       role,
		UserId:     userId,
		RegisteredClaims: jwt.RegisteredClaims{
			Issuer:    "MagicMovie",
			IssuedAt:  jwt.NewNumericDate(time.Now()),
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(time.Hour * 24 * 7)),
		},
	}
	refreshToken := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims)
	signedRefreshToken, err = refreshToken.SignedString([]byte(SECRET_REFRESH_KEY))
	if err != nil {
		return "", "", err
	}
	return signedToken, signedRefreshToken, nil
}

func UpdateAllTokens(signedToken string, signedRefreshToken string, userId string, client *mongo.Client) (err error) {
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	updateAt, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))

	updateData := bson.M{
		"$set": bson.M{
			"token":         signedToken,
			"refresh_token": signedRefreshToken,
			"updated_at":    updateAt,
		},
	}

	var userCollection *mongo.Collection = database.OpenCollection("users", client)
	_, err = userCollection.UpdateOne(ctx, bson.M{"user_id": userId}, updateData)
	if err != nil {
		return err
	}
	return nil
}

func GetAccessToken(c *gin.Context) (string, error) {
	authHeader := c.GetHeader("Authorization")
	if authHeader == "" {
		return "", errors.New("Authorization header is missing")
	}
	tokenString := authHeader[len("Bearer "):]
	if tokenString == "" {
		return "", errors.New("Token is missing from Authorization header")
	}
	return tokenString, nil
}

func ValidateToken(tokenString string) (claims *SignedDetails, err error) {
	claims = &SignedDetails{}
	token, err := jwt.ParseWithClaims(tokenString, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(SECRET_KEY), nil
	})
	if err != nil {
		return nil, err
	}
	if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
		return nil, err
	}
	if claims.ExpiresAt.Time.Before(time.Now()) {
		return nil, errors.New("Token has expired")
	}
	return claims, nil
}

func GetUserIdFromContext(c *gin.Context) (string, error) {
	userId, exists := c.Get("user_id")
	if !exists {
		return "", errors.New("User ID not found in context")
	}
	userIdStr, ok := userId.(string)
	if !ok {
		return "", errors.New("User ID in context is not a string")
	}
	return userIdStr, nil
}

func GetRoleFromContext(c *gin.Context) (string, error) {
	role, exists := c.Get("role")
	if !exists {
		return "", errors.New("Role not found in context")
	}
	roleStr, ok := role.(string)
	if !ok {
		return "", errors.New("Role in context is not a string")
	}
	return roleStr, nil
}
