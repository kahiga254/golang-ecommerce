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

func AddProductToCart(ctx context.Context, prodCollection, userCollection *mongo.Collection, productID primitive.ObjectID, userID string) error {
	searchfromdb, err := prodCollection.Find(ctx, bson.M{"_id": productID})
	if err != nil {
		log.Println(err)
		return ErrCantFindProduct
	}
	var productCart []models.ProductUser
	err = searchfromdb.All(ctx, &productCart)
	if err!=nil{
		log.Println(err)
		return ErrCantDecodeProduct
	}
	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Println(err)
		return ErrUserIdIsNotValid
	}

	filter := bson.D{primitive.E{Key:"_id", Value: id} }
	update := bson.D{{Key:"$push", Value: bson.D{primitive.E{Key:"usercart", Value: bson.D{{Key:"$each", Value:productCart}}}}}}

	_,err = userCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Println(err)
		return ErrCantUpdateUser
	}
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
		price := user_item["total"]
		total_price = price.(int32)
	 }
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
	id, err := primitive.ObjectIDFromHex(userID)
	if err != nil {
		log.Println(err)
		return ErrUserIdIsNotValid
	}
	var product_details models.ProductUser
	var order_detail models.Order

	order_detail.Order_ID = primitive.NewObjectID()
	order_detail.Ordered_At = time.Now()
	order_detail.Order_Cart = make([]models.ProductUser, 0)
	order_detail.Payment_Method.COD = true
	err = prodCollection.FindOne(ctx, bson.D{primitive.E{Key:"_id", Value: productID}}).Decode(&product_details)
	if err!=nil{
		log.Println(err)
	}
	order_detail.Price = product_details.Price

	filter := bson.D{primitive.E{Key:"_id", Value:id} }
	update := bson.D{{Key:"$push", Value: bson.D{primitive.E{Key:"orders", Value:order_detail}}}}

	_, err = userCollection.UpdateOne(ctx, filter, update)
	if err != nil {
		log.Println(err)
	}

	filter2 := bson.D{primitive.E{Key:"_id", Value:id}}
	update2 := bson.M{"$push":bson.M{"orders.&[].order_list": bson.M{"$each": product_details}}}

	_, err = userCollection.UpdateOne(ctx, filter2, update2)
	if err != nil {
		log.Println(err)
	}
	return nil
	
	
}