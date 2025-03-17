package api

import (
	"backend/helper"
	"context"
	"errors"
	"time"

	"github.com/dgrijalva/jwt-go"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

type User struct {
	ID        string    `bson:"_id,omitempty" json:"id"`
	UserCode  string    `bson:"userCode" json:"userCode"`
	Email     string    `bson:"email" json:"email"`
	Password  string    `bson:"password" json:"password"`
	Role      string    `bson:"role" json:"role"`
	CreatedAt time.Time `bson:"createdAt" json:"createdAt"`
	UpdatedAt time.Time `bson:"updatedAt" json:"updatedAt"`
}

type LoginRequest struct {
	Email    string `json:"email"`
	Password string `json:"password"`
}

type LoginResponse struct {
	Message string `json:"message"`
}

type Claims struct {
	UserCode string `json:"userCode"`
	Role     string `json:"role"`
	jwt.StandardClaims
}

var secretKey = []byte("supersecretkey1234")

func LoginUser(client *mongo.Client, input LoginRequest) (string, error) {
	collection := client.Database("users").Collection("users")
	filter := bson.D{{Key: "email", Value: input.Email}}
	var user User
	err := collection.FindOne(context.TODO(), filter).Decode(&user)
	if err != nil {
		return "", errors.New("no such user")
	}

	err = bcrypt.CompareHashAndPassword([]byte(user.Password), []byte(input.Password))
	if err != nil {
		return "", errors.New("incorrect password")
	}

	tokenString, err := generateJWT(user.UserCode, user.Role)
	if err != nil {
		return "", errors.New("could not generate token")
	}

	return tokenString, nil
}

func RegisterUser(client *mongo.Client, user User) (string, error) {
	collection := client.Database("users").Collection("users")
	if err := collection.FindOne(context.TODO(), bson.D{{Key: "email", Value: user.Email}}).Err(); err == nil {
		return "", errors.New("email already exists")
	}
	hashedPassword, err := bcrypt.GenerateFromPassword([]byte(user.Password), bcrypt.DefaultCost)
	if err != nil {
		return "", err
	}
	user.Password = string(hashedPassword)
	user.UserCode = helper.GenerateID(8)
	user.CreatedAt = time.Now()
	user.UpdatedAt = time.Now()

	tokenString, err := generateJWT(user.UserCode, user.Role)
	if err != nil {
		return "", errors.New("could not generate token")
	}

	_, err = collection.InsertOne(context.TODO(), user)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}

func DeleteUser(client *mongo.Client, userCode string) {
	collection := client.Database("users").Collection("users")
	collection.FindOneAndDelete(context.TODO(), bson.D{{Key: "userCode", Value: userCode}})
}

func GetAllUsers(client *mongo.Client) []User {
	collection := client.Database("users").Collection("users")

	filter := bson.D{{}}

	cursor, _ := collection.Find(context.TODO(), filter)
	defer cursor.Close(context.TODO())

	var users []User
	cursor.All(context.TODO(), &users)

	return users
}

func generateJWT(userCode, userRole string) (string, error) {
	claims := Claims{
		UserCode: userCode,
		Role:     userRole,
		StandardClaims: jwt.StandardClaims{
			ExpiresAt: time.Now().Add(time.Hour * 24).Unix(),
		},
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	tokenString, err := token.SignedString(secretKey)
	if err != nil {
		return "", err
	}

	return tokenString, nil
}
