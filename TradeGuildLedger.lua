TradeGuildLedger = {}

TradeGuildLedger.name = "TradeGuildLedger"

function TradeGuildLedger:Initialize()
    TradeGuildLedger.savedVariables = ZO_SavedVars:NewAccountWide("TradeGuildLedgerVars", 3, nil, {}, GetWorldName())

    TradeGuildLedger.savedVariables.world = table.concat({ "w", GetWorldName() }, ":")

    if (TradeGuildLedger.savedVariables.regions == nil) then
        TradeGuildLedger.savedVariables.regions = {}
    end
    if (TradeGuildLedger.savedVariables.items == nil) then
        TradeGuildLedger.savedVariables.items = {}
    end
    if (TradeGuildLedger.savedVariables.traits == nil) then
        TradeGuildLedger.savedVariables.traits = {}
    end
    if (TradeGuildLedger.savedVariables.listings == nil) then
        TradeGuildLedger.savedVariables.listings = {}
    end
    if (TradeGuildLedger.savedVariables.buys == nil) then
        TradeGuildLedger.savedVariables.buys = {}
    end
    if (TradeGuildLedger.savedVariables.guilds == nil) then
        TradeGuildLedger.savedVariables.guilds = {}
    end

    -- Migrations
    if (TradeGuildLedger.savedVariables.tglv ~= "{{ .Version }}") then
        -- Initial version, clear all previous data
        TradeGuildLedger.savedVariables.items = {}
        TradeGuildLedger.savedVariables.listings = {}
        TradeGuildLedger.savedVariables.buys = {}
        TradeGuildLedger.savedVariables.guilds = {}
        TradeGuildLedger.savedVariables.regions = {}
        TradeGuildLedger.savedVariables.traits = {}
        TradeGuildLedger.savedVariables.tglv = "{{ .Version }}"
    end

end

function TradeGuildLedger.OnPlayerActivated(eventCode)
    -- Get rid of listings older than 24 hours
    local timestamp = GetTimeStamp() - (60 * 60 * 24)
    for k, v in pairs(TradeGuildLedger.savedVariables.listings) do
        if (string.len(TradeGuildLedger.GetTsForListing(v)) == 0 or tonumber(TradeGuildLedger.GetTsForListing(v)) < timestamp) then
            table.remove(TradeGuildLedger.savedVariables.listings, k)
        end
    end
end

function TradeGuildLedger.OnAddOnLoaded(event, addonName)
    -- The event fires each time *any* addon loads - but we only care about when our own addon loads.
    if addonName == TradeGuildLedger.name then
        TradeGuildLedger:Initialize()

        -- Unregister Loaded Callback
        EVENT_MANAGER:UnregisterForEvent(TradeGuildLedger.name, EVENT_ADD_ON_LOADED)

        -- Register for player activated event
        EVENT_MANAGER:RegisterForEvent(TradeGuildLedger.name, EVENT_PLAYER_ACTIVATED, TradeGuildLedger.OnPlayerActivated)
    end
end

EVENT_MANAGER:RegisterForEvent(TradeGuildLedger.name, EVENT_ADD_ON_LOADED, TradeGuildLedger.OnAddOnLoaded)

function TradeGuildLedger.OnTradingHouseResponseReceived(eventCode, responseType, result)
    if (responseType == TRADING_HOUSE_RESULT_SEARCH_PENDING) then
        TradeGuildLedger.ProcessSearchResults()
    elseif (responseType == TRADING_HOUSE_RESULT_LISTINGS_PENDING or responseType == TRADING_HOUSE_RESULT_CANCEL_SALE_PENDING or responseType == TRADING_HOUSE_RESULT_POST_PENDING) then
        TradeGuildLedger.ProcessGuildListings()
    end
end

function TradeGuildLedger.ProcessSearchResults()
    local numItemsOnPage, _, _ = GetTradingHouseSearchResultsInfo()
    local npc = GetRawUnitName("interact")
    local region = GetUnitZoneIndex("interact")
    if region == nil then
        region = GetCurrentMapZoneIndex()
    end
    local guildId, _, _ = GetCurrentTradingHouseGuildDetails()
    local guildName = GetGuildName(guildId)
    local timestamp = GetTimeStamp()
    if (region ~= nil and TradeGuildLedger.savedVariables.regions[region] == nil) then
        TradeGuildLedger.savedVariables.regions[region] = table.concat({ "r", timestamp, region, GetZoneNameByIndex(region) }, ":")
    end
    for i = 1, numItemsOnPage do
        local link = GetTradingHouseSearchResultItemLink(i)
        local id = TradeGuildLedger.GetIdFromLink(link)
        local traitType, traitDescription = GetItemLinkTraitInfo(link);
        -- textureName icon, string itemName, number quality, number stackCount, string sellerName, number timeRemaining, number purchasePrice, number CurrencyType currencyType, id64 itemUniqueId, number purchasePricePerUnit
        local textureName, itemName, quality, stackCount, sellerName, timeRemaining, purchasePrice, currencyType, uid, purchasePricePerUnit = GetTradingHouseSearchResultItemInfo(i)
        table.insert(TradeGuildLedger.savedVariables.listings, table.concat({ "l", timestamp, id, quality, stackCount, sellerName, timeRemaining, purchasePrice, currencyType, Id64ToString(uid), purchasePricePerUnit, guildId, npc, region, link, traitType, (timestamp + id + purchasePrice - (purchasePricePerUnit * 3)) }, ":"))
        if (id ~= nil and TradeGuildLedger.savedVariables.items[id] == nil) then
            TradeGuildLedger.savedVariables.items[id] = table.concat({ "i", timestamp, id, quality, textureName, itemName, traitType }, ":")
        end
        if (traitType ~= nil and TradeGuildLedger.savedVariables.traits[traitType] == nil) then
            TradeGuildLedger.savedVariables.traits[traitType] = table.concat({ "t", timestamp, traitType, traitDescription }, ":")
        end
    end
    if (guildId ~= nil and guildName ~= "" and TradeGuildLedger.savedVariables.guilds[guildId] == nil) then
        TradeGuildLedger.savedVariables.guilds[guildId] = table.concat({ "g", timestamp, guildId, guildName }, ":")
    end
end

function TradeGuildLedger.GetIdFromLink(link)
    -- |H0/1:item:Id:SubType:InternalLevel:EnchantId:EnchantSubType:EnchantLevel:Writ1/TransmuteTrait:Writ2:Writ3:Writ4:Writ5:Writ6:0:0:0:Style:Crafted:Bound:Stolen:Charges:PotionEffect/WritReward|hName|h
    return string.match(link, "item:([0-9]+):")
end

function TradeGuildLedger.GetTsForListing(s)
    return string.match(s, "l:([0-9]+):")
end

function TradeGuildLedger.GetTsForItem(s)
    return string.match(s, "i:([0-9]+):")
end

function TradeGuildLedger.ProcessGuildListings()
    local guildId, guildName, _ = GetCurrentTradingHouseGuildDetails()
    if (guildId == 0 or guildName == "") then
        return
    end
    local numListing = GetNumTradingHouseListings()
    local timestamp = GetTimeStamp()
    for i = 1, numListing do
        local link = GetTradingHouseListingItemLink(i)
        local id = TradeGuildLedger.GetIdFromLink(link)
        local traitType, traitDescription = GetItemLinkTraitInfo(link);
        local textureName, itemName, quality, stackCount, sellerName, timeRemaining, price, currencyType, uid, purchasePricePerUnit = GetTradingHouseListingItemInfo(i)
        table.insert(TradeGuildLedger.savedVariables.listings, table.concat({ "l", timestamp, id, quality, stackCount, sellerName, timeRemaining, price, currencyType, Id64ToString(uid), purchasePricePerUnit, guildId, "", "", link, traitType, (timestamp + id + price - (purchasePricePerUnit * 3)) }, ":"))
        if (id ~= nil and TradeGuildLedger.savedVariables.items[id] == nil) then
            TradeGuildLedger.savedVariables.items[id] = table.concat({ "i", timestamp, id, quality, textureName, itemName, traitType }, ":")
        end
        if (traitType ~= nil and TradeGuildLedger.savedVariables.traits[traitType] == nil) then
            TradeGuildLedger.savedVariables.traits[traitType] = table.concat({ "t", timestamp, traitType, traitDescription }, ":")
        end
    end
    if (guildId ~= nil and TradeGuildLedger.savedVariables.guilds[guildId] == nil) then
        TradeGuildLedger.savedVariables.guilds[guildId] = table.concat({ "g", timestamp, guildId, guildName }, ":")
    end
end

function TradeGuildLedger.OnTradingHouseConfirmItemPurchase(eventCode, slotId)
    local textureName, itemName, quality, stackCount, sellerName, timeRemaining, purchasePrice, currencyType, uid, purchasePricePerUnit = GetTradingHouseSearchResultItemInfo(slotId)
    local guildId, guildName, _ = GetCurrentTradingHouseGuildDetails()
    local timestamp = GetTimeStamp()
    local link = GetTradingHouseSearchResultItemLink(slotId)
    local id = TradeGuildLedger.GetIdFromLink(link)
    local traitType, traitDescription = GetItemLinkTraitInfo(link);
    table.insert(TradeGuildLedger.savedVariables.buys, table.concat({ "s", timestamp, id, quality, stackCount, sellerName, timeRemaining, purchasePrice, currencyType, Id64ToString(uid), purchasePricePerUnit, guildName, guildId, link }, ":"))
    if (traitType ~= nil and TradeGuildLedger.savedVariables.traits[traitType] == nil) then
        TradeGuildLedger.savedVariables.traits[traitType] = table.concat({ "t", timestamp, traitType, traitDescription }, ":")
    end
    if (id ~= nil and TradeGuildLedger.savedVariables.items[id] == nil) then
        TradeGuildLedger.savedVariables.items[id] = table.concat({ "i", timestamp, id, quality, textureName, itemName, traitType}, ":")
    end
    if (guildId ~= nil and TradeGuildLedger.savedVariables.guilds[guildId] == nil) then
        TradeGuildLedger.savedVariables.guilds[guildId] = table.concat({ "g", timestamp, guildId, guildName }, ":")
    end
end

-- Register event handler functions
EVENT_MANAGER:RegisterForEvent(TradeGuildLedger.name, EVENT_TRADING_HOUSE_RESPONSE_RECEIVED, TradeGuildLedger.OnTradingHouseResponseReceived)
EVENT_MANAGER:RegisterForEvent(TradeGuildLedger.name, EVENT_TRADING_HOUSE_CONFIRM_ITEM_PURCHASE, TradeGuildLedger.OnTradingHouseConfirmItemPurchase)