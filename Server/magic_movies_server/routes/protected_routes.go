package routes

import (
	controller "github.com/aKeles001/golang_movie/Server/magic_movies_server/controllers"
	middleware "github.com/aKeles001/golang_movie/Server/magic_movies_server/middleware"
	"github.com/gin-gonic/gin"
)

func SetupProtectedRoutes(router *gin.Engine) {
	router.Use(middleware.AuthMiddleware())
	router.GET("/movie/:imdb_id", controller.GetMovie())
	router.POST("/addmovie", controller.AddMovie())
}
