TradeGuildLedger = {}

TradeGuildLedger.name = "TradeGuildLedger"

function TradeGuildLedger:Initialize()
    self.savedVariables = ZO_SavedVars:NewAccountWide("TradeGuildLedgerVars", 2, nil, {})
    if (TradeGuildLedger.savedVariables.items == nil) then
        TradeGuildLedger.savedVariables.items = {}
    end
    if (TradeGuildLedger.savedVariables.npcs == nil) then
        TradeGuildLedger.savedVariables.npcs = {}
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
    local numItemsOnPage, currentPage, _ = GetTradingHouseSearchResultsInfo()
    local npc = GetRawUnitName("interact")
    if (TradeGuildLedger.savedVariables.npcs[npc] == nil) then
        TradeGuildLedger.savedVariables.npcs[npc] = {}
        TradeGuildLedger.savedVariables.npcs[npc].items = {}
    end
    local timestamp = GetTimeStamp()
    for i = 1, numItemsOnPage do
        local link = GetTradingHouseSearchResultItemLink(i)
        -- textureName icon, string itemName, number quality, number stackCount, string sellerName, number timeRemaining, number purchasePrice, number CurrencyType currencyType, id64 itemUniqueId, number purchasePricePerUnit
        local textureName, itemName, quality, stackCount, sellerName, timeRemaining, purchasePrice, currencyType, uid, purchasePricePerUnit = GetTradingHouseSearchResultItemInfo(i)
        table.insert(TradeGuildLedger.savedVariables.npcs[npc].items, {ts=timestamp, l=link, quality=quality, sc=stackCount, sn=sellerName, tr=timeRemaining, pp=purchasePrice, ct=currencyType, uid=uid, pppu=purchasePricePerUnit})
        if (TradeGuildLedger.savedVariables.items[link] == nil) then
            TradeGuildLedger.savedVariables.items[link] = {ts=timestamp, tn=textureName, itn=itemName, quality=quality}
        end
    end
end

function TradeGuildLedger.ProcessGuildListings()
    local guildID, _, _ = GetCurrentTradingHouseGuildDetails()
    local guildName = GetGuildName(guildID)
    local numListing = GetNumTradingHouseListings()
    if (TradeGuildLedger.savedVariables.guilds == nil) then
        TradeGuildLedger.savedVariables.guilds = {}
    end
    if (TradeGuildLedger.savedVariables.guilds[guildName] == nil) then
        TradeGuildLedger.savedVariables.guilds[guildName] = {}
        TradeGuildLedger.savedVariables.guilds[guildName].items = {}
    end
    local timestamp = GetTimeStamp()
    for i = 1, numListing do
        local link = GetTradingHouseListingItemLink(i)
        local textureName, itemName, quality, stackCount, sellerName, timeRemaining, price, currencyType, uid, purchasePricePerUnit = GetTradingHouseListingItemInfo(i)
        table.insert(TradeGuildLedger.savedVariables.guilds[guildName].items, {ts=timestamp, l=link, quality=quality, sc=stackCount, sn=sellerName, tr=timeRemaining, pp=price, ct=currencyType, uid=uid, pppu=purchasePricePerUnit})
        if (TradeGuildLedger.savedVariables.items[link] == nil) then
            TradeGuildLedger.savedVariables.items[link] = {ts=timestamp, tn=textureName, itn=itemName, quality=quality}
        end
    end
end

function TradeGuildLedger.OnTradingHouseOpened()

end

function TradeGuildLedger.OnTradingHouseClosed()

end

function TradeGuildLedger.OnOldStoreHistoryRequested(eventCode, guildId, category)

end

function TradeGuildLedger.OnTradingHouseConfirmItemPurchase(eventCode, pendingPurchaseIndex)

end

-- Register event handler functions
EVENT_MANAGER:RegisterForEvent(TradeGuildLedger.name, EVENT_ADD_ON_LOADED, TradeGuildLedger.OnAddOnLoaded)
EVENT_MANAGER:RegisterForEvent(TradeGuildLedger.name, EVENT_CLOSE_TRADING_HOUSE, TradeGuildLedger.OnTradingHouseClosed)
EVENT_MANAGER:RegisterForEvent(TradeGuildLedger.name, EVENT_TRADING_HOUSE_RESPONSE_RECEIVED, TradeGuildLedger.OnTradingHouseResponseReceived)
EVENT_MANAGER:RegisterForEvent(TradeGuildLedger.name, EVENT_OPEN_TRADING_HOUSE, TradeGuildLedger.OnTradingHouseOpened)
EVENT_MANAGER:RegisterForEvent(TradeGuildLedger.name, EVENT_GUILD_HISTORY_RESPONSE_RECEIVED, TradeGuildLedger.OnOldStoreHistoryRequested)
EVENT_MANAGER:RegisterForEvent(TradeGuildLedger.name, EVENT_TRADING_HOUSE_CONFIRM_ITEM_PURCHASE, TradeGuildLedger.OnTradingHouseConfirmItemPurchase)