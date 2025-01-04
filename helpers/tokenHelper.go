package helper

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/Maski0/jwt-template-golang/database"
	"github.com/golang-jwt/jwt/v5"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

type SignedDetails struct {
	Email     string
	FirstName string
	LastName  string
	Uid       string
	User_type string
	jwt.RegisteredClaims
}

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")

var SECRET_KEY string = os.Getenv("SECRET_KEY")

func GenerateAllTokens(email string, firstName string, lastName string, userType string, uid string) (signedToken string, signedRefreshToken string, err error) {
	claims := &SignedDetails{
		Email:     email,
		FirstName: firstName,
		LastName:  lastName,
		Uid:       uid,
		User_type: userType,
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Local().Add(time.Hour * time.Duration(24))), // This is 24 Hr token expiree
		},
	}
	refreshClaims := &SignedDetails{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Local().Add(time.Hour * time.Duration(168))),
		},
	}
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, claims).SignedString([]byte(SECRET_KEY))
	if err != nil {
		log.Panic(err)
		return
	}

	refreshToken, err := jwt.NewWithClaims(jwt.SigningMethodHS256, refreshClaims).SignedString([]byte(SECRET_KEY))

	if err != nil {
		log.Panic(err)
		return
	}

	return token, refreshToken, nil
}

func UpdateAllTokens(signedToken string, signedRefreshToken string, userId string) {
	var ctxWithTimeout, cancel = context.WithTimeout(context.Background(), 100*time.Second)

	var updateObj bson.D

	updateObj = append(updateObj, bson.E{Key: "token", Value: signedToken})
	updateObj = append(updateObj, bson.E{Key: "refresh_token", Value: signedRefreshToken})

	Updated_at, _ := time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
	updateObj = append(updateObj, bson.E{Key: "updated_at", Value: Updated_at})

	filter := bson.M{"user_id": userId}
	opt := options.UpdateOne().SetUpsert(true)

	_, err := userCollection.UpdateOne(
		ctxWithTimeout,
		filter,
		bson.D{
			{Key: "$set", Value: updateObj},
		},
		opt,
	)
	defer cancel()

	if err != nil {
		log.Panic(err)
		return
	}
}
