package main

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"strings"
)

type StoryInput struct {
	Story string `json:"story"`
}

type VFT struct {
	PartyA            string   `json:"party_a"`
	PartyB            string   `json:"party_b"`
	AgreementType     string   `json:"agreement_type"`
	AmountPaid        float64  `json:"amount_paid"`
	Currency          string   `json:"currency"`
	PerformanceStatus string   `json:"performance_status"`
	EvidenceAvailable []string `json:"evidence_available"`
}

type LAWONEResponse struct {
	VFT          VFT             `json:"vft"`
	LAWONEScore  float64         `json:"lawone_score"`
	Nodes        map[string]bool `json:"nodes"`
	NextQuestion string          `json:"next_question"`
}

func calculateScore(vft VFT) (float64, map[string]bool) {
	nodes := map[string]bool{"node_a": false, "node_b": false, "node_c": false, "node_d": false}
	score := 0.0
	if vft.PartyA != "" && vft.PartyB != "" {
		nodes["node_a"] = true
		score += 0.25
	}
	if vft.AmountPaid > 0 {
		nodes["node_b"] = true
		score += 0.25
	}
	if vft.PerformanceStatus == "breach" {
		nodes["node_c"] = true
		score += 0.25
	}
	if len(vft.EvidenceAvailable) > 0 {
		nodes["node_d"] = true
		score += 0.25
	}
	return score, nodes
}

func analyzeStory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}

	var input StoryInput
	json.NewDecoder(r.Body).Decode(&input)

	geminiKey := os.Getenv("GEMINI_API_KEY")
	prompt := fmt.Sprintf(`You are a legal fact extractor. Extract facts from this story and return ONLY valid JSON, no explanation, no markdown:
{"party_a":"","party_b":"","agreement_type":"","amount_paid":0,"currency":"INR","performance_status":"breach/pending/complete","evidence_available":[]}
Story: %s`, input.Story)

	reqBody := map[string]interface{}{
		"contents": []map[string]interface{}{
			{"parts": []map[string]string{{"text": prompt}}},
		},
	}
	reqBytes, _ := json.Marshal(reqBody)
	url := "https://generativelanguage.googleapis.com/v1beta/models/gemini-1.5-flash:generateContent?key=" + geminiKey

	resp, err := http.Post(url, "application/json", bytes.NewBuffer(reqBytes))
	if err != nil {
		// fallback if Gemini fails
		fallback := LAWONEResponse{
			VFT:          VFT{PartyA: "Unknown", PartyB: "Unknown", AmountPaid: 0, Currency: "INR", PerformanceStatus: "pending"},
			LAWONEScore:  0.25,
			Nodes:        map[string]bool{"node_a": false, "node_b": false, "node_c": false, "node_d": false},
			NextQuestion: "What is the full name of the other party?",
		}
		json.NewEncoder(w).Encode(fallback)
		return
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)

	var geminiResp map[string]interface{}
	json.Unmarshal(body, &geminiResp)

	var vft VFT
	vft.Currency = "INR"

	func() {
		defer func() { recover() }()
		candidates := geminiResp["candidates"].([]interface{})
		text := candidates[0].(map[string]interface{})["content"].(map[string]interface{})["parts"].([]interface{})[0].(map[string]interface{})["text"].(string)
		text = strings.TrimSpace(text)
		start := strings.Index(text, "{")
		end := strings.LastIndex(text, "}")
		if start >= 0 && end >= 0 {
			json.Unmarshal([]byte(text[start:end+1]), &vft)
		}
	}()

	score, nodes := calculateScore(vft)

	nextQ := ""
	if !nodes["node_a"] {
		nextQ = "What is the ful
