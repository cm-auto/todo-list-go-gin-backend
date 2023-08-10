package middlewares

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
)

func GetGeneric[T any](c *gin.Context) (T, bool) {
	data, ok := c.Get("genericBodyData")
	casted, ok := data.(T)
	return casted, ok
}

func GenerateJsonValidatorMiddleware[T any]() gin.HandlerFunc {
	return func(c *gin.Context) {

		data := *new(T)

		contentType := c.Request.Header.Get("Content-Type")
		if contentType != "application/json" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "Unsupported media type"})
			c.Abort()
			return
		}

		if err := c.BindJSON(&data); err != nil {
			fmt.Println(err)
			validationErrs, ok := err.(validator.ValidationErrors)
			if !ok {
				c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid JSON body"})
				c.Abort()
				return
			}

			var errors []string
			for _, validationErr := range validationErrs {
				fieldName := validationErr.Field()

				jsonName := fieldName
				field, ok := reflect.TypeOf(data).FieldByName(fieldName)
				if ok {
					value := field.Tag.Get("json")
					if value != "" {
						jsonName = value
					}
				}

				errorMessage := jsonName + " " + validationErr.Tag()
				errors = append(errors, errorMessage)
			}
			c.JSON(http.StatusBadRequest, gin.H{"errors": errors})
			c.Abort()
			return
		}

		c.Set("genericBodyData", data)

		c.Next()
	}
}
