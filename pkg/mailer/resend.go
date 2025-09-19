package mailer

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

type emailReq struct {
	From    string   `json:"from"`
	To      []string `json:"to"`
	Subject string   `json:"subject"`
	Html    string   `json:"html"`
}

func SendMarkdown(subject, md string) error {
	api := os.Getenv("RESEND_API_KEY")
	from := os.Getenv("FROM_EMAIL")
	to := os.Getenv("SUBSTACK_POST_EMAIL")

	if api == "" || from == "" || to == "" {
		return fmt.Errorf("missing RESEND_API_KEY, FROM_EMAIL, or SUBSTACK_POST_EMAIL")
	}

	// Wrap markdown as <pre> so formatting survives
	html := "<pre style=\"font: 14px/1.4 ui-monospace, SFMono-Regular, Menlo, Monaco, Consolas, 'Liberation Mono', monospace; white-space: pre-wrap;\">" +
		templateEscape(md) + "</pre>"

	body, _ := json.Marshal(emailReq{
		From:    from,
		To:      []string{to},
		Subject: subject,
		Html:    html,
	})

	req, _ := http.NewRequest("POST", "https://api.resend.com/emails", bytes.NewReader(body))
	req.Header.Set("Authorization", "Bearer "+api)
	req.Header.Set("Content-Type", "application/json")

	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	if resp.StatusCode >= 300 {
		return fmt.Errorf("resend: status %d", resp.StatusCode)
	}
	return nil
}

func templateEscape(s string) string {
	replacements := map[string]string{"&": "&amp;", "<": "&lt;", ">": "&gt;"}
	out := ""
	for _, r := range s {
		if v, ok := replacements[string(r)]; ok {
			out += v
		} else {
			out += string(r)
		}
	}
	return out
}
