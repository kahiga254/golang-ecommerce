package database

import (
	"context"
	"fmt"
	"log"
	"time"

	
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

func DBSet() *mongo.Client{

	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()

	client, err := mongo.Connect(ctx, options.Client().ApplyURI("mongodb+srv://adamskahiga:36596768Bantu.@cluster0.anyi0.mongodb.net/?retryWrites=true&w=majority&appName=Cluster0"))
	 
	if err != nil {
		log.Fatal(err)
	}

	err = client.Ping(ctx, nil)
	if err != nil {
		log.Println("failed to connect to mongo db")
		return nil
	}

	fmt.Println("Connected to MongoDB!")
	return client
}

 var Client *mongo.Client = DBSet()

func UserData(client *mongo.Client, collectionName string) *mongo.Collection{
	var collection *mongo.Collection = client.Database("ecommerce").Collection(collectionName)
	return collection
}

func ProductData(client *mongo.Client, collectionName string) *mongo.Collection{
	var productCollection *mongo.Collection = client.Database("ecommerce").Collection(collectionName)
	return productCollection
}