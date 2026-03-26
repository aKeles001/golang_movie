package routes

import (
	controller "github.com/aKeles001/golang_movie/Server/magic_movies_server/controllers"
	"github.com/gin-gonic/gin"
)

func SetupUnprotectedRoutes(router *gin.Engine) {
	router.GET("/movies", controller.GetMovies())
	router.POST("/register", controller.RegisterUser())
	router.POST("/login", controller.LoginUser())
}
