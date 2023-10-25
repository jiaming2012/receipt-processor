package database

import (
	"os"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"sync"
)

var _db *gorm.DB
var mutex sync.Mutex

func Setup() error {
	var err error
	_db, err = gorm.Open(postgres.Open(os.Getenv("DATABASE_URL")), &gorm.Config{})
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
