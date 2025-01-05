package controllers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strconv"
	"time"

	"github.com/Maski0/jwt-template-golang/database"
	helper "github.com/Maski0/jwt-template-golang/helpers"
	"github.com/Maski0/jwt-template-golang/models"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"go.mongodb.org/mongo-driver/v2/bson"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"golang.org/x/crypto/bcrypt"
)

var userCollection *mongo.Collection = database.OpenCollection(database.Client, "user")
var validate = validator.New()

func HashPassword(password string) string {
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}
	return string(bytes)
}

func VerifyPassword(userPassword string, providedPassword string) (bool, string) {
	err := bcrypt.CompareHashAndPassword([]byte(providedPassword), []byte(userPassword))
	check := true
	msg := ""

	if err != nil {
		msg = fmt.Sprintf("Password is Incorrect")
		check = false
	}
	return check, msg
}

func SignUp() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var ctxWithTimeout, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var newUser models.User
		if err := ctx.BindJSON(&newUser); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		validationErr := validate.Struct(newUser)
		if validationErr != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}
		emailCount, err := userCollection.CountDocuments(ctxWithTimeout, bson.M{"email": newUser.Email})
		defer cancel()
		if err != nil {
			log.Panic(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error occured while checking the user"})
		}
		password := HashPassword(*newUser.Password)
		newUser.Password = &password
		phoneCount, err := userCollection.CountDocuments(ctxWithTimeout, bson.M{"phone": newUser.Phone})
		defer cancel()
		if err != nil {
			log.Panic(err)
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "Error occured while checking the user"})
		}

		if phoneCount > 0 && emailCount > 0 {
			ctx.JSON(http.StatusConflict, gin.H{"error": "Email & phone number already exists!!"})
		}
		if phoneCount > 0 {
			ctx.JSON(http.StatusConflict, gin.H{"error": "phone number already exists!!"})
		}
		if emailCount > 0 {
			ctx.JSON(http.StatusConflict, gin.H{"error": "Email already exists!!"})
		}
		newUser.Created_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		newUser.Updated_at, _ = time.Parse(time.RFC3339, time.Now().Format(time.RFC3339))
		newUser.ID = bson.NewObjectID()
		newUser.User_id = newUser.ID.Hex()
		token, refreshToken, _ := helper.GenerateAllTokens(*newUser.Email, *newUser.FirstName, *newUser.LastName, *newUser.UserType, newUser.User_id)
		newUser.Token = &token
		newUser.Refresh_token = &refreshToken

		resultInsertionNumber, insertErr := userCollection.InsertOne(ctxWithTimeout, newUser)
		if insertErr != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "user item was not created"})
			return
		}
		defer cancel()
		ctx.JSON(http.StatusOK, resultInsertionNumber)
	}
}

func Login() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		var ctxWithTimeout, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()
		var existingUser models.User
		var foundUser models.User

		if err := ctx.BindJSON(&existingUser); err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}

		err := userCollection.FindOne(ctx, bson.M{"email": existingUser.Email}).Decode(&foundUser)
		defer cancel()
		if err != nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "email or password is incorrect"})
			return
		}

		passwordIsValid, msg := VerifyPassword(*existingUser.Password, *foundUser.Password)
		defer cancel()
		if passwordIsValid != true {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": msg})
			return
		}

		if foundUser.Email == nil {
			ctx.JSON(http.StatusNotFound, gin.H{"error": "user not found !!!"})
			return
		}
		token, refreshToken, _ := helper.GenerateAllTokens(*foundUser.Email, *foundUser.FirstName, *foundUser.LastName, *foundUser.UserType, foundUser.User_id)
		helper.UpdateAllTokens(token, refreshToken, foundUser.User_id)
		err = userCollection.FindOne(ctxWithTimeout, bson.M{"user_id": foundUser.User_id}).Decode(&foundUser)

		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, foundUser)
	}
}

func GetUsers() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		log.Println("Keys in context:", ctx.Keys)
		if err := helper.CheckUserType(ctx, "ADMIN"); err != nil {
			ctx.JSON(http.StatusUnauthorized, gin.H{"error": err.Error()})
			return
		}
		var ctxWithTimeout, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		recordPerPage, recordPageErr := strconv.Atoi(ctx.Query("recordPerPage"))
		if recordPageErr != nil || recordPerPage < 1 {
			recordPerPage = 10
		}
		page, pageErr := strconv.Atoi(ctx.Query("page"))
		if pageErr != nil || page < 1 {
			page = 1
		}
		startIndex := (page - 1) * recordPerPage
		var err error
		startIndex, err = strconv.Atoi(ctx.Query("startIndex"))

		matchStage := bson.D{{Key: "$match", Value: bson.D{{}}}}
		groupStage := bson.D{
			{Key: "$group", Value: bson.D{{Key: "_id", Value: bson.D{{Key: "_id", Value: "null"}}},
				{Key: "total_count", Value: bson.D{{Key: "$sum", Value: 1}}},
				{Key: "data", Value: bson.D{{Key: "$push", Value: "$$ROOT"}}},
			}}}
		projectStage := bson.D{
			{Key: "$project", Value: bson.D{
				{Key: "_id", Value: 0},
				{Key: "total_count", Value: 1},
				{Key: "user_items", Value: bson.D{{Key: "$slice", Value: []interface{}{"$data", startIndex, recordPerPage}}}},
			}},
		}
		result, err := userCollection.Aggregate(ctxWithTimeout, mongo.Pipeline{
			matchStage, groupStage, projectStage})
		defer cancel()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": "error occured while listing users"})
		}
		var allUsers []bson.M
		if err = result.All(ctxWithTimeout, &allUsers); err != nil {
			log.Fatal(err)
		}
		ctx.JSON(http.StatusOK, &allUsers[0])
	}
}

func GetUser() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		userId := ctx.Param("user_id")

		if err := helper.MatchUserTypeToUid(ctx, userId); err != nil {
			ctx.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
		var con, cancle = context.WithTimeout(context.Background(), 100*time.Second)
		var user models.User
		err := userCollection.FindOne(con, bson.M{"user_id": userId}).Decode(&user)
		defer cancle()
		if err != nil {
			ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		ctx.JSON(http.StatusOK, user)
	}
}
