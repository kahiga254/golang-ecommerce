package controllers

import (
	"context"
	"fmt"
	"go-ecommerce/database"
	"go-ecommerce/models"
	generate"go-ecommerce/tokens"
	"log"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"golang.org/x/crypto/bcrypt"
)

var UserCollection *mongo.Collection = database.UserData(database.Client, "users")
var ProductCollection *mongo.Collection = database.ProductData(database.Client, "products")
var	validate = validator.New()

func HashPassword(password string) string{
	bytes, err := bcrypt.GenerateFromPassword([]byte(password), 14)
	if err != nil {
		log.Panic(err)
	}
	return string(bytes)

}

func VerifyPassword(userPassword string, givenPassword string) (bool,string) {
 err := bcrypt.CompareHashAndPassword([]byte(userPassword), []byte(givenPassword))
 valid := true
 msg := ""

 if err != nil {
	 msg = "Password is incorrect"
	 valid = false
 }
 return valid, msg
}

func SignUp() gin.HandlerFunc{

	return func(c *gin.Context){
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user models.User
		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		validationErr := validate.Struct(user)
		if validationErr != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": validationErr.Error()})
			return
		}

		count, err := UserCollection.CountDocuments(ctx, bson.M{"email": user.Email})
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}
		if count > 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "User already exists"})
			return
		}

		count, err = UserCollection.CountDocuments(ctx, bson.M{"phone": user.Phone})

		if err != nil {
			log.Panic(err)
			c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
			return
		}	
		
		if count> 0 {
			c.JSON(http.StatusConflict, gin.H{"error": "Phone number already exists"})
			return
		}
		
		password := HashPassword(*user.Password)
		user.Password = &password

		user.Created_At = time.Now()
		user.Updated_At = time.Now()
		userID := user.ID.Hex()
		user.User_ID = &userID
		token, refreshtoken, _ := generate.TokenGenerator(*user.Email, *user.First_Name, *user.Last_Name, *user.User_ID)
		user.Token = &token
		user.Refresh_Token = &refreshtoken
		user.UserCart = make([]models.ProductUser, 0)
		user.Address_Details = make([]models.Address, 0)
		user.Order_Status = make([]models.Order, 0) 
		_, inserterr := UserCollection.InsertOne(ctx, user)
		if inserterr != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "the user did not get created"})
			return
		}
		defer cancel()

		c.JSON(http.StatusCreated,  "User created successfully")

	}
}

func Login() gin.HandlerFunc{
	return func(c *gin.Context) {
		var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
		defer cancel()

		var user models.User
		var founduser models.User

		if err := c.BindJSON(&user); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		err := UserCollection.FindOne(ctx, bson.M{"email": user.Email}).Decode(&founduser)

		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "User not found"})
			return
		}

		PasswordIsValid, msg := VerifyPassword(*founduser.Password, *user.Password)

		if !PasswordIsValid {
			c.JSON(http.StatusUnauthorized, gin.H{"error": msg})
			fmt.Println(msg)
			return
		}
		token, refreshToken, _ := generate.TokenGenerator(*founduser.Email, *founduser.First_Name, *founduser.Last_Name, *founduser.User_ID)

		generate.UpdateAllTokens( token, refreshToken, *founduser.User_ID)

		c.JSON(http.StatusFound, founduser)
	}
}

//func ProductViewerAdmin() gin.HandlerFunc{

//}

func SearchProduct()gin.HandlerFunc{
	return func(c *gin.Context){
	var productlist []models.Product
	var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
	defer cancel()

	cursor, err := ProductCollection.Find(ctx, bson.D{{}})
	if err != nil {
		c.IndentedJSON(http.StatusInternalServerError, "Something went wron please try again")
		return
	}

	err = cursor.All(ctx, &productlist)
	
	if err != nil {
		log.Println(err)
		c.AbortWithStatus(http.StatusInternalServerError)
		return
	}

	if err := cursor.Err(); err != nil {
		log.Println(err)
		c.IndentedJSON(400, "invalid")
		return
	}

	c.IndentedJSON(200, productlist)
	}

}

func SearchProductByQuery() gin.HandlerFunc{
	return func(c *gin.Context){
			var searchProducts []models.Product
			queryParam := c.Query("name")

			if queryParam == "" {
				log.Println("query is empty")
				c.Header("Content-Type", "application/json")
				c.JSON(http.StatusNotFound, gin.H{"error": "Invalid search index"})
				c.Abort()
				return
			}

			var ctx, cancel = context.WithTimeout(context.Background(), 100*time.Second)
			defer cancel()

			searchquerydb, err := ProductCollection.Find(ctx, bson.M{"name": bson.M{"$regex": queryParam}})

			if err != nil {
				c.IndentedJSON(404, "No product found")
				return
			}

			err = searchquerydb.All(ctx, &searchProducts)
			if err != nil {
				log.Println(err)
				c.IndentedJSON(400, "invalid")
				return
			}

			defer searchquerydb.Close(ctx)

			if err := searchquerydb.Err(); err != nil {
				log.Println(err)
				c.IndentedJSON(400, "invalid request")
				return
			}
			c.IndentedJSON(200, searchProducts)
	}
}
