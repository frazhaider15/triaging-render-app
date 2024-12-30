package services

import (
	"fmt"

	"31g.co.uk/triaging/db"
	"31g.co.uk/triaging/util"
	"github.com/gin-gonic/gin"
)

func RenderPage(path, token string, userDataDictionary map[string]interface{}) (map[string]interface{}, error) {
	key := token
	var dataDictionary map[string]interface{}
	var err error

	appToken, err := db.GetAppTokenByToken(util.EncodeStringToBase64(token))
	if err != nil {
		return nil, err
	}

	if !CheckIfDataDictionaryInitialised(key) {
		dataDictionary, err = CreateEmptyDataDictionary(fmt.Sprintf("%v", appToken.AppId), key)
		if err != nil {
			return nil, err
		}
	} else {
		dataDictionary = GetDataDictionary(key)
	}
	updateNestedMaps(dataDictionary, userDataDictionary)

	route, err := db.GetPageRouteByPathAndAppId(appToken.AppId, path)
	if err != nil {
		return nil, err
	}

	page, err := db.GetFormById(float64(route.PageId))
	if err != nil {
		return nil, err
	}
	
	responseData := gin.H{
		"dataDictionary": dataDictionary,
		"form":           page.BuilderJson,
		//"theme":          themeJson,
	}
	return responseData, nil
}
