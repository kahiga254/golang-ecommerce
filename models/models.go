package models

import (
	"time"

	"go.mongodb.org/mongo-driver/bson/primitive"
)

type User struct {

	ID				primitive.ObjectID `json:"_id" bson:"_id,omitempty"`
	First_Name		*string				`json:"first_name" validate:"required,min=2,max=30"`
	Last_Name		*string				`json:"last_name" validate:"required,min=2,max=30"`
	Password		*string				`json:"password" validate:"required,min=6"`
	Email			*string				`json:"email" validate:"email,required"`
	Phone			*string				`json:"phone" validate:"required"`
	Token			*string				`json:"token"`
	Refresh_Token	*string				`json:"refresh_token"`
	Created_At	     time.Time			`json:"created_at"`
	Updated_At       time.Time			`json:"updated_at"`
	User_ID		     *string			`json:"user_id"`
	UserCart		 []ProductUser		`json:"user_cart" bson:"usercart"`
	Address_Details	 []Address			`json:"address_details" bson:"address"`
	Order_Status	 []Order			`json:"order_status" bson:"orders"`
}


type Product struct {
	Product_ID		 primitive.ObjectID	 `json:"_id" bson:"_id,omitempty"`
	Product_Name	 *string			`Json:"product_name"`
	Price			 *uint64			`json:"price"`
	Rating			 *uint8				`json:"rating"`
	Image			  *string 			`json:"image"`

}

type ProductUser struct {
	Product_ID       primitive.ObjectID	`json:"_id" bson:"_id,omitempty"`
	Procduct_Name	 *string			`Json:"product_name" bson:"product_name"`
	Price			 *uint64			`json:"price" bson:"price"`
	Rating			 *uint8				`json:"rating" bson:"rating"`
	Image			  *string			`json:"image" bson:"image"`
}

type Address struct {
	Address_id       primitive.ObjectID  `json:"_id" bson:"_id,omitempty"`
	House			 *string			 `json:"house_name" bson:"house_name"`
	Street			 *string			 `json:"street_name" bson:"street_name"`
	City			 *string			 `json:"city" bson:"city_name"`
	Pincode			 *string			`json:"pin_code" bson:"pin_code"`
}

type Order struct {
	Order_ID       primitive.ObjectID	`json:"_id" bson:"_id,omitempty"`
	Order_Cart	   []ProductUser		`json:"order_list" bson:"order_list"`
	Ordered_At		time.Time			`json:"ordered_at" bson:"ordered_at"`
	Price			 *uint64			`json:"total_price" bson:"total_price"`
	Discount		 *uint64			`json:"discount" bson:"discount"`
	Payment_Method	 Payment			``

}

type Payment struct {
	Digital	bool
	COD		bool
}
