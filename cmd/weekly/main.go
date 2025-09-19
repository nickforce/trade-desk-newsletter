package main

import (
	"fmt"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"trade-desk-newsletter/pkg/models"
	"trade-desk-newsletter/pkg/render"
	"trade-desk-newsletter/pkg/state"

	"github.com/joho/godotenv"
)

func init() {
	if err := godotenv.Load(); err != nil {
		log.Println("No .env file found, relying on environment variables")
	}
}

func main() {
	st := state.New("data/state.json")
	s, err := st.Load()
	must(err)

	// Find latest date in holdings
	dates := make([]string, 0, len(s.HoldingsByDay))
	for d := range s.HoldingsByDay {
		dates = append(dates, d)
	}
	sort.Strings(dates)
	if len(dates) == 0 {
		fmt.Println("No data in state.json yet. Exiting.")
		return
	}
	latest := dates[len(dates)-1]

	// Week window: last 7 days from latest
	end, _ := time.Parse("2006-01-02", latest)
	start := end.AddDate(0, 0, -6)
	weekStart := start.Format("2006-01-02")
	weekEnd := latest

	// Radar: tickers added or removed during this window
	radar := []map[string]string{}
	for i := 1; i < len(dates); i++ {
		prev := s.HoldingsByDay[dates[i-1]]
		today := s.HoldingsByDay[dates[i]]
		added, removed := diff(prev, today)
		for _, a := range added {
			radar = append(radar, map[string]string{"Ticker": a, "Note": "surfaced this week"})
		}
		for _, r := range removed {
			radar = append(radar, map[string]string{"Ticker": r, "Note": "rotated out this week"})
		}
	}

	// Convictions: names present >= 15 days
	convictions := []models.Tenured{}
	for ticker, firstSeen := range s.FirstSeen {
		lastSeen := s.LastSeen[ticker]
		days := tenorDays(firstSeen, lastSeen)
		if days >= 15 {
			convictions = append(convictions, models.Tenured{Ticker: ticker, Days: days})
		}
	}

	// Quick flips: exited within <= 10 days
	quickFlips := []models.Tenured{}
	for ticker, firstSeen := range s.FirstSeen {
		lastSeen := s.LastSeen[ticker]
		days := tenorDays(firstSeen, lastSeen)
		if lastSeen < latest && days <= 10 {
			quickFlips = append(quickFlips, models.Tenured{Ticker: ticker, Days: days})
		}
	}

	// Sector Pulse: stub until enrichment
	sectorPulse := []map[string]string{}

	// Forward Watchlist: union of radar + quick flips
	watchlist := []map[string]string{}
	for _, r := range radar {
		watchlist = append(watchlist, map[string]string{"Ticker": r["Ticker"], "Reason": r["Note"]})
	}
	for _, q := range quickFlips {
		watchlist = append(watchlist, map[string]string{"Ticker": q.Ticker, "Reason": "recent quick flip"})
	}

	// Template data
	data := map[string]any{
		"WeekStart":   weekStart,
		"WeekEnd":     weekEnd,
		"Radar":       radar,
		"Convictions": convictions,
		"QuickFlips":  quickFlips,
		"SectorPulse": sectorPulse,
		"Watchlist":   watchlist,
	}

	md, err := render.Markdown("templates/weekly.md.tmpl", data)
	must(err)

	fmt.Println(md)

	// --- write file ---
	safeDate := strings.ReplaceAll(weekEnd, "/", "-")
	outPath := fmt.Sprintf("out/weekly-%s.md", safeDate)

	if err := os.MkdirAll("out", 0o755); err != nil {
		panic(err)
	}
	if err := os.WriteFile(outPath, []byte(md), 0o644); err != nil {
		panic(err)
	}
	fmt.Println("Wrote", outPath)

	// subj := fmt.Sprintf("Weekly Playbook — %s → %s", weekStart, weekEnd)
	// if err := mailer.SendMarkdown(subj, md); err != nil {
	// 	panic(err)
	// }
}

func diff(prev, today []string) ([]string, []string) {
	pm := map[string]struct{}{}
	tm := map[string]struct{}{}
	for _, p := range prev {
		pm[p] = struct{}{}
	}
	for _, t := range today {
		tm[t] = struct{}{}
	}

	added, removed := []string{}, []string{}
	for t := range tm {
		if _, ok := pm[t]; !ok {
			added = append(added, t)
		}
	}
	for p := range pm {
		if _, ok := tm[p]; !ok {
			removed = append(removed, p)
		}
	}
	sort.Strings(added)
	sort.Strings(removed)
	return added, removed
}

func tenorDays(from, to string) int {
	const layout = "2006-01-02"
	a, err1 := time.Parse(layout, from)
	b, err2 := time.Parse(layout, to)
	if err1 != nil || err2 != nil {
		return 0
	}
	return int(b.Sub(a).Hours()/24 + 0.5)
}

func must(err error) {
	if err != nil {
		panic(err)
	}
}
