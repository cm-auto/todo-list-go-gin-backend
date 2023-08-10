package main

import (
	"fmt"
	"os"
	"todo-list-backend/src/filedb"
	"todo-list-backend/src/middlewares"
	"todo-list-backend/src/routes"

	"github.com/joho/godotenv"

	"github.com/gin-gonic/gin"
)

func main() {
	engine := gin.Default()

	db, err := filedb.NewDatabase("data")
	if err != nil {
		panic(err)
	}
	engine.Use(func(c *gin.Context) {
		c.Set("db", db)

		c.Next()
	})

	if err = godotenv.Load(); err != nil {
		panic("Error loading .env file")
	}

	port, ok := os.LookupEnv("PORT")
	if !ok {
		port = "3000"
	}

	engine.RedirectTrailingSlash = true
	// engine.Use(middlewares.TrimSlashMiddleware)

	engine.Use(middlewares.AllowAllCorsMiddleware)

	// this can be usefull if we want to serve the frontend
	// files from this application as well
	// then we could easily change the path for the api
	// endpoints to something like "/api"
	apiPathPrefix := ""

	listsGroup := engine.Group(apiPathPrefix + "/lists")
	routes.RegisterListRoutes(listsGroup)
	entriesGroup := engine.Group(apiPathPrefix + "/entries")
	routes.RegisterEntryRoutes(entriesGroup)

	fmt.Printf("Listening on port %s...", port)
	address := fmt.Sprintf(":%s", port)
	engine.Run(address)
}
