package main

import (
	"crypto/tls"
	"encoding/json"
	"fmt"
	"io"
	"net/http"

	"github.com/gin-gonic/gin"
)

func ping(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"message": "pong",
	})
}

func parseBody(body []byte) map[string]interface{} {
	var result map[string]interface{}
	json.Unmarshal(body, &result)

	return result
}

func getTLE(c *gin.Context) {
	satID := c.Query("id")
	apiKey := c.Query("apiKey")

	if satID == "" || apiKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid Satelleite ID or API Key",
		})
		return
	}

	// fmt.Printf("Sat ID: %s\nAPI Key: %s\n", satID, apiKey)

	url := fmt.Sprintf("https://api.n2yo.com/rest/v1/satellite/tle/%s&apiKey=%s", satID, apiKey)

	resp, err := http.Get(url)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Request Failed",
		})
		return
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Unable to Read Body",
		})
		return
	}

	c.JSON(http.StatusOK, parseBody(body))

}

func main() {
	r := gin.Default()

	// /ping
	r.GET("/ping", ping)
	// /getTLE&ID=XXXXX&KEY=YYYYY -> https://api.n2yo.com/rest/v1/satellite/tle/XXXXX&apiKey=YYYYYY

	r.GET("/getTLE", getTLE)

	cfg := &tls.Config{
		MinVersion:               tls.VersionTLS13,
		CurvePreferences:         []tls.CurveID{tls.CurveP521, tls.CurveP384, tls.CurveP256},
		PreferServerCipherSuites: true,
		CipherSuites: []uint16{
			tls.TLS_ECDHE_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_ECDHE_RSA_WITH_AES_256_CBC_SHA,
			tls.TLS_RSA_WITH_AES_256_GCM_SHA384,
			tls.TLS_RSA_WITH_AES_256_CBC_SHA,
		},
	}

	server := http.Server{
		Addr:      ":9443",
		Handler:   r,
		TLSConfig: cfg,
	}

	err := server.ListenAndServeTLS("tls.crt", "tls.key")

	if err != nil {
		fmt.Printf("Error: %s", err.Error())
	}
}
