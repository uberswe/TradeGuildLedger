package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"gorm.io/gorm"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"
)

func parseIncomingPayload(p []byte, r *http.Request) {
	requestId := randomString(20)
	log.Printf("parsing %s\n", requestId)
	u := UpdateModel{
		Log: fmt.Sprintf("Incoming data at %s", time.Now().String()),
		IP:  r.RemoteAddr,
	}
	// TODO improve validation
	// Parse json
	var c Payload
	err := json.Unmarshal(p, &c)
	if err != nil {
		log.Println(err)
	}
	for _, v := range c.Content {
		for _, v2 := range v {
			vm := VersionModel{
				APIVersion:   v2.Version,
				Region:       v2.Region,
				AddonVersion: v2.Tglv,
			}
			if r := db.Where(&vm).First(&vm); r.Error != nil && !errors.Is(r.Error, gorm.ErrRecordNotFound) {
				log.Println(r.Error)
				return
			}

			if vm.ID == 0 {
				if r := db.Create(&vm); r.Error != nil {
					log.Println(r.Error)
					return
				}
			}

			for k, i := range v2.Items {
				parseItemPayload(k, i, vm)
			}
			for k, g := range v2.Guilds {
				parseListingsPayload(g, vm, k, "")
			}
			for k, npc := range v2.Npcs {
				parseListingsPayload(npc, vm, "", k)
			}
		}
	}
	if r := db.Create(&u); r.Error != nil {
		log.Println(r.Error)
		return
	}
	log.Printf("done %s\n", requestId)
}

func formatName(s string) string {
	if idx := strings.Index(s, "^"); idx != -1 {
		s = s[:idx]
	}
	return strings.Title(s)
}

func randomString(length int) string {
	rand.Seed(time.Now().UnixNano())
	b := make([]byte, length)
	rand.Read(b)
	return fmt.Sprintf("%x", b)[:length]
}

func parseItemPayload(l string, i Item, vm VersionModel) {
	var item ItemModel
	if r := db.First(&item, "link = ?", l); r.Error != nil && !errors.Is(r.Error, gorm.ErrRecordNotFound) {
		log.Println(r.Error)
		return
	}
	if item.ID == 0 {
		// If there is no item we create it
		is := ItemModel{
			Timestamp:      i.Ts,
			Name:           formatName(i.Itn),
			Quality:        i.Quality,
			Texture:        i.Tn,
			Link:           l,
			VersionModelID: vm.ID,
		}
		if r := db.Create(&is); r.Error != nil {
			log.Println(r.Error)
			return
		}
	}
	return
}

func parseListingsPayload(j json.RawMessage, vm VersionModel, guild string, npc string) {
	var c map[string][]Listing
	err := json.Unmarshal(j, &c)
	if err != nil {
		// Usually empty arrays
		return
	}
	var g GuildModel
	var nm NpcModel
	if guild != "" {
		if r := db.First(&g, "name = ?", formatName(guild)); r.Error != nil && !errors.Is(r.Error, gorm.ErrRecordNotFound) {
			log.Println(r.Error)
			return
		}
		if g.ID == 0 {
			g.Name = formatName(guild)
			if r := db.Create(&g); r.Error != nil {
				log.Println(r.Error)
				return
			}
		}
	}
	if npc != "" {
		if r := db.First(&nm, "name = ?", formatName(npc)); r.Error != nil && !errors.Is(r.Error, gorm.ErrRecordNotFound) {
			log.Println(r.Error)
			return
		}
		if nm.ID == 0 {
			nm.Name = formatName(npc)
			if r := db.Create(&nm); r.Error != nil {
				log.Println(r.Error)
				return
			}
		}
	}
	for _, v := range c {
		for _, v2 := range v {
			var listing ListingModel
			if r := db.First(&listing, "listing_uid = ?", v2.UID); r.Error != nil && !errors.Is(r.Error, gorm.ErrRecordNotFound) {
				log.Println(r.Error)
				return
			}
			var item ItemModel
			if r := db.First(&item, "link = ?", v2.L); r.Error != nil && !errors.Is(r.Error, gorm.ErrRecordNotFound) {
				log.Println(r.Error)
				return
			}
			if item.ID > 0 && listing.ID == 0 {
				var sm SellerModel
				if r := db.First(&sm, "at = ?", v2.Sn); r.Error != nil && !errors.Is(r.Error, gorm.ErrRecordNotFound) {
					log.Println(r.Error)
					return
				}
				if sm.ID == 0 {
					sm.At = v2.Sn
					if r := db.Create(&sm); r.Error != nil {
						log.Println(r.Error)
						return
					}
				}
				// If there is no item we create it
				ls := ListingModel{
					CurrencyType:   v2.Ct,
					ItemModelID:    item.ID,
					Price:          v2.Pp,
					PricePerUnit:   v2.Pppu,
					Quality:        v2.Quality,
					StackCount:     v2.Sc,
					SellerName:     v2.Sn,
					TimeRemaining:  v2.Tr,
					Timestamp:      v2.Ts,
					ListingUID:     v2.UID,
					VersionModelID: vm.ID,
					SellerModelID:  sm.ID,
					NpcModelID:     nm.ID,
					GuildModelID:   g.ID,
				}
				if r := db.Create(&ls); r.Error != nil {
					log.Println(r.Error)
					return
				}
			}
		}
	}
}

type Item struct {
	Itn     string `json:"itn"`
	Quality int    `json:"quality"`
	Tn      string `json:"tn"`
	Ts      int    `json:"ts"`
}

type Listing struct {
	// ts = timestamp, l = link, quality = quality, sc = stackCount, sn = sellerName, tr = timeRemaining, pp = price, ct = currencyType, uid = uid, pppu = purchasePricePerUnit
	Ct      int     `json:"ct"`
	L       string  `json:"l"`
	Pp      int     `json:"pp"`
	Pppu    float64 `json:"pppu"`
	Quality int     `json:"quality"`
	Sc      int     `json:"sc"`
	Sn      string  `json:"sn"`
	Tr      int     `json:"tr"`
	Ts      int     `json:"ts"`
	UID     float64 `json:"uid"`
}

type Body struct {
	Guilds  map[string]json.RawMessage `json:"guilds,omitempty"`
	Items   map[string]Item            `json:"items,omitempty"`
	Npcs    map[string]json.RawMessage `json:"npcs,omitempty"`
	Region  string                     `json:"region,omitempty"`
	Tglv    string                     `json:"tglv,omitempty"`
	Version int                        `json:"version,omitempty"`
}

type Payload struct {
	Content map[string]map[string]Body `json:"Default"`
}
