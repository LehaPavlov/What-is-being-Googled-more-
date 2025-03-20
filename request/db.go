package request

import (
	"context"
	"fmt"
	"log"
	"main/structurs"
	"math/rand"
	"net/http"
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
func Reset(c *gin.Context) {
	usedItemIDsMu.Lock()
	defer usedItemIDsMu.Unlock()
	usedItemIDs = make(map[primitive.ObjectID]bool)
	log.Println("Used items have been reset.")
	c.Redirect(http.StatusFound, "/")

}
func RandomItem(c *gin.Context) (structurs.Item, error) {
	count, err := gameCollection.CountDocuments(ctx, bson.M{})
	if err != nil {
		return structurs.Item{}, fmt.Errorf("failed to count documents: %w", err)
	}

	if count == 0 {
		return structurs.Item{}, fmt.Errorf("no items found in the database")
	}

	usedItemIDsMu.Lock()
	availableItemIDs := make([]primitive.ObjectID, 0, count-int64(len(usedItemIDs)))
	for i := int64(0); i < count; i++ { // Итерируемся с типом int64
		var item structurs.Item
		findOptions := options.FindOne().SetSkip(i)                            //Перебираем все элементы
		err = gameCollection.FindOne(ctx, bson.M{}, findOptions).Decode(&item) // Забираем каждый элемент
		if err != nil {
			usedItemIDsMu.Unlock()
			return structurs.Item{}, err
		}
		if _, used := usedItemIDs[item.ID]; !used {
			availableItemIDs = append(availableItemIDs, item.ID) // Добавляем в список, если еще не использовали
		}
	}
	// Если все элементы использованы, возвращаем ошибку
	if len(availableItemIDs) == 0 {
		usedItemIDsMu.Unlock()
		return structurs.Item{}, fmt.Errorf("no more items available")
	}
	// Выбираем случайный ID из доступных
	randomIndex := rand.Intn(len(availableItemIDs))
	selectedItemID := availableItemIDs[randomIndex]

	// Получаем элемент по выбранному ID
	var selectedItem structurs.Item
	err = gameCollection.FindOne(ctx, bson.M{"_id": selectedItemID}).Decode(&selectedItem)
	if err != nil {
		usedItemIDsMu.Unlock()
		return structurs.Item{}, fmt.Errorf("failed to find selected item: %w", err)
	}
	usedItemIDs[selectedItemID] = true // Помечаем как использованный
	usedItemIDsMu.Unlock()
	return selectedItem, nil
}

func ResetUsedItems() {
	usedItemIDsMu.Lock()
	usedItemIDs = make(map[primitive.ObjectID]bool)
	usedItemIDsMu.Unlock()
}
func Replacement(c *gin.Context) (structurs.Item, error) {
	count, err := gameCollection.EstimatedDocumentCount(c, nil)
	if err != nil {
		return structurs.Item{}, fmt.Errorf("failed to get document count: %w", err)
	}

	if count == 0 {
		return structurs.Item{}, fmt.Errorf("collection is empty")
	}
	usedItemIDsMu.Lock()
	defer usedItemIDsMu.Unlock()
	maxAttempts := 100
	for attempts := 0; attempts < maxAttempts; attempts++ {
		randomIndex := rand.Intn(int(count))
		findOptions := options.FindOne().SetSkip(int64(randomIndex))
		var item structurs.Item
		err = gameCollection.FindOne(c, bson.M{}, findOptions).Decode(&item)
		if _, used := usedItemIDs[item.ID]; !used {
			usedItemIDs[item.ID] = true
			return item, nil
		}
	}
	return structurs.Item{}, fmt.Errorf("Все данные были использованы из бд")
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

	var replacementItem structurs.Item
	if (left && newItem1.Popularity > newItem2.Popularity) || (!left && newItem2.Popularity > newItem1.Popularity) {
		for {
			replacementItem, err = Replacement(c)
			if replacementItem.ID != newItem1.ID && replacementItem.ID != newItem2.ID {
				if left {
					return newItem1, replacementItem, nil
				} else {
					return newItem2, replacementItem, nil
				}

			} else {
				return structurs.Item{}, structurs.Item{}, nil
			}
		}
	}
	return structurs.Item{}, structurs.Item{}, nil
}
