package db

import (
	"fmt"
	"log"
	"os"
	"sync"
	"time"

	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

var (
	conn *gorm.DB
	once sync.Once
)

func Connect() *gorm.DB {
	once.Do(func() {
		dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s",
			os.Getenv("MYSQL_USER"),
			os.Getenv("MYSQL_PASS"),
			os.Getenv("MYSQL_HOST"),
			os.Getenv("MYSQL_PORT"),
			os.Getenv("MYSQL_DB"),
		)

		db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{})
		if err != nil {
			log.Fatalf("failed to connect to mysql: %v", err)
		}

		// Optional: configure connection pool
		sqlDB, err := db.DB()
		if err != nil {
			log.Fatalf("failed to get sql.DB: %v", err)
		}

		sqlDB.SetMaxOpenConns(25)
		sqlDB.SetMaxIdleConns(25)
		sqlDB.SetConnMaxLifetime(5 * time.Minute)

		conn = db
	})

	return conn
}

func GetDB() *gorm.DB {
	if conn == nil {
		return Connect()
	}
	return conn
}