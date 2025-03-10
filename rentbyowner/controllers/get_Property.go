package controllers

import (
	"encoding/json"
	"log"
	"net/http"
	"strings"
	beego "github.com/beego/beego/v2/server/web"
)

type GetProperty struct {
	beego.Controller
}

type Amenities struct {
    Amenities map[string]string `json:"Amenities"`
}

type Counts struct {
    Reviews   int `json:"Reviews,omitempty"`
	Occupancy int `json:"Occupancy"`
}

type Property struct {
    Amenities     map[string]string `json:"Amenities"`
    FeatureImage  string            `json:"FeatureImage"`
    Price         float64           `json:"Price"`
    PropertyName  string            `json:"PropertyName"`
    PropertyType  string            `json:"PropertyType"`
    Counts        Counts            `json:"Counts"`
    ReviewScore   float64           `json:"ReviewScore"`
    StarRating    int               `json:"StarRating,omitempty"`
}

type Item struct {
    Property Property `json:"Property"`
    ID       string   `json:"ID"`
}

type Response struct {
    Items []Item `json:"Items"`
}

func (c *GetProperty) GetPropertyData() {
	itemIDs := c.Ctx.Input.Param(":itemIds")
	apiKey, err := beego.AppConfig.String("API_PROPERTY")
	if err != nil {
		log.Println("Error reading API_PROPERTY: " + err.Error())
	}
	itemIDsNew := strings.Split(itemIDs, ",")
	var allIDs string
	for i, id := range itemIDsNew {
		if i == len(itemIDsNew)-1 {
			allIDs += string(id)
		} else {
			allIDs += string(id) + ","
		}
	}
	modifiedUrl := strings.ReplaceAll(apiKey, "!REPLACE!", allIDs)
	productChan := make(chan []Item)
	errChan := make(chan error)
	go func() {
		req, err := http.NewRequest("GET", modifiedUrl, nil)
		if err != nil {
			errChan <- err
			return
		}
		req.Header.Set("Accept-Language", "en-US")
		req.Header.Set("Origin", "rentbyowner.com")
		client := &http.Client{}
		response, err := client.Do(req)
		if err != nil {
			errChan <- err
			return
		}
		defer response.Body.Close()
		var data Response
		if err := json.NewDecoder(response.Body).Decode(&data); err != nil {
			errChan <- err
			return
		}
		productChan <- data.Items
	}()
	select {
		case property := <-productChan:
			c.Data["json"] = property
			c.ServeJSON()
		case err := <-errChan:
			log.Println("Error fetching property: " + err.Error())
			c.Ctx.Output.SetStatus(http.StatusInternalServerError)
			c.Data["json"] = map[string]string{"error": err.Error()}
			c.ServeJSON()
		}
}