package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	"go.mongodb.org/mongo-driver/v2/mongo"

	database "github.com/aKeles001/golang_movie/Server/magic_movies_server/database"
	routes "github.com/aKeles001/golang_movie/Server/magic_movies_server/routes"
)

func main() {

	router := gin.Default()
	err := godotenv.Load(".env")
	if err != nil {
		fmt.Println("Error loading .env file")
	}

	var client *mongo.Client = database.Connect()

	routes.SetupUnprotectedRoutes(router, client)
	routes.SetupProtectedRoutes(router, client)

	if err := router.Run(":8080"); err != nil {
		fmt.Println(err)
	}
}
