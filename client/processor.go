package client

import (
	"log"
	"strconv"

	"github.com/uberswe/tradeguildledger/pkg/parser"
)

type Processor struct {
	user     string
	region   string
	version  string
	apiV     string
	items    []parser.Item
	listings []parser.Listing
	regions  []parser.Region
}

func process(m map[string]interface{}) {
	p := recursiveFind(m, Processor{})
	syncWithRemote(p)
}

func recursiveFind(m map[string]interface{}, p Processor) Processor {
	for k, v := range m {
		if k == "items" {
			if m2, ok := v.(map[string]interface{}); ok {
				for k2, v2 := range m2 {
					if m3, ok2 := v2.(map[string]interface{}); ok2 {
						itemID, err := strconv.Atoi(k2)
						if err != nil {
							break
						}
						item := parser.Item{}
						for k3, v3 := range m3 {
							if v3, ok3 := v3.(string); ok3 {
								if k3 == "itn" {
									item.Itn = v3
								} else if k3 == "quality" {
									i, err := strconv.Atoi(v3)
									if err != nil {
										log.Println(err)
										break
									}
									item.Quality = i
								} else if k3 == "tn" {
									item.Tn = v3
								} else if k3 == "ts" {
									i, err := strconv.Atoi(v3)
									if err != nil {
										log.Println(err)
										break
									}
									item.Ts = i
								}
							}
						}
						item.ID = itemID
						p.items = append(p.items, item)
						log.Println(len(p.items))
					}
				}
			}
		} else if k == "regions" {
			if m2, ok := v.(map[string]interface{}); ok {
				for k2, v2 := range m2 {
					if s, ok2 := v2.(map[string]interface{}); ok2 {
						region := parser.Region{}
						regionIndex, err := strconv.Atoi(k2)
						if err != nil {
							break
						}
						region.Index = regionIndex
						region.Name = s["name"].(string)
						p.regions = append(p.regions, region)
					}
				}
			}
		} else if k == "npcs" || k == "guilds" {
			if m2, ok := v.(map[string]interface{}); ok {
				for k2, v2 := range m2 {
					if m3, ok2 := v2.(map[string]interface{}); ok2 {
						region := 0
						if r, ok := m3["region"].(string); ok {
							regionIndex, err := strconv.Atoi(r)
							if err != nil {
								log.Println(err)
								break
							}
							region = regionIndex
						}
						if l, ok := m3["items"].(map[string]interface{}); ok {
							for _, v4 := range l {
								if m5, ok4 := v4.(map[string]interface{}); ok4 {
									listing := parser.Listing{}
									for k5, v5 := range m5 {
										if v5, ok5 := v5.(string); ok5 {
											if k5 == "item" {
												i, err := strconv.Atoi(v5)
												if err != nil {
													log.Println(err)
													break
												}
												listing.Ii = i
											} else if k5 == "quality" {
												i, err := strconv.Atoi(v5)
												if err != nil {
													log.Println(err)
													break
												}
												listing.Quality = i
											} else if k5 == "link" {
												listing.Link = v5
											} else if k5 == "ct" {
												i, err := strconv.Atoi(v5)
												if err != nil {
													log.Println(err)
													break
												}
												listing.Ct = i
											} else if k5 == "pp" {
												i, err := strconv.Atoi(v5)
												if err != nil {
													log.Println(err)
													break
												}
												listing.Pp = i
											} else if k5 == "sc" {
												i, err := strconv.Atoi(v5)
												if err != nil {
													log.Println(err)
													break
												}
												listing.Sc = i
											} else if k5 == "sn" {
												listing.Sn = v5
											} else if k5 == "tr" {
												i, err := strconv.Atoi(v5)
												if err != nil {
													log.Println(err)
													break
												}
												listing.Tr = i
											} else if k5 == "ts" {
												i, err := strconv.Atoi(v5)
												if err != nil {
													log.Println(err)
													break
												}
												listing.Ts = i
											} else if k5 == "pppu" {
												s, err := strconv.ParseFloat(v5, 64)
												if err != nil {
													log.Println(err)
													break
												}
												listing.Pppu = s
											} else if k5 == "uid" {
												s, err := strconv.ParseFloat(v5, 64)
												if err != nil {
													log.Println(err)
													break
												}
												listing.UID = s
											}
										}
									}
									if k == "npcs" {
										listing.NpcName = k2
										listing.Region = region
									} else {
										listing.GuildName = k2
									}
									p.listings = append(p.listings, listing)
								}
							}
						}
					}
				}
			}
		} else if k == "region" {
			if r, ok := v.(string); ok {
				p.region = r
			}
		} else if k == "tglv" {
			if r, ok := v.(string); ok {
				p.version = r
			}
		} else if k == "version" {
			if r, ok := v.(string); ok {
				p.apiV = r
			}
		} else if k == "Default" {
			if m2, ok := v.(map[string]interface{}); ok {
				for k2, v2 := range m2 {
					p.user = k2
					if m3, ok := v2.(map[string]interface{}); ok {
						p = recursiveFind(m3, p)
					}
				}
			}
		} else {
			if m2, ok := v.(map[string]interface{}); ok {
				p = recursiveFind(m2, p)
			}
		}
	}
	return p
}
