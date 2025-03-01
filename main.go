package main

import (
	"os"
	"log"
	"go-ecommerce/controllers"
	"go-ecommerce/database"
	"go-ecommerce/routes"
	"go-ecommerce/middleware"

	"github.com/gin-gonic/gin"
)

func main() {
	port := os.Getenv("PORT")

	if port == "" {	
		port = "8080"
	}

	app := controllers.NewApplication(database.ProductData(database.Client, "products"), database.UserData(database.Client, "users"))

	router := gin.New()
	router.Use(gin.Logger())

	routes.UserRoutes(router)
	router.Use(middleware.Authentication())

	router.GET("/addtocart", app.AddToCart())
	router.GET("removeitem", app.RemoveItem())
	router.GET("/cartcheckout", app.BuyFromCart())
	router.GET("/instantbuy", app.InstantBuy())

	log.Fatal(router.Run(":"+port))


}
