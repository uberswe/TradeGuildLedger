TradeGuildLedger = {}

TradeGuildLedger.name = "TradeGuildLedger"

function TradeGuildLedger:Initialize()
    self.savedVariables = ZO_SavedVars:NewAccountWide("TradeGuildLedgerVars", 2, nil, {})
    if (TradeGuildLedger.savedVariables.regions == nil) then
        TradeGuildLedger.savedVariables.regions = {}
    end
    if (TradeGuildLedger.savedVariables.items == nil) then
        TradeGuildLedger.savedVariables.items = {}
    end
    if (TradeGuildLedger.savedVariables.npcs == nil) then
        TradeGuildLedger.savedVariables.npcs = {}
    end
    if (TradeGuildLedger.savedVariables.guilds == nil) then
        TradeGuildLedger.savedVariables.guilds = {}
    end
    -- Migrations
    if (TradeGuildLedger.savedVariables.tglv ~= "0.0.1") then
        -- Initial version, clear all previous data
        TradeGuildLedger.savedVariables.items = {}
        TradeGuildLedger.savedVariables.npcs = {}
        TradeGuildLedger.savedVariables.guilds = {}
        TradeGuildLedger.savedVariables.tglv = "0.0.1"
    end
    TradeGuildLedger.savedVariables.region = TradeGuildLedger.GetRegion()
    local timestamp = GetTimeStamp()
    for k, v in pairs(TradeGuildLedger.savedVariables.npcs) do
        for k2, v2 in pairs(v) do
            if type(v2) == "table" then
                for k3, v3 in pairs(v2) do
                    -- remove entries older than 24 hours
                    if (v3.ts == nil or (v3.ts + 86400) < timestamp) then
                        table.remove(TradeGuildLedger.savedVariables.npcs[k][k2], k3)
                    end
                end
            end
        end
    end
    for k, v in pairs(TradeGuildLedger.savedVariables.guilds) do
        for k2, v2 in pairs(v) do
            for k3, v3 in pairs(v2) do
                -- remove entries older than 24 hours
                if (v3.ts == nil or (v3.ts + 86400) < timestamp) then
                    table.remove(TradeGuildLedger.savedVariables.guilds[k][k2], k3)
                end
            end
        end
    end
end

function TradeGuildLedger.OnAddOnLoaded(event, addonName)
    -- The event fires each time *any* addon loads - but we only care about when our own addon loads.
    if addonName == TradeGuildLedger.name then
        TradeGuildLedger:Initialize()
    end
end

function TradeGuildLedger.OnTradingHouseResponseReceived(eventCode, responseType, result)
    if (responseType == TRADING_HOUSE_RESULT_SEARCH_PENDING) then
        TradeGuildLedger.ProcessSearchResults()
    elseif (responseType == TRADING_HOUSE_RESULT_LISTINGS_PENDING or responseType == TRADING_HOUSE_RESULT_CANCEL_SALE_PENDING or responseType == TRADING_HOUSE_RESULT_POST_PENDING) then
        TradeGuildLedger.ProcessGuildListings()
    end
end

function TradeGuildLedger.ProcessSearchResults()
    local region = GetCurrentMapZoneIndex()
    local numItemsOnPage, _, _ = GetTradingHouseSearchResultsInfo()
    local npc = GetRawUnitName("interact")
    if (TradeGuildLedger.savedVariables.npcs[npc] == nil) then
        TradeGuildLedger.savedVariables.npcs[npc] = {}
    end
    if (TradeGuildLedger.savedVariables.npcs[npc].items == nil) then
        TradeGuildLedger.savedVariables.npcs[npc].items = {}
    end
    if (TradeGuildLedger.savedVariables.regions[region] == nil) then
        TradeGuildLedger.savedVariables.regions[region] = { name = GetZoneNameByIndex(region) }
    end
    local timestamp = GetTimeStamp()
    for i = 1, numItemsOnPage do
        local link = GetTradingHouseSearchResultItemLink(i)
        local id = TradeGuildLedger.GetIdFromLink(link)
        -- textureName icon, string itemName, number quality, number stackCount, string sellerName, number timeRemaining, number purchasePrice, number CurrencyType currencyType, id64 itemUniqueId, number purchasePricePerUnit
        local textureName, itemName, quality, stackCount, sellerName, timeRemaining, purchasePrice, currencyType, uid, purchasePricePerUnit = GetTradingHouseSearchResultItemInfo(i)
        table.insert(TradeGuildLedger.savedVariables.npcs[npc].items, { ts = timestamp, item = id, link=link, quality = quality, sc = stackCount, sn = sellerName, tr = timeRemaining, pp = purchasePrice, ct = currencyType, uid = uid, pppu = purchasePricePerUnit })
        TradeGuildLedger.savedVariables.npcs[npc].region = region
        if (TradeGuildLedger.savedVariables.items[id] == nil) then
            TradeGuildLedger.savedVariables.items[id] = { ts = timestamp, tn = textureName, itn = itemName, quality = quality }
        end
    end
end

function TradeGuildLedger.GetIdFromLink(link)
    -- |H0/1:item:Id:SubType:InternalLevel:EnchantId:EnchantSubType:EnchantLevel:Writ1/TransmuteTrait:Writ2:Writ3:Writ4:Writ5:Writ6:0:0:0:Style:Crafted:Bound:Stolen:Charges:PotionEffect/WritReward|hName|h
    return string.match(link, "item:([0-9]+):")
end

function TradeGuildLedger.ProcessGuildListings()
    local guildID, _, _ = GetCurrentTradingHouseGuildDetails()
    local guildName = GetGuildName(guildID)
    local numListing = GetNumTradingHouseListings()
    if (TradeGuildLedger.savedVariables.guilds[guildName] == nil) then
        TradeGuildLedger.savedVariables.guilds[guildName] = {}
    end
    if (TradeGuildLedger.savedVariables.guilds[guildName].items == nil) then
        TradeGuildLedger.savedVariables.guilds[guildName].items = {}
    end
    local timestamp = GetTimeStamp()
    for i = 1, numListing do
        local link = GetTradingHouseListingItemLink(i)
        local id = TradeGuildLedger.GetIdFromLink(link)
        local textureName, itemName, quality, stackCount, sellerName, timeRemaining, price, currencyType, uid, purchasePricePerUnit = GetTradingHouseListingItemInfo(i)
        table.insert(TradeGuildLedger.savedVariables.guilds[guildName].items, { ts = timestamp, item = id, link=link, quality = quality, sc = stackCount, sn = sellerName, tr = timeRemaining, pp = purchasePrice, ct = currencyType, uid = uid, pppu = purchasePricePerUnit })
        if (TradeGuildLedger.savedVariables.items[id] == nil) then
            TradeGuildLedger.savedVariables.items[id] = { ts = timestamp, tn = textureName, itn = itemName, quality = quality }
        end
    end
end

function TradeGuildLedger.OnTradingHouseConfirmItemPurchase(eventCode, pendingPurchaseIndex)
    -- TODO implement purchase tracking
end

function TradeGuildLedger.GetRegion()
    local lastPlatform = GetCVar("LastPlatform")
    local lastRealm = GetCVar("LastRealm")
    if (lastPlatform == "Live") then
        return "NA"
    elseif (lastPlatform == "Live-EU") then
        return "EU"
    elseif (lastRealm:find("^NA") ~= nil) then
        return "NA"
    end
    return "EU"
end

-- Register event handler functions
EVENT_MANAGER:RegisterForEvent(TradeGuildLedger.name, EVENT_ADD_ON_LOADED, TradeGuildLedger.OnAddOnLoaded)
EVENT_MANAGER:RegisterForEvent(TradeGuildLedger.name, EVENT_TRADING_HOUSE_RESPONSE_RECEIVED, TradeGuildLedger.OnTradingHouseResponseReceived)
EVENT_MANAGER:RegisterForEvent(TradeGuildLedger.name, EVENT_TRADING_HOUSE_CONFIRM_ITEM_PURCHASE, TradeGuildLedger.OnTradingHouseConfirmItemPurchase)