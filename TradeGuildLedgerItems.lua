TradeGuildLedgerItems = {
    {{ range $val := . }}
    ["{{ $val.ID }}"] = {
        ["name"] = "{{ $val.Name }}",
        ["24"] = "{{ $val.TwentyFourHourAverage }}",
        ["7"] = "{{ $val.SevenDayAverage }}",
        ["30"] = "{{ $val.ThirtyDayAverage }}",
        ["low"] = "{{ $val.LowBuy }}",
        ["high"] = "{{ $val.HighBuy }}"
    }
    {{ end }}
}