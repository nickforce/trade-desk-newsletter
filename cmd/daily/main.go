package main

import (
	"encoding/json"
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
	// Load rotations.json
	b, err := os.ReadFile("data/rotations.json")
	must(err)
	var payload models.RotationsPayload
	must(json.Unmarshal(b, &payload))

	st := state.New("data/state.json")
	s, err := st.Load()
	must(err)

	today := payload.LatestDate
	todayHoldings := payload.StockTickers

	// Get yesterday's holdings
	prevDay := prevDateKey(s, today)
	prevHoldings := s.HoldingsByDay[prevDay]

	// Save today's holdings
	s.HoldingsByDay[today] = todayHoldings
	// Update first_seen/last_seen
	for _, t := range todayHoldings {
		if _, ok := s.FirstSeen[t]; !ok {
			s.FirstSeen[t] = today
		}
		s.LastSeen[t] = today
	}
	must(st.Save(s))

	// Compute adds/removals
	added, removed := diff(prevHoldings, todayHoldings)

	// If no changes, skip posting
	if len(added) == 0 && len(removed) == 0 {
		fmt.Println("No changes today. Skipping Substack/Discord post.")
		return
	}

	// Convictions: names present >= 15 days
	convictions := []models.Tenured{}
	for ticker, firstSeen := range s.FirstSeen {
		lastSeen, ok := s.LastSeen[ticker]
		if !ok {
			continue
		}
		days := tenorDays(firstSeen, lastSeen)
		if days >= 15 {
			convictions = append(convictions, models.Tenured{Ticker: ticker, Days: days})
		}
	}

	// Quick flips: exited within <= 10 days
	quickFlips := []models.Tenured{}
	for _, t := range removed {
		fs := s.FirstSeen[t]
		ls := s.LastSeen[t]
		days := tenorDays(fs, ls)
		if days <= 10 {
			quickFlips = append(quickFlips, models.Tenured{Ticker: t, Days: days})
		}
	}

	// Build template data
	data := map[string]any{
		"Date":        today,
		"Added":       added,
		"Removed":     removed,
		"Convictions": convictions,
		"QuickFlips":  quickFlips,
	}

	md, err := render.Markdown("templates/daily.md.tmpl", data)
	must(err)

	fmt.Println(md)

	// --- write file ---
	safeDate := strings.ReplaceAll(today, "/", "-")
	outPath := fmt.Sprintf("out/daily-%s.md", safeDate)
	if err := os.MkdirAll("out", 0o755); err != nil {
		panic(err)
	}
	if err := os.WriteFile(outPath, []byte(md), 0o644); err != nil {
		panic(err)
	}
	fmt.Println("Wrote", outPath)

	// subj := fmt.Sprintf("In Play — %s", today)
	// if err := mailer.SendMarkdown(subj, md); err != nil {
	// 	panic(err)
	// }
}

func prevDateKey(st *models.State, today string) string {
	keys := make([]string, 0, len(st.HoldingsByDay))
	for k := range st.HoldingsByDay {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	if len(keys) == 0 {
		return ""
	}
	return keys[len(keys)-1]
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
