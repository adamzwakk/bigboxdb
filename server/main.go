package main

import (
	"log"
	"os"
	"net/http"
	"slices"

	"github.com/gin-gonic/gin"
	"github.com/joho/godotenv"
	
	"github.com/adamzwakk/bigboxdb-server/db"
	"github.com/adamzwakk/bigboxdb-server/models"
	"github.com/adamzwakk/bigboxdb-server/handlers"
)

func main() {
	if os.Getenv("APP_ENV") != "production" {
		if err := godotenv.Load("./../.env"); err != nil {
			log.Println("No .env file found (ok for production)")
		}
	}

	args := os.Args[1:]

	db.InitRedis()

	if slices.Contains(args, "migrate") {
		// SEED/MIGRATE DB
		database := db.GetDB()
		if err := database.AutoMigrate(
			&models.Game{},
			&models.Variant{},
			&db.SeedMeta{},
		); err != nil {
			log.Fatal(err)
		}
		if err := db.RunAllSeeds(database); err != nil {
			log.Fatal(err)
		}
		log.Println("Any pending migrations run!")
		
	} else if slices.Contains(args, "import") {
		zpath := args[1]

		if err := handlers.ImportLocal(zpath); err != nil {
			log.Fatal(err.Error())
			return
		}
	} else if slices.Contains(args, "host") {
		// MAIN WEB SERVER
		r := gin.Default()
		
		{
			a := r.Group("/api")

			a.GET("/health", func(c *gin.Context) {
				c.JSON(http.StatusOK, gin.H{
					"status": "ok",
				})
			})

			g := a.Group("/games")
			g.GET("/all", handlers.GamesAll)

			dev := a.Group("/developers")
			dev.GET("/all", handlers.GamesAll)

			pub := a.Group("/publishers")
			pub.GET("/all", handlers.GamesAll)

			v := a.Group("/variants")
			v.GET("/all", handlers.VariantsAll)
			v.GET("/latest", handlers.VariantsLatest)

			ad := a.Group("/admin")
			ad.Use(handlers.AuthMiddleware())
			{
				ad.PUT("/import", handlers.AdminImport)
			}
		}

		r.NoRoute(handlers.ServeIndex)
		r.Static("/assets", "./web/assets")
		r.Static("/img", "./web/img")
		r.Static("/basis", "./web/basis")
		r.Static("/scans", "./uploads/scans")

		 r.Use(func(c *gin.Context) {
			c.Header("Content-Security-Policy", "worker-src 'self' blob:;")
			c.Next()
		})

		r.Run(":8080")
	}
}