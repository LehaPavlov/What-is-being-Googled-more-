package request

import (
	"context"
	"fmt"
	"main/structurs"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var gameCollection *mongo.Collection
var client *mongo.Client
var ctx = context.TODO()

func init() {
	clientOptions := options.Client().ApplyURI("mongodb://localhost:27017/")
	var err error
	client, err = mongo.Connect(ctx, clientOptions)
	if err != nil {
		fmt.Println("Ошибка подключения к MongoDB:", err)
		panic(err)
	}
	if client != nil {
		gameCollection = client.Database("Game").Collection("DataGame")
	} else {
		panic("Не удалось инициализировать клиент MongoDB")
	}
}

func RandomItem(с *gin.Context) (structurs.Item, error) {
	if gameCollection == nil {
		return structurs.Item{}, fmt.Errorf("gameCollection is nil.  Check MongoDB connection in init()")
	}

	count, err := gameCollection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return structurs.Item{}, fmt.Errorf("failed to count documents: %w", err)
	}

	if count == 0 {
		return structurs.Item{}, fmt.Errorf("no items found in the database")
	}

	cursor, err := gameCollection.Aggregate(ctx, []bson.M{
		{"$sample": bson.M{"size": 1}},
	})
	if err != nil {
		return structurs.Item{}, fmt.Errorf("failed to aggregate: %w", err)
	}
	defer func() {
		if cursor != nil {
			cursor.Close(ctx)
		}
	}()

	var item structurs.Item
	if cursor.Next(ctx) {
		if err := cursor.Decode(&item); err != nil {
			return structurs.Item{}, fmt.Errorf("failed to decode item: %w", err)
		}
		return item, nil
	}

	return structurs.Item{}, fmt.Errorf("no item found after random selection")
}
