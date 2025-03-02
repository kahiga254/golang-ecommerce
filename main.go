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

	prodCollection := database.ProductData(database.Client, "products")
	userCollection := database.UserData(database.Client, "users")
	log.Println("Initializing prodCollection:", prodCollection.Name(), "in database:", prodCollection.Database().Name())
    log.Println("Initializing userCollection:", userCollection.Name(), "in database:", userCollection.Database().Name())
	app := controllers.NewApplication(prodCollection, userCollection)

	router := gin.New()
	router.Use(gin.Logger())

	routes.UserRoutes(router)
	router.GET("/", func(c *gin.Context) {
		c.JSON(200, gin.H{"message": "Welcome to the Go E-Commerce API!"})
	})
    
	protected := router.Group("/")
	protected.Use(middleware.Authentication())
	protected.GET("/cartcheckout", app.BuyFromCart())//
	protected.GET("/addtocart", app.AddToCart())//
	protected.GET("/removeitem", app.RemoveItem())//
	protected.GET("/listcart", app.GetItemFromCart())//
	protected.GET("/instantbuy", app.InstantBuy())//

	log.Fatal(router.Run(":"+port))


}
