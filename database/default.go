package database

import (
	"log"
	"os"
	"sync"
	"time"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var _db *gorm.DB
var mutex sync.Mutex

func Setup() error {
	newLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second, // Slow SQL threshold
			LogLevel:                  logger.Info, // Log level
			IgnoreRecordNotFoundError: true,        // Ignore ErrRecordNotFound error for logger
			ParameterizedQueries:      true,        // Don't include params in the SQL log
			Colorful:                  true,        // Disable color
		},
	)

	var err error
	_db, err = gorm.Open(postgres.Open(os.Getenv("DATABASE_URL")), &gorm.Config{
		Logger: newLogger,
	})
	if err != nil {
		return err
	}

	return nil
}

// host=infra.a.pinggy.online user=myuser password=test123 dbname=mydb port=21996 sslmode=disable TimeZone=UTC
func GetDB() *gorm.DB {
	mutex.Lock()
	return _db
}

func ReleaseDB() {
	mutex.Unlock()
}
