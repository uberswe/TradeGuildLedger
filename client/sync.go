package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"time"

	"github.com/uberswe/tradeguildledger/pkg/parser"
	"github.com/uberswe/tradeguildledger/pkg/payloads"
)

func syncWithRemote(p Processor) {
	logData = append(logData, "Syncing with server...")
	list.Refresh()
	list.Select(len(logData) - 1)

	itemUrl := url + "/api/v2/items"

	itemsJson, err := getFromAPI(itemUrl)

	if err != nil {
		log.Println(err)
		return
	}

	var items []payloads.Item
	err = json.Unmarshal(itemsJson, &items)

	if err != nil {
		log.Println(err)
		return
	}

	payloadItems := []payloads.FullItem{}

	for _, v := range p.items {
		found := false
		for _, item := range items {
			// Check if the item/listing already exists on the server
			if item.UID == v.ID {
				// TODO we could add a check here for differences in names, localization, etc.
				found = true
				break
			}
		}
		if found {
			continue
		}
		payloadItems = append(payloadItems, payloads.FullItem{
			ID:          v.ID,
			ItemName:    v.Itn,
			Quality:     v.Quality,
			TextureName: v.Tn,
			Timestamp:   v.Ts,
		})
	}

	if len(payloadItems) > 0 {
		b, err := json.Marshal(payloads.SendItemsRequest{
			Items:        payloadItems,
			APIVersion:   p.apiV,
			AddonVersion: p.version,
			Region:       p.region,
		})
		if err != nil {
			log.Println(err)
			return
		}
		postToAPI(itemUrl, b)
	}

	listingUrl := url + "/api/v2/listings"

	listingsJson, err := getFromAPI(listingUrl)

	var listings []payloads.Listing
	err = json.Unmarshal(listingsJson, &listings)

	if err != nil {
		log.Println(err)
		return
	}

	payloadListings := []payloads.FullListing{}

	for _, v := range p.listings {
		found := false
		for _, listing := range listings {
			// Check if the item/listing already exists on the server
			if listing.ListingUID == v.UID {
				// TODO we could add a check here for differences in names, localization, etc.
				found = true
				break
			}
		}
		if found {
			continue
		}

		r := regionFromIndex(p.regions, v.Region)

		payloadListings = append(payloadListings, payloads.FullListing{
			UID:           v.UID,
			Price:         v.Pp,
			CurrencyType:  v.Ct,
			ItemID:        v.Ii,
			Link:          v.Link,
			PricePerUnit:  v.Pppu,
			Quality:       v.Quality,
			StackCount:    v.Sc,
			SellerName:    v.Sn,
			TimeRemaining: v.Tr,
			SeenTimestamp: v.Ts,
			NpcName:       v.NpcName,
			GuildName:     v.GuildName,
			RegionIndex:   r.Index,
			RegionName:    r.Name,
		})
	}

	if len(payloadListings) > 0 {
		b, err := json.Marshal(payloads.SendListingsRequest{
			Listings:     payloadListings,
			APIVersion:   p.apiV,
			AddonVersion: p.version,
			Region:       p.region,
		})
		if err != nil {
			log.Println(err)
			return
		}
		postToAPI(listingUrl, b)
	}

	addLog(fmt.Sprintf("Last processed at %s", time.Now().Format("2006-01-02 15:04:05")))
}

func postToAPI(url string, data []byte) ([]byte, error) {
	addLog(fmt.Sprintf("Sending data to %s", url))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))
	req.Header.Set("Content-Type", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("%s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	return body, nil
}

func getFromAPI(url string) ([]byte, error) {
	addLog(fmt.Sprintf("Fetching data from %s", url))

	req, err := http.NewRequest("GET", url, nil)
	req.Header.Set("Accept", "application/json")

	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("%s", resp.Status)
	}

	body, err := ioutil.ReadAll(resp.Body)

	if err != nil {
		return nil, err
	}

	return body, nil
}

func regionFromIndex(regions []parser.Region, index int) parser.Region {
	for _, v := range regions {
		if v.Index == index {
			return v
		}
	}
	return parser.Region{}
}
