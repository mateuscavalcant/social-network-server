package main

import
(
	"github.com/joho/godotenv"
	"log"
	"social-network-server/pkg/database"
	"github.com/gin-gonic/gin"
	"social-network-server/api/routes"

)



func main() {
	err := godotenv.Load()
	if err != nil {
		log.Fatal("Error loading .env file")
	}

	database.InitializeDB()

	r := gin.Default()
	r.Use(gzip.Gzip(gzip.DefaultCompression))

	routes.InitRoutes(r.Group("/"))
	err = http.ListenAndServe(":8080", r)
	if err != nil {
		log.Fatal("Failed to start server: ", err)
	}


}
