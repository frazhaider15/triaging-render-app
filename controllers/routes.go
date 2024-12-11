package controllers

import (
	"time"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

func GetGinRoutes(r *gin.Engine) {

	r.Use(cors.New(cors.Config{
		AllowOrigins:     []string{"*"},
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		MaxAge:           12 * time.Hour,
	}))

	// For testing purpose
	r.GET("", Testing)

	// User information
	render := r.Group("render")
	{
		render.POST("flow", WorkFlowRendering)
		render.POST("previous-flow", PreviousForm)
		render.GET("check-last-node", CheckLastNode)
	}
}