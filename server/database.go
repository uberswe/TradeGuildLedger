package server

import (
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"log"
	"os"
	"time"
)

func initDB() {
	var err error

	tglLogger := logger.New(
		log.New(os.Stdout, "\r\n", log.LstdFlags), // io writer
		logger.Config{
			SlowThreshold:             time.Second,
			LogLevel:                  logger.Warn,
			IgnoreRecordNotFoundError: true,
		},
	)

	db, err = gorm.Open(sqlite.Open("database.db"), &gorm.Config{
		Logger:                                   tglLogger,
		DisableForeignKeyConstraintWhenMigrating: true,
	})
	if err != nil {
		log.Fatal(err)
	}

	err = db.AutoMigrate(
		ItemModel{},
		ListingModel{},
		UpdateModel{},
		GuildModel{},
		NpcModel{},
		SellerModel{},
		VersionModel{})
	if err != nil {
		log.Fatal(err)
	}
}

type ItemModel struct {
	gorm.Model
	Timestamp      int
	Name           string
	Quality        int
	Texture        string
	Link           string
	VersionModelID uint
	ListingModels  []ListingModel
}

type ListingModel struct {
	gorm.Model
	CurrencyType   int
	ItemModelID    uint
	Price          int
	PricePerUnit   float64
	Quality        int
	StackCount     int
	SellerName     string
	TimeRemaining  int
	Timestamp      int
	ListingUID     float64
	VersionModelID uint
	SellerModelID  uint
	GuildModelID   uint
	NpcModelID     uint
}

type UpdateModel struct {
	gorm.Model
	Log string
	// In order to detect and prevent abuse
	IP string
}

type GuildModel struct {
	gorm.Model
	Name          string
	ListingModels []ListingModel
}

type NpcModel struct {
	gorm.Model
	Name          string
	ListingModels []ListingModel
}

type SellerModel struct {
	gorm.Model
	At            string
	ListingModels []ListingModel
}

type VersionModel struct {
	gorm.Model
	APIVersion    int
	Region        string
	AddonVersion  string
	ItemModels    []ItemModel
	ListingModels []ListingModel
}
