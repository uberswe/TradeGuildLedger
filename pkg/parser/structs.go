package parser

import "encoding/json"

type Item struct {
	Itn     string `json:"itn"`
	ID      int    `json:"id"`
	Quality int    `json:"quality"`
	Tn      string `json:"tn"`
	Ts      int    `json:"ts"`
}

type Region struct {
	Name  string `json:"name"`
	Index int    `json:"index"`
}

type Listing struct {
	// ts = timestamp, item = itemid, quality = quality, sc = stackCount, sn = sellerName, tr = timeRemaining, pp = price, ct = currencyType, uid = uid, pppu = purchasePricePerUnit
	Ct int `json:"ct"`
	Pp int `json:"pp"`
	Ii int `json:"item"`
	//     |H0/1:item:Id:SubType:InternalLevel:EnchantId:EnchantSubType:EnchantLevel:Writ1/TransmuteTrait:Writ2:Writ3:Writ4:Writ5:Writ6:0:0:0:Style:Crafted:Bound:Stolen:Charges:PotionEffect/WritReward|hName|h
	Link      string  `json:"link"`
	Pppu      float64 `json:"pppu"`
	Quality   int     `json:"quality"`
	Sc        int     `json:"sc"`
	Sn        string  `json:"sn"`
	Tr        int     `json:"tr"`
	Ts        int     `json:"ts"`
	UID       float64 `json:"uid"`
	NpcName   string  `json:"npc_name,omitempty"`
	GuildName string  `json:"guild_name,omitempty"`
	Region    int     `json:"region,omitempty"`
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
