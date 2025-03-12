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
	var item structurs.Item

	for {
		randomItem, err := RandomItem(c)
		if err != nil {
			return structurs.Item{}, err
		}
		usedItemIDsMu.Lock()
		_, alreadyUsed := usedItemIDs[randomItem.ID]
		if !alreadyUsed {
			usedItemIDs[randomItem.ID] = true
			usedItemIDsMu.Unlock()
			item = randomItem
			return item, nil
		}
		usedItemIDsMu.Unlock()
		c.HTML(200, "main.html", gin.H{
			"message": "Вы победили!",
		})
	}
}

func NextRoundLeft(c *gin.Context, item1 string, item2 string) (structurs.Item, structurs.Item, error) {
	var newItem1 structurs.Item
	var newItem2 structurs.Item
	ItemID1, err := primitive.ObjectIDFromHex(item1)
	ItemID2, err := primitive.ObjectIDFromHex(item2)
	log.Println(item1)
	if err != nil {
		log.Println("Ошибка при преобразовании ObjectID из hex:", err)
	}

	filter := bson.M{"_id": ItemID1}
	err = gameCollection.FindOne(c, filter).Decode(&newItem1)
	if err != nil {
		log.Println(err)
		if err == mongo.ErrNoDocuments {
			c.AbortWithStatusJSON(404, gin.H{"error": "Item1 not found"})
		} else {
			c.AbortWithError(500, fmt.Errorf("failed to find item1: %w", err))
		}
		return structurs.Item{}, structurs.Item{}, fmt.Errorf("invalid item1 ID: %w", err)
	}

	filter = bson.M{"_id": ItemID2}
	err = gameCollection.FindOne(c, filter).Decode(&newItem2)
	if err != nil {
		log.Println(err)
		if err == mongo.ErrNoDocuments {
			c.AbortWithStatusJSON(404, gin.H{"error": "Item2 not found"})
		} else {
			c.AbortWithError(500, fmt.Errorf("failed to find item2: %w", err))
		}
		return structurs.Item{}, structurs.Item{}, fmt.Errorf("invalid item2 ID: %w", err)
	}

	if newItem1.Popularity > newItem2.Popularity {
		replacementItem, err := Replacement(c)
		if err != nil {
			log.Println("Ошибка при получении Replacement:", err)
			c.AbortWithError(500, fmt.Errorf("failed to get replacement item: %w", err))
			return structurs.Item{}, structurs.Item{}, err
		}
		return newItem1, replacementItem, nil
	} else {
		return structurs.Item{}, structurs.Item{}, fmt.Errorf("Вы проиграли!")
	}
}

func NextRoundRight(c *gin.Context, item1 string, item2 string) (structurs.Item, structurs.Item, error) {
	var newItem1 structurs.Item
	var newItem2 structurs.Item
	ItemID1, err := primitive.ObjectIDFromHex(item1)
	ItemID2, err := primitive.ObjectIDFromHex(item2)
	log.Println(item1)
	if err != nil {
		log.Println("Ошибка при преобразовании ObjectID из hex:", err)
	}

	filter := bson.M{"_id": ItemID1}
	err = gameCollection.FindOne(c, filter).Decode(&newItem1)
	if err != nil {
		log.Println(err)
		if err == mongo.ErrNoDocuments {
			c.AbortWithStatusJSON(404, gin.H{"error": "Item1 not found"})
		} else {
			c.AbortWithError(500, fmt.Errorf("failed to find item1: %w", err))
		}
		return structurs.Item{}, structurs.Item{}, fmt.Errorf("invalid item1 ID: %w", err)
	}

	filter = bson.M{"_id": ItemID2}
	err = gameCollection.FindOne(c, filter).Decode(&newItem2)
	if err != nil {
		log.Println(err)
		if err == mongo.ErrNoDocuments {
			c.AbortWithStatusJSON(404, gin.H{"error": "Item2 not found"})
		} else {
			c.AbortWithError(500, fmt.Errorf("failed to find item2: %w", err))
		}
		return structurs.Item{}, structurs.Item{}, fmt.Errorf("invalid item2 ID: %w", err)
	}

	if newItem2.Popularity > newItem1.Popularity {
		replacementItem, err := Replacement(c)
		if err != nil {
			log.Println("Ошибка при получении Replacement:", err)
			c.AbortWithError(500, fmt.Errorf("failed to get replacement item: %w", err))
			return structurs.Item{}, structurs.Item{}, err
		}
		return replacementItem, newItem2, nil
	} else {
		return structurs.Item{}, structurs.Item{}, fmt.Errorf("Вы проиграли!")
	}
}
