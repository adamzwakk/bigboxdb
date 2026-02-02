package handlers

import (
    "net/http"
	"strings"
	"strconv"
	"os"
	"errors"

    "github.com/gin-gonic/gin"
	"gorm.io/gorm"

	"github.com/adamzwakk/bigboxdb-server/db"
	"github.com/adamzwakk/bigboxdb-server/models"
)

func AuthMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
		// Pretty much only for testing, dont actually use this
		insecure_enabled, _ := strconv.ParseBool(os.Getenv("BBDB_INSECURE_ADMIN"))
		if insecure_enabled {
			return
		}

		authHeader := c.GetHeader("Authorization")
        if authHeader == "" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": "Authorization header is required",
            })
            return
        }

		parts := strings.Split(authHeader, " ")
        if len(parts) != 2 || parts[0] != "Bearer" {
            c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
                "error": "Invalid authorization format",
            })
            return
        }

		database := db.GetDB()
		var user models.User
		result := database.Select("id").Where("api_key = ?", parts[1]).First(&user)
		if errors.Is(result.Error, gorm.ErrRecordNotFound) {
			c.AbortWithStatusJSON(http.StatusUnauthorized, gin.H{
				"error": "Invalid API key",
			})
			return
		}

        c.Next()
    }
}