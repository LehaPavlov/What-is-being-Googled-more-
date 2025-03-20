package page

import (
	"log"
	"main/request"
	"main/structurs"
	"sync"

	"github.com/gin-gonic/gin"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

var usedItemIDs = make(map[primitive.ObjectID]bool)
var usedItemIDsMu sync.Mutex

func Game(c *gin.Context) {
	usedItemIDsMu.Lock()
	usedItemIDs = make(map[primitive.ObjectID]bool)
	usedItemIDsMu.Unlock()
	item1, err := request.RandomItem(c)
	if err != nil {
		log.Println(err)
		c.HTML(500, "main.html", gin.H{"error": "Ошибка при получении первого элемента"})
		return
	}
	item2, err := request.RandomItem(c)
	if err != nil {
		log.Println(err)
		c.HTML(500, "main.html", gin.H{"error": "Ошибка при получении второго элемента"})
		return
	}
	if item1.ID == item2.ID {
		item2, err = request.RandomItem(c)
		if err != nil {
			log.Println("Ошибка при генерации")
			c.HTML(500, "main.html", gin.H{"error": "Ошибка при повторной генерации второго элемента"})
			return
		}
	}
	data := map[string]interface{}{
		"Item1": item1,
		"Item2": item2,
	}

	c.HTML(200, "main.html", data)
}

func Result(c *gin.Context) {
	formType := c.PostForm("form_type")
	var item1 structurs.Item
	var item2 structurs.Item
	var err error

	Item1ID := c.PostForm("left")
	Item2ID := c.PostForm("right")

	if formType == "item1" {
		item1, item2, err = request.NextRoundLeft(c, Item1ID, Item2ID)
	} else if formType == "item2" {
		item1, item2, err = request.NextRoundRight(c, Item1ID, Item2ID)
	} else {
		c.HTML(400, "main.html", gin.H{"error": "Неверный form_type"})
		return
	}

	if err != nil {
		log.Println(err)
		c.HTML(500, "main.html", gin.H{"error": "Ошибка в игре: " + err.Error()})
		return
	}

	if item1.ID.IsZero() && item2.ID.IsZero() {
		c.HTML(200, "main.html", gin.H{"error": "Вы проиграли"})
		return
	}
	if item1.ID.IsZero() || item2.ID.IsZero() {
		c.HTML(200, "main.html", gin.H{"error": "Вы выиграли!"})
		return
	}
	log.Println("Айтем 1 ", item1, "Айтем 2: ", item2)
	data := gin.H{
		"Item1": &item1,
		"Item2": &item2,
	}
	c.HTML(200, "main.html", data)
}
