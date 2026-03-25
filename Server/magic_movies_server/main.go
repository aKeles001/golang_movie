package main

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

func main() {

	router := gin.Default()

	router.GET("/hello", func(c *gin.Context) {
		c.String(200, "Hello, World!")
	})

	router.GET("/movies", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"message": "GetMovies endpoint",
		})
	})

	if err := router.Run(":8080"); err != nil {
		fmt.Println(err)
	}
}
