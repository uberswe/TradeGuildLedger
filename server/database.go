package server

import (
	"log"
	"os"
	"time"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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
		VersionModel{},
		RegionModel{})
	if err != nil {
		log.Fatal(err)
	}
}

type ItemModel struct {
	gorm.Model
	Timestamp      int
	Name           string
	Slug           string
	Quality        int
	Texture        string
	VersionModelID uint
	UID            int  // item id in the game
	Active         bool // New items from untrusted sources should not be shown automatically
	ListingModels  []ListingModel
}

type ListingModel struct {
	gorm.Model
	CurrencyType   int
	ItemModelID    uint
	ItemModel      ItemModel
	Price          int
	PricePerUnit   float64
	Quality        int
	StackCount     int
	TimeRemaining  int
	Timestamp      int
	ListingUID     float64
	VersionModelID uint
	SellerModelID  uint
	SellerModel    SellerModel
	GuildModelID   uint
	GuildModel     GuildModel
	NpcModelID     uint
	NpcModel       NpcModel
	Link           string
	RegionModelID  uint
	RegionModel    RegionModel
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
	Slug          string
	ListingModels []ListingModel
	Active        bool
}

type SellerModel struct {
	gorm.Model
	At            string
	RegionModel   RegionModel
	RegionModelID uint
	ListingModels []ListingModel
}

type RegionModel struct {
	gorm.Model
	Index        int
	Name         string
	SellerModels []SellerModel
}

type VersionModel struct {
	gorm.Model
	APIVersion    string
	Region        string
	AddonVersion  string
	ItemModels    []ItemModel
	ListingModels []ListingModel
}
