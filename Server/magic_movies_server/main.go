package main

import (
	"fmt"

	"github.com/gin-gonic/gin"

	routes "github.com/aKeles001/golang_movie/Server/magic_movies_server/routes"
)

func main() {

	router := gin.Default()

	routes.SetupUnprotectedRoutes(router)
	routes.SetupProtectedRoutes(router)

	if err := router.Run(":8080"); err != nil {
		fmt.Println(err)
	}
}
