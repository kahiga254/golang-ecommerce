package controllers

import (
	"context"
	"errors"
	"go-ecommerce/database"
	"go-ecommerce/models"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	
)

type Application struct{
	prodCollection *mongo.Collection
	userCollection *mongo.Collection
}

func NewApplication(prodCollection *mongo.Collection, userCollection *mongo.Collection) *Application{
	return &Application{
		prodCollection: prodCollection,
		userCollection: userCollection,
	}
}

func (app *Application) AddToCart() gin.HandlerFunc{
	return func (c *gin.Context){	
	productQueryID := c.Query("id")
	if productQueryID == "" {
		log.Println("Product ID is required")
	 c.JSON(http.StatusBadRequest, errors.New("product ID is required"))
		return
	}

	userQueryID := c.Query("user_id")
	
	if userQueryID == "" {
		log.Println("User ID is required")
		 c.JSON(http.StatusBadRequest, errors.New("user ID is required"))
		return
	}

	productID, err := primitive.ObjectIDFromHex(productQueryID)
	if err != nil {
		log.Println("Invalid Product ID format",err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	userID := userQueryID

	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	log.Println("Product ID:", productQueryID)
	log.Println("User ID:", userQueryID)
	log.Println("Calling AddProductToCart...")

	err = database.AddProductToCart(ctx, app.userCollection, app.prodCollection, productID, userID)
	if err != nil {
		log.Println("Error adding product to cart:", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to add product to cart"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "Product added to cart"})
 }
}

func (app *Application) RemoveItem() gin.HandlerFunc{
		return func(c *gin.Context){
			productQueryID := c.Query("id")
			if productQueryID == "" {
				log.Println("Product ID is required")
				_ = c.AbortWithError(http.StatusBadRequest, errors.New("product ID is required"))
				return
			}
		
			userQueryID := c.Query("user_id")
			if userQueryID == "" {
				log.Println("User ID is required")
				_ = c.AbortWithError(http.StatusBadRequest, errors.New("user ID is required"))
				return
			}
		
			productID, err := primitive.ObjectIDFromHex(productQueryID)
		
			if err != nil {
				log.Println(err)
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
		
			var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
			defer cancel()	

			err = database.RemoveCartItem(ctx, app.userCollection, app.prodCollection,  productID, userQueryID) 
			if err != nil {
				c.IndentedJSON(http.StatusInternalServerError, err)
				return
			}
			c.IndentedJSON(http.StatusOK, "Product removed from cart")
		}
}

func (app *Application) GetItemFromCart() gin.HandlerFunc{
	return func(c *gin.Context){
		user_id := c.Query("id")	

		if user_id == "" {
			c.Header("Content-Type", "application/json")
			c.JSON(http.StatusNotFound, gin.H{"error": "User ID is required"})
			c.Abort()
			return
		}

		userid, _ := primitive.ObjectIDFromHex(user_id)

		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var filledcart models.User
		err := UserCollection.FindOne(ctx, bson.D{primitive.E{Key:"_id", Value: userid}}).Decode(&filledcart)

		if err != nil {
			log.Println(err)
			c.IndentedJSON(500, "not found")
			return
		}

		filter_match := bson.D{{Key:"$match", Value: bson.D{primitive.E{Key:"_id", Value: userid}}}}
		unwind := bson.D{{Key:"$unwind", Value:bson.D{primitive.E{Key:"path", Value:"$usercart"}}}}
		grouping := bson.D{{Key:"$group", Value: bson.D{primitive.E{Key:"_id", Value:"$_id"}, {Key:"total", Value: bson.D{primitive.E{Key:"$sum", Value:"$usercart.price"}}}}}}

		pointcursor, err := UserCollection.Aggregate(ctx, mongo.Pipeline{filter_match, unwind, grouping})
		if err != nil {
			log.Println(err)
		}

		var listing []bson.M
		if err = pointcursor.All(ctx, &listing); err != nil {
			log.Println(err)
			c.AbortWithStatus(http.StatusInternalServerError)
		}
		for _, json := range listing {
			c.IndentedJSON(200, json["total"])
			c.IndentedJSON(200, filledcart.UserCart)
		}
		ctx.Done()

	}

}

func (app *Application) BuyFromCart() gin.HandlerFunc{

	return func(c *gin.Context){
		userQueryID := c.Query("id")
		
		if userQueryID == "" {
			log.Println("User ID is required")
			_ = c.AbortWithError(http.StatusBadRequest, errors.New("user ID is required"))
		}

		var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()

		err := database.BuyItemFromCart(ctx, app.userCollection, userQueryID)
		if err != nil {
			c.IndentedJSON(http.StatusInternalServerError, err)
		}
		c.IndentedJSON(http.StatusOK, "Product bought")
	}
}

func (app *Application) InstantBuy() gin.HandlerFunc{

	return func(c *gin.Context){
		productQueryID := c.Query("id")
			if productQueryID == "" {
				log.Println("Product ID is required")
				_ = c.AbortWithError(http.StatusBadRequest, errors.New("product ID is required"))
				return
			}
		
			userQueryID := c.Query("user_id")
			if userQueryID == "" {
				log.Println("User ID is required")
				_ = c.AbortWithError(http.StatusBadRequest, errors.New("user ID is required"))
				return
			}
		
			productID, err := primitive.ObjectIDFromHex(productQueryID)
		
			if err != nil {
				log.Println(err)
				c.AbortWithStatus(http.StatusInternalServerError)
				return
			}
		
			var ctx, cancel = context.WithTimeout(context.Background(), 5*time.Second)
			defer cancel()	
			err = database.InstantBuyer(ctx, app.prodCollection, app.userCollection, productID, userQueryID)
			if err != nil {
				c.IndentedJSON(http.StatusInternalServerError, err)
				return
			}
			c.IndentedJSON(http.StatusOK, "Product bought")
	}
}