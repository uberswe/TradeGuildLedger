package server

import (
	"errors"
	"fmt"
	"github.com/BenJetson/humantime"
	"html/template"
	"log"
	"net/http"
	"path/filepath"
	"strconv"
	"time"

	"github.com/julienschmidt/httprouter"
	"gorm.io/gorm"
)

func traders(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	lp := filepath.Join("web", "layout.html")
	ip := filepath.Join("web", "traders.html")

	limit := 20

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

	var npcs []NpcModel
	if res := db.Offset(offsetCount * limit).
		Limit(limit).
		Order("id desc").
		Find(&npcs); res.Error != nil && !errors.Is(res.Error, gorm.ErrRecordNotFound) {
		log.Println(res.Error)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	tmpl, err := template.ParseFiles(lp, ip)
	if err != nil {
		log.Println(err)
		return
	}
	npcData := NpcData{
		Npcs: npcs,
	}
	npcData.Offset = offsetCount
	npcData.NextOffset = offsetCount + 1
	npcData.PrevOffset = offsetCount - 1
	npcData.URLPath = r.URL.Path
	npcData.DarkMode = findDarkmode
	npcData.FormatLink = linkFormatter
	npcData.Region = findRegion
	npcData.DarkModeLink = darkModeLinkFormatter
	npcData.Title = "NPCs"

	err = tmpl.ExecuteTemplate(w, "layout", listings)
	if err != nil {
		log.Println(err)
		return
	}
}

func trader(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	lp := filepath.Join("web", "layout.html")
	ip := filepath.Join("web", "trader.html")

	slug := p.ByName("slug")
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
		Where("npc_models.slug = ?", slug).
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

	name := ""
	slug = ""
	region := ""

	var listingViews []ListingView
	for _, listing := range listings {
		tm := time.Unix(int64(listing.Timestamp), 0)
		shr := humantime.Since(tm)

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

		if region == "" || name == "" || slug == "" {
			region = listing.RegionModel.Name
			name = listing.NpcModel.Name
			slug = listing.NpcModel.Slug
		}
	}

	traderData := TraderData{
		Listings:   listingViews,
		Search:     search,
		RegionName: region,
		TraderName: name,
		Slug:       slug,
	}
	traderData.Offset = offsetCount
	traderData.NextOffset = offsetCount + 1
	traderData.PrevOffset = offsetCount - 1
	traderData.URLPath = r.URL.Path
	traderData.DarkMode = findDarkmode
	traderData.FormatLink = linkFormatter
	traderData.Region = findRegion
	traderData.DarkModeLink = darkModeLinkFormatter
	traderData.Title = traderData.TraderName

	err = tmpl.ExecuteTemplate(w, "layout", traderData)
	if err != nil {
		log.Println(err)
		return
	}
}
