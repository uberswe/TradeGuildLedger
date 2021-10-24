package payloads

import "github.com/uberswe/tradeguildledger/pkg/parser"

type SendDataRequest struct {
	APIKey       string            `json:"1"`
	AddonVersion string            `json:"2"`
	Items        parser.ParsedData `json:"3"`
}
