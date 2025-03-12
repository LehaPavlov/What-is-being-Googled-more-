package page

import (
	"log"
	"main/request"

	"github.com/gin-gonic/gin"
)

func Game(c *gin.Context) {
	item1, err := request.RandomItem(c)
	if err != nil {
		log.Println(err)
	}
	item2, err := request.RandomItem(c)
	if err != nil {
		log.Println(err)
	}
	if item1.ID == item2.ID {
		item2, err = request.RandomItem(c)
		if err != nil {
			log.Println("Ошибка при генерации")
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
	if formType == "item1" {
		Item1ID := c.PostForm("left")
		Item2ID := c.PostForm("right")
		item1, item2, err := request.NextRoundLeft(c, Item1ID, Item2ID)
		log.Println(item1)
		if err != nil {
			if err.Error() == "Вы проиграли" {
				c.HTML(200, "main.html", gin.H{
					"error": "You lost!",
				})
				return
			}
		}
		if c.Writer.Status() != 200 {
			return
		}
		c.HTML(200, "main.html", gin.H{
			"Item1": &item1,
			"Item2": &item2,
		})
	}
	if formType == "item2" {
		Item1ID := c.PostForm("left")
		Item2ID := c.PostForm("right")
		item1, item2, err := request.NextRoundRight(c, Item1ID, Item2ID)
		log.Println(item1)
		log.Println(item2)
		log.Println(err)
		if err != nil {
			if err.Error() == "Вы проиграли!" {
				c.HTML(200, "main.html", gin.H{
					"error": "You lost!",
				})
				return
			}
		}

		if c.Writer.Status() != 200 {
			return
		}
		c.HTML(200, "main.html", gin.H{
			"Item1": &item1,
			"Item2": &item2,
		})
	}
}
