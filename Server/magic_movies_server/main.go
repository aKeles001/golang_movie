package main

import (
	"fmt"

	"github.com/gin-gonic/gin"

	controller "github.com/aKeles001/golang_movie/Server/magic_movies_server/controllers"
)

func main() {

	router := gin.Default()

	router.GET("/hello", func(c *gin.Context) {
		c.String(200, "Hello, World!")
	})

	router.GET("/movies", controller.GetMovies())
	router.GET("/movies/:imdb_id", controller.GetMovie())
	router.POST("/addmovie", controller.AddMovie())

	if err := router.Run(":8080"); err != nil {
		fmt.Println(err)
	}
}
