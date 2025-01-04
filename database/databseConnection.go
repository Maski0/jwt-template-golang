package database

import (
	"context"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func DBinstance() *mongo.Client {
	err := godotenv.Load(".env")
	if err != nil {
		log.Fatal("Error Loading .env file")
	}

	mongoDB := os.Getenv("MONGODB_URL")

	client, err := mongo.Connect(options.Client().ApplyURI(mongoDB))
	if err != nil {
		log.Fatal(err)
	}

	_, cancle := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancle()

	fmt.Println("Connected To MongoDB")

	return client
}

var Client *mongo.Client = DBinstance()

func OpenCollection(client *mongo.Client, collectionName string) *mongo.Collection {
	var collection *mongo.Collection = client.Database("cluster01").Collection(collectionName)
	return collection
}
