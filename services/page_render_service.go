package services

import (
	"encoding/json"
	"fmt"

	"31g.co.uk/triaging/db"
	"31g.co.uk/triaging/models"
	"31g.co.uk/triaging/util"
	"github.com/gin-gonic/gin"
	"github.com/rs/xid"
	"gorm.io/gorm"
)

func RenderPage(path, token, sessionId string, userDataDictionary map[string]interface{}) (map[string]interface{}, error) {
	key := sessionId
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
	formJson, _ := json.Marshal(page.BuilderJson)
	var themeJson string
	theme, err := db.GetThemeByAppIdAndFlowId(fmt.Sprintf("%d", appToken.AppId), fmt.Sprintf("%d", route.PageId), "PAGE")
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			themeJson = ""
		} else {
			return nil, err
		}
	} else {
		themeJsonStr, _ := json.Marshal(theme.Theme)
		themeJson = string(themeJsonStr)
	}

	responseData := gin.H{
		"dataDictionary": dataDictionary,
		"form":           string(formJson),
		"theme":          string(themeJson),
	}
	return responseData, nil
}

func CreateAppToken(userId, description, tokenType, scope string, appId, flowId uint) (string, error) {
	token := xid.New().String()

	db.CreateAppToken(&models.AppToken{
		UserId:      userId,
		FlowId:      flowId,
		AppId:       appId,
		Token:       util.EncodeStringToBase64(token),
		Description: description,
		Type:        tokenType,
		Scope:       scope,
	})
	return token, nil
}
