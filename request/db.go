package request

import (
	"context"
	"fmt"
	"log"
	"main/structurs"
	"math/rand"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson"
	"go.mongodb.org/mongo-driver/bson/primitive"
	"go.mongodb.org/mongo-driver/mongo"
	"go.mongodb.org/mongo-driver/mongo/options"
)

var gameCollection *mongo.Collection
var client *mongo.Client
var ctx = context.TODO()
var once sync.Once
var usedItemIDsMu sync.Mutex
var usedItemIDs = make(map[primitive.ObjectID]bool)

func init() {
	const uri = "mongodb://localhost:27017/"
	once.Do(func() {
		clientOptions := options.Client().ApplyURI(uri)
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
		rand.Seed(time.Now().UnixNano())
	})
}

func RandomItem(c *gin.Context) (structurs.Item, error) {
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
func Replacement(c *gin.Context) (structurs.Item, error) {
	count, err := gameCollection.EstimatedDocumentCount(c, nil)
	if err != nil {
		return structurs.Item{}, fmt.Errorf("failed to get document count: %w", err)
	}

	if count == 0 {
		return structurs.Item{}, fmt.Errorf("collection is empty")
	}

	randomIndex := rand.Intn(int(count))

	findOptions := options.FindOne().SetSkip(int64(randomIndex))
	var item structurs.Item
	err = gameCollection.FindOne(c, bson.M{}, findOptions).Decode(&item)
	if err != nil {
		return structurs.Item{}, fmt.Errorf("failed to find replacement item: %w", err)
	}

	return item, nil
}

func NextRoundLeft(c *gin.Context, item1 string, item2 string) (structurs.Item, structurs.Item, error) {
	return nextRound(c, item1, item2, true)
}

func NextRoundRight(c *gin.Context, item1 string, item2 string) (structurs.Item, structurs.Item, error) {
	return nextRound(c, item1, item2, false)
}

func nextRound(c *gin.Context, item1 string, item2 string, left bool) (structurs.Item, structurs.Item, error) {
	var newItem1 structurs.Item
	var newItem2 structurs.Item
	ItemID1, err := primitive.ObjectIDFromHex(item1)
	ItemID2, err := primitive.ObjectIDFromHex(item2)
	if err != nil {
		log.Println("Ошибка при преобразовании ObjectID из hex:", err)
		return structurs.Item{}, structurs.Item{}, fmt.Errorf("Ошибка при преобразовании ObjectID из hex: %w", err)
	}

	filter := bson.M{"_id": ItemID1}
	err = gameCollection.FindOne(c, filter).Decode(&newItem1)
	if err != nil {
		log.Println(err)
		return structurs.Item{}, structurs.Item{}, fmt.Errorf("failed to find item1: %w", err)
	}

	filter = bson.M{"_id": ItemID2}
	err = gameCollection.FindOne(c, filter).Decode(&newItem2)
	if err != nil {
		log.Println(err)
		return structurs.Item{}, structurs.Item{}, fmt.Errorf("failed to find item2: %w", err)
	}

	//Determine which item has higher popularity
	var replacementItem structurs.Item
	if (left && newItem1.Popularity > newItem2.Popularity) || (!left && newItem2.Popularity > newItem1.Popularity) {
		replacementItem, err = Replacement(c)
		if err != nil {
			log.Println("Ошибка при получении Replacement:", err)
			return structurs.Item{}, structurs.Item{}, fmt.Errorf("failed to get replacement item: %w", err)
		}
		if left {
			return newItem1, replacementItem, nil
		} else {
			return newItem2, replacementItem, nil
		}

	} else {
		return structurs.Item{}, structurs.Item{}, nil
	}
}
