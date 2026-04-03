package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"os"
)

func analyzeStory(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Access-Control-Allow-Origin", "*")
	w.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS")
	w.Header().Set("Access-Control-Allow-Headers", "Content-Type")
	w.Header().Set("Content-Type", "application/json")
	if r.Method == "OPTIONS" {
		w.WriteHeader(http.StatusOK)
		return
	}
	result := map[string]interface{}{
		"lawone_score": 0.50,
		"vft": map[string]interface{}{
			"party_a":            "Contractor",
			"party_b":            "Client",
			"amount_paid":        50000,
			"currency":           "INR",
			"performance_status": "breach",
		},
		"nodes": map[string]bool{
			"node_a": true,
			"node_b": true,
			"node_c": false,
			"node_d": false,
		},
		"next_question": "Do you have a receipt or bank transfer screenshot for the payment?",
	}
	json.NewEncoder(w).Encode(result)
}

func main() {
	http.HandleFunc("/analyze", analyzeStory)
	port := os.Getenv("PORT")
	if port == "" {
		port = "8080"
	}
	fmt.Println("LAWONE backend running on port " + port)
	http.ListenAndServe(":"+port, nil)
}
