package client

import (
	"encoding/json"

	"github.com/uberswe/tradeguildledger/server"
)

func getLatestAddonVersion(addonName string) error {
	res, err := getFromAPI(url + "/api/v2/addon/version")
	if err != nil {
		return err
	}
	var response server.APIResponse
	err = json.Unmarshal(res, &response)
	if err != nil {
		return err
	}
	if response.Message == getAddonVersion() {
		return nil
	}
	// TODO download new version
	return nil
}

func getAddonVersion() string {
	// TODO get latest version from addon files
	return "0.0.0"
}
