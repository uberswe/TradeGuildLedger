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
		RegionModel{},
		TraitModel{})
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
	TraitType      int
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
	TraitType      int
	Match          float64
	Timestamp      int
	ListingUID     int64
	VersionModelID uint
	SellerModelID  uint
	SellerModel    SellerModel
	GuildModelID   uint
	GuildModel     GuildModel
	NpcModelID     uint
	NpcModel       NpcModel
	Link1          string
	Link2          string
	Link3          string
	Link4          string
	Link5          string
	Link6          string
	Link7          string
	Link8          string
	Link9          string
	Link10         string
	Link11         string
	Link12         string
	Link13         string
	Link14         string
	Link15         string
	Link16         string
	Link17         string
	Link18         string
	Link19         string
	Link20         string
	Link21         string
	Link22         string
	Link23         string
	RegionModelID  uint
	RegionModel    RegionModel
}

type BuyModel struct {
	gorm.Model
	CurrencyType   int
	ItemModelID    uint
	ItemModel      ItemModel
	Price          int
	PricePerUnit   float64
	Quality        int
	StackCount     int
	TimeRemaining  int
	TraitType      int
	Match          float64
	Timestamp      int
	ListingUID     int64
	VersionModelID uint
	SellerModelID  uint
	SellerModel    SellerModel
	GuildModelID   uint
	GuildModel     GuildModel
	NpcModelID     uint
	NpcModel       NpcModel
	Link1          string
	Link2          string
	Link3          string
	Link4          string
	Link5          string
	Link6          string
	Link7          string
	Link8          string
	Link9          string
	Link10         string
	Link11         string
	Link12         string
	Link13         string
	Link14         string
	Link15         string
	Link16         string
	Link17         string
	Link18         string
	Link19         string
	Link20         string
	Link21         string
	Link22         string
	Link23         string
}

type UpdateModel struct {
	gorm.Model
	Log string
	// In order to detect and prevent abuse
	IP string
}

type SeenModel struct {
	gorm.Model
	Name        string
	Description string
	Type        int
	Timestamp   int
}

type TraitModel struct {
	gorm.Model
	Name        string
	Description string
	Type        int
	Timestamp   int
}

type GuildModel struct {
	gorm.Model
	Name          string
	UID           int
	Timestamp     int
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
	Timestamp    int
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
