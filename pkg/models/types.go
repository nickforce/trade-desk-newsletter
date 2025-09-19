package models

type RotationsPayload struct {
	LatestDate    string   `json:"latestDate"`
	StockTickers  []string `json:"stockTickers"`
	DividendYield string   `json:"dividendYield"`
}

type State struct {
	HoldingsByDay map[string][]string `json:"holdings_by_day"`
	FirstSeen     map[string]string   `json:"first_seen"` // ticker -> first seen date (YYYY-MM-DD)
	LastSeen      map[string]string   `json:"last_seen"`  // ticker -> last seen date
}

type Tenured struct {
	Ticker string
	Days   int
}
