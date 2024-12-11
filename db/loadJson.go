package db

import (
	"encoding/json"
	"fmt"
	"os"

	"31g.co.uk/triaging/models"
)

var JsonData map[string]models.App

func LoadDataJson() {
	file, err := os.Open("data.json")
	if err != nil {
		fmt.Println("Error opening file:", err)
		return
	}
	defer file.Close()

	// Decode the JSON data into a map[string]interface{}
	var data map[string]map[string]models.App
	decoder := json.NewDecoder(file)
	err = decoder.Decode(&data)
	if err != nil {
		fmt.Println("Error decoding JSON:", err)
		return
	}
	appData := data["apps"]
	//var ok bool
	JsonData = appData

}
