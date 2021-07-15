package server

import (
	"errors"
	"fmt"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/BenJetson/humantime"
	"github.com/gosimple/slug"
	"github.com/julienschmidt/httprouter"
	"gorm.io/gorm"
)

type ListingData struct {
	Listings   []ListingView
	Offset     int
	NextOffset int
	PrevOffset int
	Search     string
}

type ListingView struct {
	ItemName                   string
	ItemColor                  string
	ItemSlug                   string
	Price                      int
	PricePerUnit               float64
	Quality                    int
	StackCount                 int
	TimeRemaining              int
	Timestamp                  int
	SellerName                 string
	RegionName                 string
	TraderName                 string
	TraderSlug                 string
	TimeRemainingHumanReadable string
	SeenHumanReadable          string
}

func listings(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	lp := filepath.Join("web", "layout.html")
	ip := filepath.Join("web", "listings.html")

	queryValues := r.URL.Query()
	search := queryValues.Get("search")

	limit := 100

	offsetCount := 0
	offset := p.ByName("offset")
	if offset != "" {
		i, err := strconv.Atoi(offset)
		if err != nil {
			// handle error
			fmt.Println(err)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		offsetCount = i
	}

	// TODO add pagination
	var listings []ListingModel
	if res := db.Model(&[]ListingModel{}).
		Preload("ItemModel").
		Preload("NpcModel").
		Preload("SellerModel").
		Preload("RegionModel").
		Joins("left join item_models on listing_models.item_model_id = item_models.id").
		Joins("left join npc_models on listing_models.npc_model_id = npc_models.id").
		Where("item_models.name LIKE ?", fmt.Sprintf("%%%s%%", search)).
		Where("item_models.active = 1").
		Where("npc_models.active = 1").
		Order("listing_models.id desc").
		Offset(offsetCount * limit).
		Limit(limit).
		Find(&listings); res.Error != nil && !errors.Is(res.Error, gorm.ErrRecordNotFound) {
		log.Println(res.Error)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles(lp, ip)
	if err != nil {
		log.Println(err)
		return
	}

	listingViews := []ListingView{}
	for _, listing := range listings {
		tm := time.Unix(int64(listing.Timestamp), 0)
		shr := humantime.Since(tm)
		if listing.ItemModel.Slug == "" {
			listing.ItemModel.Slug = slug.Make(listing.ItemModel.Name)
			db.Save(&listing.ItemModel)
			if r := db.Save(&listing.ItemModel); r.Error != nil {
				log.Println(r.Error)
			}
		}
		if listing.NpcModel.Slug == "" {
			listing.NpcModel.Slug = slug.Make(listing.NpcModel.Name)
			if r := db.Save(&listing.NpcModel); r.Error != nil {
				log.Println(r.Error)
			}
		}
		listingViews = append(listingViews, ListingView{
			ItemName:                   listing.ItemModel.Name,
			ItemColor:                  ItemColor(listing.Quality),
			ItemSlug:                   listing.ItemModel.Slug,
			Price:                      listing.Price,
			PricePerUnit:               listing.PricePerUnit,
			Quality:                    listing.Quality,
			StackCount:                 listing.StackCount,
			TimeRemaining:              listing.TimeRemaining,
			Timestamp:                  listing.Timestamp,
			TraderName:                 listing.NpcModel.Name,
			TraderSlug:                 listing.NpcModel.Slug,
			SellerName:                 listing.SellerModel.At,
			RegionName:                 listing.RegionModel.Name,
			TimeRemainingHumanReadable: "",
			SeenHumanReadable:          shr,
		})
	}

	err = tmpl.ExecuteTemplate(w, "layout", ListingData{
		Listings:   listingViews,
		Offset:     offsetCount,
		NextOffset: offsetCount + 1,
		PrevOffset: offsetCount - 1,
		Search:     search,
	})
	if err != nil {
		log.Println(err)
		return
	}
}

func ItemColor(q int) string {
	// Normal (white)
	// Fine (green)
	// Superior (blue)
	// Epic (purple)
	// Legendary (orange)
	switch q {
	case 0:
		return "#868e96"
	case 1:
		return "#868e96"
	case 2:
		return "#40c057"
	case 3:
		return "#228be6"
	case 4:
		return "#be4bdb"
	case 5:
		return "#fd7e14"
	}
	return "#000000"
}
