package main

import (
	"fmt"
	"os"

	"31g.co.uk/triaging/controllers"
	"31g.co.uk/triaging/db"
	"github.com/gin-gonic/gin"
)

func main() {
	//config.LoadEnvVariables()

	db.LoadDataJson()

	r := gin.Default()
	controllers.GetGinRoutes(r)

	fmt.Println("Starting on http://0.0.0.0:" + os.Getenv("PORT"))
	r.Run()
}
