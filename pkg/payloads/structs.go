package payloads

type Item struct {
	UID int `json:"uid"`
}

type Listing struct {
	ListingUID float64 `json:"listig_uid"`
}

type FullItem struct {
	ItemName    string `json:"item_name"`
	ID          int    `json:"id"`
	Quality     int    `json:"quality"`
	TextureName string `json:"texture_name"`
	Timestamp   int    `json:"timestamp"`
}

type FullListing struct {
	CurrencyType  int     `json:"currency_type"`
	Price         int     `json:"price"`
	ItemID        int     `json:"item_id"`
	Link          string  `json:"link"`
	PricePerUnit  float64 `json:"price_per_unit"`
	Quality       int     `json:"quality"`
	StackCount    int     `json:"stack_count"`
	SellerName    string  `json:"seller_name"`
	TimeRemaining int     `json:"time_remaining"`
	SeenTimestamp int     `json:"seen_timestamp"`
	UID           float64 `json:"uid"`
	NpcName       string  `json:"npc_name,omitempty"`
	RegionIndex   int     `json:"region_index"`
	RegionName    string  `json:"region_name"`
	GuildName     string  `json:"guild_name,omitempty"`
}

type SendItemsRequest struct {
	APIVersion   string     `json:"api_version"`
	APIKey       string     `json:"api_key"`
	Region       string     `json:"region"`
	AddonVersion string     `json:"addon_version"`
	Items        []FullItem `json:"items"`
}

type SendListingsRequest struct {
	APIVersion   string        `json:"api_version"`
	APIKey       string        `json:"api_key"`
	Region       string        `json:"region"`
	AddonVersion string        `json:"addon_version"`
	Listings     []FullListing `json:"listings"`
}
