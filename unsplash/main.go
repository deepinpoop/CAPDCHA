package main

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"strings"
)

const unsplashAccessKey = "ajjFLji7SSrb6iQ01Et3Z9iHFq9CmSosMQkl1lK3Ha4"

type unsplashPhoto struct {
	Urls struct {
		Regular string `json:"regular"`
	} `json:"urls"`
}

func nextBackgroundHandler(w http.ResponseWriter, r *http.Request) {
	apiURL := "https://api.unsplash.com/photos/random?query=nature&orientation=landscape"
	req, err := http.NewRequest("GET", apiURL, nil)
	if err != nil {
		http.Error(w, "failed to create request", http.StatusInternalServerError)
		return
	}
	req.Header.Set("Authorization", "Client-ID "+unsplashAccessKey)

	client := &http.Client{}
	metaResp, err := client.Do(req)
	if err != nil {
		http.Error(w, "failed to fetch from unsplash", http.StatusBadGateway)
		return
	}
	defer metaResp.Body.Close()

	if metaResp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(metaResp.Body)
		http.Error(w, fmt.Sprintf("unsplash error %d: %s", metaResp.StatusCode, body), http.StatusBadGateway)
		return
	}

	var photo unsplashPhoto
	if err := json.NewDecoder(metaResp.Body).Decode(&photo); err != nil {
		http.Error(w, "failed to parse unsplash response", http.StatusBadGateway)
		return
	}

	if photo.Urls.Regular == "" {
		http.Error(w, "no image url in response", http.StatusBadGateway)
		return
	}

	imgResp, err := client.Get(photo.Urls.Regular)
	if err != nil {
		http.Error(w, "failed to download image", http.StatusBadGateway)
		return
	}
	defer imgResp.Body.Close()

	if imgResp.StatusCode != http.StatusOK {
		http.Error(w, "image download failed", http.StatusBadGateway)
		return
	}

	contentType := imgResp.Header.Get("Content-Type")
	if !strings.HasPrefix(contentType, "image/") {
		http.Error(w, "unexpected content type", http.StatusBadGateway)
		return
	}

	w.Header().Set("Content-Type", contentType)
	w.Header().Set("Cache-Control", "public, max-age=86400")
	w.Header().Set("X-Content-Type-Options", "nosniff")

	io.Copy(w, imgResp.Body)
}

func main() {
	http.HandleFunc("/api/next-background", nextBackgroundHandler)
	http.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		http.ServeFile(w, r, "index.html")
	})

	fmt.Println("Server running on http://localhost:3002")
	log.Fatal(http.ListenAndServe(":3002", nil))
}
