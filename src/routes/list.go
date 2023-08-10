package routes

import (
	"fmt"
	"net/http"
	"todo-list-backend/src/filedb"
	"todo-list-backend/src/middlewares"
	"todo-list-backend/src/models"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
)

func getListCollection(c *gin.Context) *filedb.Collection[models.List] {
	dbAny, _ := c.Get("db")
	db := dbAny.(*filedb.Database)
	collection := db.GetListCollection()
	return collection
}

func doesListWithIdExist(c *gin.Context, id string) bool {
	listCollection := getListCollection(c)
	listOption := listCollection.FindOne(func(list *models.List) bool {
		return list.ID == id
	})
	return listOption != nil
}

func getAllLists(c *gin.Context) {
	collection := getListCollection(c)
	lists := collection.GetAll()
	c.JSON(200, lists)
}

func getListByID(c *gin.Context) {
	id := c.Param("id")
	listCollection := getListCollection(c)
	listOption := listCollection.FindOne(func(list *models.List) bool {
		return list.ID == id
	})

	if listOption == nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"message": "List not found",
		})
		return
	}
	c.JSON(200, listOption)
}

func getListByIDAndItsEntries(c *gin.Context) {
	id := c.Param("id")
	listCollection := getListCollection(c)
	listOption := listCollection.FindOne(func(list *models.List) bool {
		return list.ID == id
	})

	if listOption == nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"message": "List not found",
		})
		return
	}
	entries := getEntriesOfList(c, id)
	body := models.NewParentAndChildren(listOption, entries)
	c.JSON(200, body)
}

type postListRequestData struct {
	Name string `json:"name" binding:"required"`
}

func postList(c *gin.Context) {
	data, _ := middlewares.GetGeneric[postListRequestData](c)

	listCollection := getListCollection(c)
	list := models.List{
		ID:   uuid.NewString(),
		Name: data.Name,
	}
	if err := listCollection.Append(list); err != nil {
		fmt.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
		})
		return
	}

	c.JSON(http.StatusCreated, list)
}

type patchListByIDRequestData struct {
	NameOption *string `json:"name" binding:"omitempty"`
}

func patchListByID(c *gin.Context) {
	data, _ := middlewares.GetGeneric[patchListByIDRequestData](c)
	id := c.Param("id")

	listCollection := getListCollection(c)
	listOption, err := listCollection.PatchOne(func(list *models.List) bool {
		return list.ID == id
	}, func(list *models.List) {
		if data.NameOption != nil {
			list.Name = *data.NameOption
		}
	})
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
		})
		return
	}

	if listOption == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "List not found",
		})
		return
	}
	c.JSON(http.StatusOK, listOption)
}

func deleteListByID(c *gin.Context) {
	id := c.Param("id")

	entryCollection := getEntryCollection(c)
	_, err := entryCollection.DeleteMany(func(entry *models.Entry) bool {
		return entry.ListID == id
	})
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
		})
		return
	}

	listCollection := getListCollection(c)
	listOption, err := listCollection.DeleteOne(func(list *models.List) bool {
		return list.ID == id
	})

	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
		})
		return
	}
	if listOption == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "List not found",
		})
		return
	}

	c.Status(http.StatusNoContent)
}

func RegisterListRoutes(group *gin.RouterGroup) {
	group.GET("", getAllLists)
	group.GET("/:id", getListByID)
	group.GET("/:id/entries", getListByIDAndItsEntries)
	group.POST("", middlewares.GenerateJsonValidatorMiddleware[postListRequestData](), postList)
	group.PATCH("/:id", middlewares.GenerateJsonValidatorMiddleware[patchListByIDRequestData](), patchListByID)
	group.DELETE("/:id", deleteListByID)
}
