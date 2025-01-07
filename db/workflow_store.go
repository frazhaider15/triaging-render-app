package db

import (
	"fmt"

	"31g.co.uk/triaging/models"
	"gorm.io/gorm"
)

func GetFlowById(flowId int64, appId string) (models.Flow, error) {
	for i, flow := range JsonData[appId].Details.Flows {
		if flow.ID == uint(flowId) && !flow.DeletedAt.Valid {
			return JsonData[appId].Details.Flows[i], nil
		}
	}
	return models.Flow{}, gorm.ErrRecordNotFound
}

func GetAppTokenByToken(inputToken string) (models.AppToken, error) {
	for _, app := range JsonData {
		for i, token := range app.Details.AppTokens {
			if token.Token == inputToken && !token.DeletedAt.Valid {
				return app.Details.AppTokens[i], nil
			}
		}
	}
	return models.AppToken{}, gorm.ErrRecordNotFound
}

func GetAllDataTypesByAppId(appId uint) ([]models.DataTypeWithFields, error) {
	dataTypes := JsonData[fmt.Sprintf("%v", appId)].Details.DataTypes
	return dataTypes, nil
}

func GetAllDataFieldsBydataTypeID(dataTypeID uint) ([]models.DataTypeField, error) {
	for _, app := range JsonData {
		for i, dt := range app.Details.DataTypes {
			if dt.ID == dataTypeID && !dt.DeletedAt.Valid {
				return app.Details.DataTypes[i].Fields, nil
			}
		}
	}
	return nil, gorm.ErrRecordNotFound
}

func GetThemeByAppIdAndFlowId(appId, flowId string) (models.Theme, error) {
	var theme models.Theme
	flowTheme, err := GetFlowThemeByFlowId(flowId)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			theme, err = GetDefaultTheme(appId)
			if err != nil {
				return theme, err
			}
			return theme, nil
		}
		return theme, err
	}
	theme, err = GetThemeById(flowTheme.ThemeId)
	if err != nil {
		return models.Theme{}, err
	}
	return theme, nil
}

func GetFlowThemeByFlowId(flowId string) (models.FlowTheme, error) {
	for _, app := range JsonData {
		for i, ft := range app.Details.FlowThemes {
			if ft.FlowId == flowId && !ft.DeletedAt.Valid {
				return app.Details.FlowThemes[i], nil
			}
		}
	}
	return models.FlowTheme{}, gorm.ErrRecordNotFound
}

func GetDefaultTheme(appId string) (models.Theme, error) {
	for i, theme := range JsonData[appId].Details.Themes {
		if theme.IsDefault && !theme.DeletedAt.Valid {
			return JsonData[appId].Details.Themes[i], nil
		}
	}
	return models.Theme{}, gorm.ErrRecordNotFound
}

func GetThemeById(id string) (models.Theme, error) {
	for _, app := range JsonData {
		for i, theme := range app.Details.Themes {
			if fmt.Sprintf("%v", theme.ID) == id && !theme.DeletedAt.Valid {
				return app.Details.Themes[i], nil
			}
		}
	}
	return models.Theme{}, gorm.ErrRecordNotFound
}

func GetFormById(id float64) (models.Form, error) {
	for _, app := range JsonData {
		for i, form := range app.Details.Forms {
			if float64(form.ID) == id && !form.DeletedAt.Valid {
				return app.Details.Forms[i], nil
			}
		}
	}
	return models.Form{}, gorm.ErrRecordNotFound
}

func GetPageRouteByPathAndAppId(appId uint, path string) (models.PageRoute, error) {
	aid := fmt.Sprintf("%v", appId)
	for i, route := range JsonData[aid].Details.PageRoutes {
		if route.Path == path && !route.DeletedAt.Valid {
			return JsonData[aid].Details.PageRoutes[i], nil
		}
	}
	return models.PageRoute{}, gorm.ErrRecordNotFound
}
