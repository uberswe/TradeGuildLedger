package server

import (
	"encoding/json"
	"errors"
	"fmt"
	"github.com/gosimple/slug"
	"github.com/julienschmidt/httprouter"
	"github.com/uberswe/tradeguildledger/pkg/payloads"
	"gorm.io/gorm"
	"io/ioutil"
	"log"
	"math"
	"net/http"
	"regexp"
	"strconv"
	"strings"
	"time"
)

func receiveData(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
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
	var dataRequest payloads.SendDataRequest
	err = json.Unmarshal(p, &dataRequest)
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	log.Printf("received %d listings for %s\n", len(dataRequest.Items.Listings), requestId)

	vm := VersionModel{
		APIVersion:   dataRequest.Items.Version,
		Region:       dataRequest.Items.Server,
		AddonVersion: dataRequest.AddonVersion,
	}

	if r := db.Where(&vm).First(&vm); r.Error != nil && !errors.Is(r.Error, gorm.ErrRecordNotFound) {
		log.Println("Version", r.Error)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}

	if vm.ID == 0 {
		if r := db.Create(&vm); r.Error != nil {
			log.Println("Version", r.Error)
			return
		}
	}

	go func() {
		// TODO optimize this later
		err = db.Transaction(func(tx *gorm.DB) error {
			for _, v := range dataRequest.Items.Guilds {
				guildString := strings.Replace(v, ": ", "_ ", -1)
				parts := strings.Split(guildString, ":")
				// { "g", timestamp, guildId, guildName }
				timestamp, err := strconv.Atoi(parts[0])
				if err != nil {
					// handle error
					fmt.Println(err)
					break
				}
				guildUID, err := strconv.Atoi(parts[1])
				if err != nil {
					// handle error
					fmt.Println(err)
					break
				}
				guild := GuildModel{
					Name:      parts[2],
					UID:       guildUID,
					Timestamp: timestamp,
				}
				if r := db.First(&guild, "uid = ?", guild.UID); r.Error != nil && !errors.Is(r.Error, gorm.ErrRecordNotFound) {
					log.Println("Guild", r.Error)
					return r.Error
				}
				if guild.ID == 0 {
					if r := db.Create(&guild); r.Error != nil {
						log.Println("Guild", r.Error)
						return r.Error
					}
				}
			}

			for _, v := range dataRequest.Items.Traits {
				// { "t", timestamp, traitType, traitDescription }
				parts := strings.Split(v, ":")

				timestamp, err := strconv.Atoi(parts[0])
				if err != nil {
					// handle error
					fmt.Println(err)
					break
				}
				traitType, err := strconv.Atoi(parts[1])
				if err != nil {
					// handle error
					fmt.Println(err)
					break
				}
				trait := TraitModel{
					Description: parts[2],
					Type:        traitType,
					Timestamp:   timestamp,
				}
				if r := db.First(&trait, "type = ?", trait.Type); r.Error != nil && !errors.Is(r.Error, gorm.ErrRecordNotFound) {
					log.Println("Trait", r.Error)
					return r.Error
				}
				if trait.ID == 0 {
					if r := db.Create(&trait); r.Error != nil {
						log.Println("Trait", r.Error)
						return r.Error
					}
				}
			}

			for _, v := range dataRequest.Items.Regions {
				// { "r", timestamp, region, GetZoneNameByIndex(region) }
				parts := strings.Split(v, ":")

				timestamp, err := strconv.Atoi(parts[0])
				if err != nil {
					// handle error
					fmt.Println(err)
					break
				}
				regionUID, err := strconv.Atoi(parts[1])
				if err != nil {
					// handle error
					fmt.Println(err)
					break
				}
				region := RegionModel{
					Index:        regionUID,
					Name:         parts[2],
					Timestamp:    timestamp,
					SellerModels: nil,
				}
				if r := db.First(&region, "`index` = ?", region.Index); r.Error != nil && !errors.Is(r.Error, gorm.ErrRecordNotFound) {
					log.Println("Region", r.Error)
					return r.Error
				}
				if region.ID == 0 {
					if r := db.Create(&region); r.Error != nil {
						log.Println("Region", r.Error)
						return r.Error
					}
				}
			}

			for _, v := range dataRequest.Items.Items {
				itemString := strings.Replace(v, ": ", "_ ", -1)
				parts := strings.Split(itemString, ":")
				// { "i", timestamp, id, quality, textureName, itemName, traitType }
				timestamp, err := strconv.Atoi(parts[0])
				if err != nil {
					// handle error
					fmt.Println(err)
					break
				}
				itemID, err := strconv.Atoi(parts[1])
				if err != nil {
					// handle error
					fmt.Println(err)
					break
				}
				quality, err := strconv.Atoi(parts[2])
				if err != nil {
					// handle error
					fmt.Println(err)
					break
				}
				traitType, err := strconv.Atoi(parts[5])
				if err != nil {
					// handle error
					fmt.Println(err)
					break
				}
				itemName := formatName(strings.Replace(parts[4], "_ ", ": ", -1))
				i := ItemModel{
					Timestamp:      timestamp,
					Name:           itemName,
					Slug:           slug.Make(itemName),
					Quality:        quality,
					Texture:        parts[3],
					VersionModelID: vm.ID,
					TraitType:      traitType,
					UID:            itemID,
					Active:         true,
				}
				if r := db.First(&i, "uid = ?", i.UID); r.Error != nil && !errors.Is(r.Error, gorm.ErrRecordNotFound) {
					log.Println("Item", r.Error)
					return r.Error
				}
				if i.ID == 0 {
					var match ItemModel
					if r := db.First(&match, "slug = ?", i.Slug); r.Error != nil && !errors.Is(r.Error, gorm.ErrRecordNotFound) {
						log.Println("Item", r.Error)
						return r.Error
					}
					if match.ID > 0 {
						i.Slug = fmt.Sprintf("%s-%d", i.Slug, i.UID)
					}
					if r := db.Create(&i); r.Error != nil {
						log.Println("Item", r.Error)
						return r.Error
					}
				}
			}

			for _, v := range dataRequest.Items.Listings {
				// { "l", timestamp, id, quality, stackCount, sellerName, timeRemaining, purchasePrice, currencyType, uid, purchasePricePerUnit, guildId, npc, region, link, traitType, (timestamp + id + purchasePrice - (purchasePricePerUnit * 3)) }
				parts := strings.Split(v, ":")
				if len(parts) == 38 {
					re := regexp.MustCompile("[0-9]+")
					res := re.FindAllString(parts[37], 1)
					if len(res) != 1 {
						fmt.Println("No validation provided")
						break
					}
					match, err := strconv.ParseFloat(res[0], 64)
					if err != nil {
						fmt.Println("Match", err)
						break
					}
					timestamp, err := strconv.Atoi(parts[0])
					if err != nil {
						// handle error
						fmt.Println(err)
						break
					}
					price, err := strconv.Atoi(parts[6])
					if err != nil {
						// handle error
						fmt.Println(err)
						break
					}
					currencyType, err := strconv.Atoi(parts[7])
					if err != nil {
						// handle error
						fmt.Println(err)
						break
					}
					quality, err := strconv.Atoi(parts[2])
					if err != nil {
						// handle error
						fmt.Println(err)
						break
					}
					stackCount, err := strconv.Atoi(parts[3])
					if err != nil {
						// handle error
						fmt.Println(err)
						break
					}
					timeRemaining, err := strconv.Atoi(parts[5])
					if err != nil {
						// handle error
						fmt.Println(err)
						break
					}
					traitType, err := strconv.Atoi(parts[36])
					if err != nil {
						// handle error
						fmt.Println(err)
						break
					}
					listingID, err := strconv.Atoi(parts[8])
					if err != nil {
						fmt.Println(err)
						break
					}
					itemUID, err := strconv.Atoi(parts[1])
					if err != nil {
						fmt.Println(err)
						break
					}
					pricePerUnit, err := strconv.ParseFloat(parts[9], 64)
					if err != nil {
						fmt.Println(err)
						break
					}
					if math.Floor(match) == math.Floor(float64(timestamp)+float64(itemUID)+float64(price)-(pricePerUnit*3)) {
						sellerName := parts[4]
						guildID := parts[10]
						npcName := parts[11]
						regionID := parts[12]

						var g GuildModel
						var nm NpcModel
						var sm SellerModel
						var im ItemModel
						var rm RegionModel
						// Check if trader exists
						if npcName != "" {
							if r := db.First(&nm, "name = ?", formatName(npcName)); r.Error != nil && !errors.Is(r.Error, gorm.ErrRecordNotFound) {
								log.Println("Npc", r.Error)
								return r.Error
							}
							if nm.ID == 0 {
								nm.Name = formatName(npcName)
								nm.Slug = slug.Make(formatName(nm.Name))
								// Make an admin area where this cann be approved
								nm.Active = true
								if r := db.Create(&nm); r.Error != nil {
									log.Println(r.Error)
									return r.Error
								}
							}
						}
						// Check if guild exists
						if guildID != "" {
							if r := db.First(&g, "uid = ?", formatName(guildID)); r.Error != nil && !errors.Is(r.Error, gorm.ErrRecordNotFound) {
								log.Println("Guild", r.Error)
								return r.Error
							}
						}
						// Check if seller exists
						if r := db.First(&sm, "at = ?", formatName(sellerName)); r.Error != nil && !errors.Is(r.Error, gorm.ErrRecordNotFound) {
							log.Println("Seller", r.Error)
							return r.Error
						}
						if sm.ID == 0 {
							sm.At = formatName(sellerName)
							if r := db.Create(&sm); r.Error != nil {
								log.Println("Seller", r.Error)
								return r.Error
							}
						}
						// check if item exists
						if r := db.First(&im, "uid = ?", itemUID); r.Error != nil {
							//  1628034813:152039:2:1:@Okis_Dokis:2578348:340:1:2296308039763.4:340:638134:Goh^M:180:|H0:item:152039:3:1:0:0:0:0:0:0:0:0:0:0:0:0:0:0:0:0:0:0|h|h:0:1628186172
							log.Println("Item", r.Error, itemUID, v)
							return r.Error
						}

						// Check if region exists
						if r := db.First(&rm, "`index` = ?", regionID); r.Error != nil && !errors.Is(r.Error, gorm.ErrRecordNotFound) {
							log.Println("Region", r.Error)
							return r.Error
						}

						l := ListingModel{
							CurrencyType:   currencyType,
							ItemModelID:    im.ID,
							Price:          price,
							PricePerUnit:   pricePerUnit,
							Quality:        quality,
							StackCount:     stackCount,
							TimeRemaining:  timeRemaining,
							Timestamp:      timestamp,
							ListingUID:     int64(listingID),
							VersionModelID: vm.ID,
							SellerModelID:  sm.ID,
							GuildModelID:   g.ID,
							NpcModelID:     nm.ID,
							RegionModelID:  rm.ID,
							Link1:          parts[13],
							Link2:          parts[14],
							Link3:          parts[15],
							Link4:          parts[16],
							Link5:          parts[17],
							Link6:          parts[18],
							Link7:          parts[19],
							Link8:          parts[20],
							Link9:          parts[21],
							Link10:         parts[22],
							Link11:         parts[23],
							Link12:         parts[24],
							Link13:         parts[25],
							Link14:         parts[26],
							Link15:         parts[27],
							Link16:         parts[28],
							Link17:         parts[29],
							Link18:         parts[30],
							Link19:         parts[31],
							Link20:         parts[32],
							Link21:         parts[33],
							Link22:         parts[34],
							Link23:         parts[35],
							TraitType:      traitType,
							Match:          match,
						}
						if r := db.First(&l, "listing_uid = ?", l.ListingUID); r.Error != nil && !errors.Is(r.Error, gorm.ErrRecordNotFound) {
							log.Println("Listing", r.Error)
							return r.Error
						}
						if l.ID == 0 {
							if r := db.Create(&l); r.Error != nil {
								log.Println("Listing", r.Error)
								return r.Error
							}
						}
					} else {
						log.Println("Validation failed", math.Floor(match), math.Floor(float64(timestamp)+float64(itemUID)+float64(price)-(pricePerUnit*3)))
						break
					}
				} else {
					log.Println("Parts invalid count", parts, len(parts))
					break
				}
			}

			//for _ := range dataRequest.Items.Buys {
			//	// { "s", timestamp, id, quality, stackCount, sellerName, timeRemaining, purchasePrice, currencyType, uid, purchasePricePerUnit, guildName, guildId, link }
			//}

			if r := db.Create(&u); r.Error != nil {
				log.Println(r.Error)
				return r.Error
			}

			log.Printf("done %s\n", requestId)
			return nil
		})
		if err != nil {
			log.Println(err)
		}
	}()

	// Response
	w.Header().Set("Content-Type", "application/json")
	jData, err := json.Marshal(APIResponse{
		Message:   "received",
		RequestID: requestId,
	})
	if err != nil {
		log.Println(err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	log.Printf("response %s\n", requestId)
	w.WriteHeader(http.StatusOK)
	_, _ = w.Write(jData)
}
