-- First, we create a namespace for our addon by declaring a top-level table that will hold everything else.
TradeGuildLedger = {}

-- This isn't strictly necessary, but we'll use this string later when registering events.
-- Better to define it in a single place rather than retyping the same string.
TradeGuildLedger.name = "TradeGuildLedger"

-- Next we create a function that will initialize our addon
function TradeGuildLedger:Initialize()
    self.savedVariables = ZO_SavedVars:NewAccountWide("TradeGuildLedgerVars", 2, nil, {})
end



-- Then we create an event handler function which will be called when the "addon loaded" event
-- occurs. We'll use this to initialize our addon after all of its resources are fully loaded.
function TradeGuildLedger.OnAddOnLoaded(event, addonName)
    -- The event fires each time *any* addon loads - but we only care about when our own addon loads.
    if addonName == TradeGuildLedger.name then
        TradeGuildLedger:Initialize()
    end
end

function TradeGuildLedger.OnTradingHouseClosed()
    d("Trading House Closed")
end

function TradeGuildLedger.OnTradingHouseResponseReceived(eventCode, responseType, result)
    d("Trading House Response Received " .. responseType)
    if (responseType == TRADING_HOUSE_RESULT_SEARCH_PENDING) then
        TradeGuildLedger.ProcessSearchResults()
    end
end

function TradeGuildLedger.ProcessSearchResults()
    local numItemsOnPage, currentPage, _ = GetTradingHouseSearchResultsInfo()
    d("numItemsOnPage: " .. numItemsOnPage)
    d("currentPage: " .. currentPage)
    local npc = GetRawUnitName("interact")
    d("npcName: " .. GetRawUnitName("interact"))
    if (TradeGuildLedger.savedVariables.npcs == nil) then
        TradeGuildLedger.savedVariables.npcs = {}
    end
    if (TradeGuildLedger.savedVariables.npcs[npc] == nil) then
        TradeGuildLedger.savedVariables.npcs[npc] = {}
        TradeGuildLedger.savedVariables.npcs[npc].items = {}
    end
    for i = 1, numItemsOnPage do
        local link = GetTradingHouseSearchResultItemLink(i)
        --textureName icon, string itemName, number quality, number stackCount, string sellerName, number timeRemaining, number purchasePrice, number CurrencyType currencyType, id64 itemUniqueId, number purchasePricePerUnit
        local textureName, itemName, quality, stackCount, sellerName, timeRemaining, purchasePrice, currencyType, uid, purchasePricePerUnit = GetTradingHouseSearchResultItemInfo(i)
        table.insert(TradeGuildLedger.savedVariables.npcs[npc].items, {link, textureName, itemName, quality, stackCount, sellerName, timeRemaining, purchasePrice, currencyType, uid, purchasePricePerUnit})
    end
end

function TradeGuildLedger.OnTradingHouseOpened()
    d("Trading House Opened")
end

function TradeGuildLedger.OnOldStoreHistoryRequested(eventCode, guildId, category)
    d("Trading House History Requested")
end

function TradeGuildLedger.OnTradingHouseConfirmItemPurchase(eventCode, pendingPurchaseIndex)
    d("Trading House Confirm Item Purchase")
end

-- Finally, we'll register our event handler function to be called when the proper event occurs.
EVENT_MANAGER:RegisterForEvent(TradeGuildLedger.name, EVENT_ADD_ON_LOADED, TradeGuildLedger.OnAddOnLoaded)
EVENT_MANAGER:RegisterForEvent(TradeGuildLedger.name, EVENT_CLOSE_TRADING_HOUSE, TradeGuildLedger.OnTradingHouseClosed)
EVENT_MANAGER:RegisterForEvent(TradeGuildLedger.name, EVENT_TRADING_HOUSE_RESPONSE_RECEIVED, TradeGuildLedger.OnTradingHouseResponseReceived)
EVENT_MANAGER:RegisterForEvent(TradeGuildLedger.name, EVENT_OPEN_TRADING_HOUSE, TradeGuildLedger.OnTradingHouseOpened)
EVENT_MANAGER:RegisterForEvent(TradeGuildLedger.name, EVENT_GUILD_HISTORY_RESPONSE_RECEIVED, TradeGuildLedger.OnOldStoreHistoryRequested)
EVENT_MANAGER:RegisterForEvent(TradeGuildLedger.name, EVENT_TRADING_HOUSE_CONFIRM_ITEM_PURCHASE, TradeGuildLedger.OnTradingHouseConfirmItemPurchase)