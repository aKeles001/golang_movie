package database

import (
	"log"
	"os"

	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"
	"go.mongodb.org/mongo-driver/v2/mongo/options"
)

func Connect() *mongo.Client {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Error loading .env file")
	}
	MongoDB := os.Getenv("MONGO_URI")
	if MongoDB == "" {
		log.Fatal("MONGO_URI not set in .env file")
	}
	clientOptions := options.Client().ApplyURI(MongoDB)
	client, err := mongo.Connect(nil, clientOptions)
	if err != nil {
		return nil
	}
	return client
}

func OpenCollection(collectionName string, client *mongo.Client) *mongo.Collection {
	err := godotenv.Load(".env")
	if err != nil {
		log.Println("Error loading .env file")
	}
	databaseName := os.Getenv("DATABASE_NAME")

	collection := client.Database(databaseName).Collection(collectionName)
	if collection == nil {
		return nil
	}
	return collection
}
