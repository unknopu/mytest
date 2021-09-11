package main

import (
	"mongodb-go/controller"
	// "mongodb-go/models"
	// "context"
	// "fmt"
	// "reflect"
	// "time"
	"github.com/gin-gonic/gin"
	// "go.mongodb.org/mongo-driver/mongo"
	// "go.mongodb.org/mongo-driver/mongo/options"
)

func main() {
	r := gin.Default()
	r.LoadHTMLGlob("cst/*.html")

	r.GET("/", controller.DisplayAllUser)
	r.POST("/", controller.Registoration)
	r.GET("/search", controller.Search)
	r.GET("/ranking", controller.Ranking)
	r.POST("challenge/", controller.Challenge)

	r.Run()
}
