package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/gosimple/slug"
	"github.com/julienschmidt/httprouter"
	"github.com/uberswe/tradeguildledger/pkg/payloads"
	"gorm.io/gorm"
)

func receiveItems(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	requestId := randomString(20)
	log.Printf("parsing %s\n", requestId)
	u := UpdateModel{
		Log: fmt.Sprintf("Item data received at %s", time.Now().Format(time.RFC1123)),
		IP:  r.RemoteAddr,
	}
	p, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// TODO improve validation
	// Parse json
	var itemRequest payloads.SendItemsRequest
	err = json.Unmarshal(p, &itemRequest)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Printf("received %d items for %s\n", len(itemRequest.Items), requestId)

	vm := VersionModel{
		APIVersion:   itemRequest.APIVersion,
		Region:       itemRequest.Region,
		AddonVersion: itemRequest.AddonVersion,
	}
	if r := db.Where(&vm).First(&vm); r.Error != nil && !errors.Is(r.Error, gorm.ErrRecordNotFound) {
		log.Println(r.Error)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if vm.ID == 0 {
		if r := db.Create(&vm); r.Error != nil {
			log.Println(r.Error)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	created := 0

	for _, item := range itemRequest.Items {
		var i ItemModel
		if r := db.First(&i, "uid = ?", item.ID); r.Error != nil && !errors.Is(r.Error, gorm.ErrRecordNotFound) {
			log.Println(r.Error)
			return
		}
		if i.ID == 0 {
			// If there is no item we create it
			is := ItemModel{
				Timestamp:      item.Timestamp,
				Name:           formatName(item.ItemName),
				Slug:           slug.Make(formatName(item.ItemName)),
				Quality:        item.Quality,
				Texture:        item.TextureName,
				VersionModelID: vm.ID,
				UID:            item.ID,
				// Make an admin area where this cann be approved
				Active: true,
			}
			if r := db.Create(&is); r.Error != nil {
				log.Println(r.Error)
				return
			}
			created++
		}
	}
	log.Printf("created %d items for %s\n", created, requestId)
	log.Printf("done %s\n", requestId)

	if r := db.Create(&u); r.Error != nil {
		log.Println(r.Error)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Response
	w.Header().Set("Content-Type", "application/json")
	jData, err := json.Marshal(APIResponse{
		Message: "received",
	})
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jData)
}

func fetchItems(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// In the future we need to distinguish between regions and maybe languages
	var items []payloads.Item
	if r := db.Model(&[]ItemModel{}).Find(&items); r.Error != nil && !errors.Is(r.Error, gorm.ErrRecordNotFound) {
		log.Println(r.Error)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Response
	w.Header().Set("Content-Type", "application/json")
	jData, err := json.Marshal(items)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jData)
}

func receiveListings(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	requestId := randomString(20)
	log.Printf("parsing %s\n", requestId)
	u := UpdateModel{
		Log: fmt.Sprintf("Listing data received at %s", time.Now().Format(time.RFC1123)),
		IP:  r.RemoteAddr,
	}
	p, err := ioutil.ReadAll(r.Body)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// TODO improve validation
	// Parse json
	var listingRequest payloads.SendListingsRequest
	err = json.Unmarshal(p, &listingRequest)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	vm := VersionModel{
		APIVersion:   listingRequest.APIVersion,
		Region:       listingRequest.Region,
		AddonVersion: listingRequest.AddonVersion,
	}
	if r := db.Where(&vm).First(&vm); r.Error != nil && !errors.Is(r.Error, gorm.ErrRecordNotFound) {
		log.Println(r.Error)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if vm.ID == 0 {
		if r := db.Create(&vm); r.Error != nil {
			log.Println(r.Error)
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
	}

	go func() {
		for _, listing := range listingRequest.Listings {
			var l ListingModel
			if r := db.First(&l, "listing_uid = ?", listing.UID); r.Error != nil && !errors.Is(r.Error, gorm.ErrRecordNotFound) {
				log.Println(r.Error)
				return
			}
			if l.ID == 0 {
				var g GuildModel
				var nm NpcModel
				var sm SellerModel
				var im ItemModel
				var rm RegionModel
				// Check if trader exists
				if listing.NpcName != "" {
					if r := db.First(&nm, "name = ?", formatName(listing.NpcName)); r.Error != nil && !errors.Is(r.Error, gorm.ErrRecordNotFound) {
						log.Println(r.Error)
						return
					}
					if nm.ID == 0 {
						nm.Name = formatName(listing.NpcName)
						nm.Slug = slug.Make(formatName(nm.Name))
						// Make an admin area where this cann be approved
						nm.Active = true
						if r := db.Create(&nm); r.Error != nil {
							log.Println(r.Error)
							return
						}
					}
				}
				// Check if guild exists
				if listing.GuildName != "" {
					if r := db.First(&g, "name = ?", formatName(listing.GuildName)); r.Error != nil && !errors.Is(r.Error, gorm.ErrRecordNotFound) {
						log.Println(r.Error)
						return
					}
					if g.ID == 0 {
						g.Name = formatName(listing.GuildName)
						if r := db.Create(&g); r.Error != nil {
							log.Println(r.Error)
							return
						}
					}
				}
				// Check if seller exists
				if r := db.First(&sm, "at = ?", formatName(listing.SellerName)); r.Error != nil && !errors.Is(r.Error, gorm.ErrRecordNotFound) {
					log.Println(r.Error)
					return
				}
				if sm.ID == 0 {
					sm.At = formatName(listing.SellerName)
					if r := db.Create(&sm); r.Error != nil {
						log.Println(r.Error)
						return
					}
				}
				// check if item exists
				if r := db.First(&im, "uid = ?", listing.ItemID); r.Error != nil {
					log.Println(r.Error)
					return
				}

				// Check if region exists
				if r := db.First(&rm, "`index` = ?", listing.RegionIndex); r.Error != nil && !errors.Is(r.Error, gorm.ErrRecordNotFound) {
					log.Println(r.Error)
					return
				}

				if rm.ID == 0 {
					rm.Index = listing.RegionIndex
					rm.Name = listing.RegionName
					if listing.RegionName == "" {
						log.Println(errors.New("region name is empty"))
						return
					}
					if r := db.Create(&rm); r.Error != nil {
						log.Println(r.Error)
						return
					}
				}

				// If there is no listing we create it
				ls := ListingModel{
					Timestamp:      listing.SeenTimestamp,
					CurrencyType:   listing.CurrencyType,
					ItemModelID:    im.ID,
					Price:          listing.Price,
					PricePerUnit:   listing.PricePerUnit,
					Quality:        listing.Quality,
					StackCount:     listing.StackCount,
					TimeRemaining:  listing.TimeRemaining,
					ListingUID:     listing.UID,
					SellerModelID:  sm.ID,
					GuildModelID:   g.ID,
					NpcModelID:     nm.ID,
					Link:           listing.Link,
					VersionModelID: vm.ID,
					RegionModelID:  rm.ID,
				}
				if r := db.Create(&ls); r.Error != nil {
					log.Println(r.Error)
					return
				}
			}
		}
		log.Printf("done %s\n", requestId)
	}()

	if r := db.Create(&u); r.Error != nil {
		log.Println(r.Error)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Response
	w.Header().Set("Content-Type", "application/json")
	jData, err := json.Marshal(APIResponse{
		Message: "received",
	})
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jData)
}

func fetchListings(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	// TODO In the future we need to distinguish between regions and maybe languages
	var listings []payloads.Listing
	if r := db.Model(&[]ListingModel{}).Find(&listings); r.Error != nil && !errors.Is(r.Error, gorm.ErrRecordNotFound) {
		log.Println(r.Error)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	// Response
	w.Header().Set("Content-Type", "application/json")
	jData, err := json.Marshal(listings)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jData)
}

func fetchAddonVersion(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
	jData, err := json.Marshal(APIResponse{
		Message: addonVersion,
	})
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jData)
}
