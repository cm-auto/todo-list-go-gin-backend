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

func getEntriesOfList(c *gin.Context, listId string) []models.Entry {
	collection := getEntryCollection(c)
	entries := collection.Find(func(entry *models.Entry) bool {
		return entry.ListID == listId
	})
	return entries
}

func getEntryCollection(c *gin.Context) *filedb.Collection[models.Entry] {
	dbAny, _ := c.Get("db")
	db := dbAny.(*filedb.Database)
	collection := db.GetEntryCollection()
	return collection
}

func getAllEntries(c *gin.Context) {
	collection := getEntryCollection(c)
	entries := collection.GetAll()
	c.JSON(http.StatusOK, entries)
}

func getEntryByID(c *gin.Context) {
	id := c.Param("id")
	entryCollection := getEntryCollection(c)
	entryOption := entryCollection.FindOne(func(entry *models.Entry) bool {
		return entry.ID == id
	})

	if entryOption == nil {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"message": "Entry not found",
		})
		return
	}
	c.JSON(http.StatusOK, entryOption)
}

type postEntryRequestData struct {
	ListID string `json:"listId" binding:"required"`
	Name   string `json:"name" binding:"required"`
	Done   bool   `json:"done"`
}

func postEntry(c *gin.Context) {
	data, _ := middlewares.GetGeneric[postEntryRequestData](c)

	if !doesListWithIdExist(c, data.ListID) {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"message": "List not found",
		})
		return
	}

	entryCollection := getEntryCollection(c)
	entry := models.Entry{
		ID:     uuid.NewString(),
		ListID: data.ListID,
		Name:   data.Name,
		Done:   data.Done,
	}
	if err := entryCollection.Append(entry); err != nil {
		fmt.Println(err)
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
		})
		return
	}

	c.JSON(http.StatusCreated, entry)
}

type patchEntryByIDRequestData struct {
	ListIDOption *string `json:"listId,omitempty"`
	NameOption   *string `json:"name,omitempty"`
	DoneOption   *bool   `json:"done,omitempty"`
}

func patchEntryByID(c *gin.Context) {
	data, _ := middlewares.GetGeneric[patchEntryByIDRequestData](c)
	id := c.Param("id")

	entryCollection := getEntryCollection(c)
	doesListExist := true
	entryOption, err := entryCollection.PatchOne(func(entry *models.Entry) bool {
		return entry.ID == id
	}, func(entry *models.Entry) {
		if data.ListIDOption != nil {
			if !doesListWithIdExist(c, *data.ListIDOption) {
				doesListExist = false
				return
			}
			entry.ListID = *data.ListIDOption
		}
		if data.NameOption != nil {
			entry.Name = *data.NameOption
		}
		if data.DoneOption != nil {
			entry.Done = *data.DoneOption
		}
	})
	if !doesListExist {
		c.AbortWithStatusJSON(http.StatusNotFound, gin.H{
			"message": "List not found",
		})
		return
	}
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
		})
		return
	}

	if entryOption == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Entry not found",
		})
		return
	}
	c.JSON(http.StatusOK, entryOption)
}

func deleteEntryByID(c *gin.Context) {
	id := c.Param("id")

	entryCollection := getEntryCollection(c)
	entryOption, err := entryCollection.DeleteOne(func(entry *models.Entry) bool {
		return entry.ID == id
	})
	if err != nil {
		c.AbortWithStatusJSON(http.StatusInternalServerError, gin.H{
			"error": "Internal server error",
		})
		return
	}
	if entryOption == nil {
		c.JSON(http.StatusNotFound, gin.H{
			"message": "Entry not found",
		})
		return
	}

	c.Status(http.StatusNoContent)
}

func RegisterEntryRoutes(group *gin.RouterGroup) {
	group.GET("", getAllEntries)
	group.GET("/:id", getEntryByID)
	group.POST("", middlewares.GenerateJsonValidatorMiddleware[postEntryRequestData](), postEntry)
	group.PATCH("/:id", middlewares.GenerateJsonValidatorMiddleware[patchEntryByIDRequestData](), patchEntryByID)
	group.DELETE("/:id", deleteEntryByID)
}
