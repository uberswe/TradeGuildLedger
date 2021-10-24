package client

import (
	"bytes"
	"encoding/json"
	"fmt"
	"github.com/uberswe/tradeguildledger/pkg/parser"
	"github.com/uberswe/tradeguildledger/pkg/payloads"
	"io/ioutil"
	"log"
	"net/http"
	"time"
)

func syncWithRemote(d parser.ParsedData) {
	receiveUrl := url + "/api/v3/receive"
	if len(d.Listings) > 0 {
		b, err := json.Marshal(payloads.SendDataRequest{
			APIKey:       apiKey,
			AddonVersion: version,
			Items:        d,
		})
		if err != nil {
			log.Println(err)
			return
		}
		_, err = postToAPI(receiveUrl, b)
		if err != nil {
			addLog("Unable to sync with server")
			log.Println(err)
			return
		}
	}
	addLog(fmt.Sprintf("Processed %d listings at %s", len(d.Listings), time.Now().Format("2006-01-02 15:04:05")))
}

func postToAPI(url string, data []byte) ([]byte, error) {
	addLog(fmt.Sprintf("Sending data to %s", url))

	req, err := http.NewRequest("POST", url, bytes.NewBuffer(data))

	if err != nil {
		log.Println(err)
		return nil, err
	}
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

	if err != nil {
		log.Println(err)
		addLog("Error: could not fetch data")
		addLog("Please make sure you have the latest version of TradeGuildLedgerClient.exe")
		return nil, err
	}

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
