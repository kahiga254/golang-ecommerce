package database

import (
	"context"
	"errors"
	"go-ecommerce/models"
	"log"
	"time"

	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
)
var (
	ErrCantFindProduct = errors.New("product not found")
	ErrCantDecodeProduct = errors.New("can't find the product ")
	ErrUserIdIsNotValid = errors.New("user id is not valid")
	ErrCantUpdateUser = errors.New("cannot add this product to the cart")
	ErrCantRemoveItemCart = errors.New("cannot remove this item from the cart")
	ErrCantGetItem = errors.New("cannot get this item from the cart")
	ErrCantBuyCarItem = errors.New("cannot update the purchase")

)

func AddProductToCart(ctx context.Context, userCollection, prodCollection *mongo.Collection, productID primitive.ObjectID, userID string) error {
	searchfromdb, err := prodCollection.Find(ctx, bson.M{"_id": productID})
	if err != nil {
		log.Println("Database error while finding product:", err)
		return ErrCantFindProduct
	}
	defer searchfromdb.Close(ctx)

	// Check if a product was found
	if !searchfromdb.Next(ctx) {
		log.Println("No product found with ID:", productID)
		return ErrCantFindProduct
	}


	var productCart []models.ProductUser
	err = searchfromdb.All(ctx, &productCart)
	if err!=nil{
		log.Println("Error decoding product:", err)
		return ErrCantDecodeProduct
	}

	// Check if productCart is empty
	if len(productCart) == 0 {
		log.Println("No products found for ID:", productID)
		return ErrCantFindProduct
	}

	// Convert userID to ObjectID
	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Println("Invalid User ID format:", err)
		return ErrUserIdIsNotValid
	}

	// Ensure the user exists in the database
	filter := bson.M{"_id": id}
	userCount, err := userCollection.CountDocuments(ctx, filter)
	if err != nil {
		log.Println("Database error while checking user existence:", err)
		return ErrCantUpdateUser
	}
	if userCount == 0 {
		log.Println("User not found:", userID)
		return ErrCantUpdateUser
	} 

	//Update the user's cart
	update := bson.M{"$push": bson.M{"usercart": bson.M{"$each": productCart}},
	}

	result,err := userCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Println("Error updating user cart:", err)
		return ErrCantUpdateUser
	}

	// Check if the update was applied
	if result.MatchedCount == 0 {
		log.Println("User not found in database")
		return ErrCantUpdateUser
	}

	log.Println("Successfully added product to cart for user:", userID)
	return nil
}

func RemoveCartItem(ctx context.Context, prodColllection, userCollection *mongo.Collection, productID primitive.ObjectID, userID string) error {
	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Println(err)
		return ErrUserIdIsNotValid
	}
	filter := bson.D{{Key:"_id", Value: id}}
	update := bson.M{"$pull": bson.M{
		    "usercart": bson.M{"_id": productID},
	} }

	_, err =userCollection.UpdateMany(ctx, filter, update)
	if err != nil {
		log.Println(err)
		return ErrCantRemoveItemCart
	}
	return nil

}

func BuyItemFromCart(ctx context.Context, userCollection *mongo.Collection, userID string) error {
	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Println(err)
		return ErrUserIdIsNotValid
	}
	var getCartitems models.User
	var ordercart models.Order

	ordercart.Order_ID = primitive.NewObjectID()
	ordercart.Ordered_At = time.Now()
	ordercart.Order_Cart = make([]models.ProductUser, 0)
	ordercart.Payment_Method.COD = true

	unwind := bson.D{{Key:"$unwind", Value:bson.D{primitive.E{Key:"path", Value:"$usercart"}}}}
	grouping := bson.D{{Key:"$group", Value:bson.D{primitive.E{Key:"_id", Value:"$_id"}, {Key:"total", Value:bson.D{primitive.E{Key:"$sum", Value:"$usercart.price"}}}, {Key:"cart", Value:bson.D{primitive.E{Key:"$push", Value:"$usercart"}}}}}} 

	currentresults, err := userCollection.Aggregate(ctx, mongo.Pipeline{unwind, grouping})
	ctx.Done()
	if err != nil {
		panic(err)
	}

	 var getuser []bson.M
	 if err = currentresults.All(ctx, &getuser); err != nil {
		panic(err)
	 }
	 var total_price int32

	 for _, user_item := range getuser{
		price, ok := user_item["total"].(int64)
		if !ok {
			log.Println("Error converting total price to int64")
			return ErrCantBuyCarItem
		}

		total_price = int32(price)
	 }
	 ordercart.Price = new(uint64)
	 *ordercart.Price = uint64(total_price)

	 filter := bson.D{primitive.E{Key:"_id"}}
	 update := bson.D{{Key:"$push", Value: bson.D{primitive.E{Key:"orders", Value:ordercart}}}}
	 _, err = userCollection.UpdateMany(ctx, filter, update)
	 if err != nil {
		 log.Println(err)
	 }

	 err = userCollection.FindOne(ctx, bson.D{primitive.E{Key:"_id", Value:id}}).Decode(&getCartitems)
	 if err != nil {
		 log.Println(err)
	 }

	 filter2 := bson.D{primitive.E{Key:"_id", Value:id}}
	 update2 := bson.M{"$push":bson.M{"orders.&[].order_list": bson.M{"$each": getCartitems.UserCart}}}
	 _, err = userCollection.UpdateOne(ctx, filter2, update2)
	 if err != nil {
		 log.Println(err)
		}

		usercart_empty := make([]models.ProductUser, 0)
		filter3 := bson.D{primitive.E{Key:"_id", Value: id}}
		update3 := bson.D{{Key:"$set", Value: bson.D{primitive.E{Key:"usercart", Value:usercart_empty}}}}
		_, err = userCollection.UpdateOne(ctx, filter3, update3)
		if err != nil {
			return ErrCantBuyCarItem
		}
		return nil

}

func InstantBuyer(ctx context.Context, prodCollection, userCollection *mongo.Collection, productID primitive.ObjectID, userID string) error {
	log.Println("Using prodCollection:", prodCollection.Name(), "in database:", prodCollection.Database().Name())
    log.Println("Fetching product with ID:", productID.Hex())
	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Println(err)
		return ErrUserIdIsNotValid
	}
	var product_details models.ProductUser
	err = prodCollection.FindOne(ctx, bson.M{"_id": productID}).Decode(&product_details)
	log.Println("Fetching product with ID:", productID.Hex())
	if err != nil {
		log.Println("Error fetching product details:", err)
		return err
	}

	// Create order details
	order_detail := models.Order{
		Order_ID:       primitive.NewObjectID(),
		Ordered_At:     time.Now(),
		Order_Cart:     []models.ProductUser{product_details},
		Payment_Method: models.Payment{COD: true},
		Price:          product_details.Price,
	}


    // Update user orders
	filter := bson.M{"_id": id}
	update := bson.M{"$push": bson.M{"orders": order_detail}}

	updateResult,err := userCollection.UpdateOne(ctx, filter, update)
	if err!=nil{
		log.Println("Error updating user orders:", err)
		return err
	}
	if updateResult.MatchedCount == 0 {
		log.Println("Warning: No matching user found for update")
		return errors.New("user not found")
	}

	// Add product to order_list inside orders array
	update2 := bson.M{"$push":bson.M{"orders.$[].order_list": product_details,},}
	updateResult2, err := userCollection.UpdateOne(ctx, filter, update2)
	if err != nil {
		log.Println("Error adding product to order_list:", err)
		return err
	}

	if updateResult2.MatchedCount == 0 {
		log.Println("Warning: No matching user found for order list update")
	}

	log.Println("Instant buy completed successfully!")
	return nil	
}