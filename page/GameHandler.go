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
