package routes

import (
	controller "github.com/aKeles001/golang_movie/Server/magic_movies_server/controllers"
	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/v2/mongo"
)

func SetupUnprotectedRoutes(router *gin.Engine, client *mongo.Client) {
	router.GET("/movies", controller.GetMovies(client))
	router.POST("/register", controller.RegisterUser(client))
	router.POST("/login", controller.LoginUser(client))

}
