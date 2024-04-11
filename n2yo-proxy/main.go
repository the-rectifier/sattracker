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

func doReq(url string) (int, map[string]interface{}) {
	resp, err := http.Get(url)

	if err != nil {
		return http.StatusBadRequest, gin.H{
			"message": "Request Failed",
		}
	}

	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)

	if err != nil {
		return http.StatusBadRequest, gin.H{
			"message": "Unable to Read Body",
		}
	}

	return http.StatusOK, parseBody(body)
}

func getAbove(c *gin.Context) {
	obsLat := c.Params.ByName("obs_lat")
	obsLon := c.Params.ByName("obs_lon")
	obsAlt := c.Params.ByName("obs_alt")
	srcRad := c.Params.ByName("src_rad")
	catId := c.Params.ByName("cat_id")

	apiKey := c.Query("apiKey")

	if apiKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid Satelleite ID or API Key",
		})
		return
	}

	url := fmt.Sprintf("https://api.n2yo.com/rest/v1/satellite/above/%s/%s/%s/%s/%s?apiKey=%s", obsLat, obsLon, obsAlt, srcRad, catId, apiKey)

	code, json := doReq(url)

	c.JSON(code, json)
}

func getRadioPasses(c *gin.Context) {
	satID := c.Params.ByName("id")
	obsLat := c.Params.ByName("obs_lat")
	obsLon := c.Params.ByName("obs_lon")
	obsAlt := c.Params.ByName("obs_alt")
	minElv := c.Params.ByName("min_elv")
	days := c.Params.ByName("days")

	apiKey := c.Query("apiKey")

	if apiKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid Satelleite ID or API Key",
		})
		return
	}

	url := fmt.Sprintf("https://api.n2yo.com/rest/v1/satellite/visualpasses/%s/%s/%s/%s/%s/%s?apiKey=%s", satID, obsLat, obsLon, obsAlt, minElv, days, apiKey)

	code, json := doReq(url)

	c.JSON(code, json)
}

func getVisualPasses(c *gin.Context) {
	satID := c.Params.ByName("id")
	obsLat := c.Params.ByName("obs_lat")
	obsLon := c.Params.ByName("obs_lon")
	obsAlt := c.Params.ByName("obs_alt")
	minVis := c.Params.ByName("min_vis")
	days := c.Params.ByName("days")

	apiKey := c.Query("apiKey")

	if apiKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid Satelleite ID or API Key",
		})
		return
	}

	url := fmt.Sprintf("https://api.n2yo.com/rest/v1/satellite/visualpasses/%s/%s/%s/%s/%s/%s?apiKey=%s", satID, obsLat, obsLon, obsAlt, minVis, days, apiKey)

	code, json := doReq(url)

	c.JSON(code, json)

}

func getSatPos(c *gin.Context) {
	satID := c.Params.ByName("id")
	obsLat := c.Params.ByName("obs_lat")
	obsLon := c.Params.ByName("obs_lon")
	obsAlt := c.Params.ByName("obs_alt")
	sec := c.Params.ByName("sec")

	apiKey := c.Query("apiKey")

	if apiKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid Satelleite ID or API Key",
		})
		return
	}

	url := fmt.Sprintf("https://api.n2yo.com/rest/v1/satellite/positions/%s/%s/%s/%s/%s?apiKey=%s", satID, obsLat, obsLon, obsAlt, sec, apiKey)

	code, json := doReq(url)

	c.JSON(code, json)

}

func getTLE(c *gin.Context) {
	satID := c.Params.ByName("id")
	apiKey := c.Query("apiKey")

	if satID == "" || apiKey == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"message": "Invalid Satelleite ID or API Key",
		})
		return
	}

	// fmt.Printf("Sat ID: %s\nAPI Key: %s\n", satID, apiKey)

	url := fmt.Sprintf("https://api.n2yo.com/rest/v1/satellite/tle/%s?apiKey=%s", satID, apiKey)

	code, json := doReq(url)

	c.JSON(code, json)

}

func main() {
	r := gin.Default()

	// /ping
	r.GET("/ping", ping)

	//  /getTLE/{id} -> https://api.n2yo.com/rest/v1/satellite/tle/25544?apiKey=589P8Q-SDRYX8-L842ZD-5Z9
	r.GET("/getTLE/:id", getTLE)

	// /getSatPos/{id}/{observer_lat}/{observer_lng}/{observer_alt}/{seconds} -> https://api.n2yo.com/rest/v1/satellite/positions/25544/41.702/-76.014/0/2/&apiKey=589P8Q-SDRYX8-L842ZD-5Z9
	r.GET("/getSatPos/:id/:obs_lat/:obs_lon/:obs_alt/:sec", getSatPos)

	//  /getVisualPasses/{id}/{observer_lat}/{observer_lng}/{observer_alt}/{days}/{min_visibility} -> https://api.n2yo.com/rest/v1/satellite/visualpasses/25544/41.702/-76.014/0/2/300/&apiKey=589P8Q-SDRYX8-L842ZD-5Z9
	r.GET("/getVisualPasses/:id/:obs_lat/:obs_lon/:obs_alt/:days/:min_vis", getVisualPasses)

	// /getRadioPasses/{id}/{observer_lat}/{observer_lng}/{observer_alt}/{days}/{min_elevation} -> https://api.n2yo.com/rest/v1/satellite/radiopasses/25544/41.702/-76.014/0/2/40/&apiKey=589P8Q-SDRYX8-L842ZD-5Z9
	r.GET("/getRadioPasses/:id/:obs_lat/:obs_lon/:obs_alt/:days/:min_elv", getRadioPasses)

	// /getAbove/{observer_lat}/{observer_lng}/{observer_alt}/{search_radius}/{category_id} ->  https://api.n2yo.com/rest/v1/satellite/above/41.702/-76.014/0/70/18/&apiKey=589P8Q-SDRYX8-L842ZD-5Z9
	r.GET("/getAbove/:obs_lat/:obs_lon/:obs_alt/:src_rad/:cat_id", getAbove)

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
