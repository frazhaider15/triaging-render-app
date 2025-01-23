package controllers

import (
	"encoding/json"
	"fmt"
	"net/http"

	"31g.co.uk/triaging/db"
	"31g.co.uk/triaging/services"
	"github.com/gin-gonic/gin"
	"gonih.org/stack"
	"gorm.io/gorm"
)

func Testing(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"message": "triaging app ok response"})
}
func WorkFlowRendering(c *gin.Context) {
	appToken := c.Query("token")
	sessionId := c.Query("sessionId")
	var appId, flowId string

	appIdint, flowIdint, err := services.ValidateAppToken(appToken)
	if err != nil {
		appId = c.Query("appId")
		flowId = c.Query("flowId")
	} else {
		appId = fmt.Sprintf("%v", appIdint)
		flowId = fmt.Sprintf("%v", flowIdint)
	}
	if appId == "" || flowId == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "app id or flow id not found"})
		return
	}
	currentNode := c.Query("currentNode")

	var inMemoryMap map[string]interface{}
	c.BindJSON(&inMemoryMap)
	key := fmt.Sprintf("%v_%v", appId, flowId)
	_, ok := services.CurrentFlow[key]
	if !ok || currentNode == "" {
		services.CurrentFlow[key] = map[string]interface{}{
			"flowLevel":     0,
			"previousFlows": stack.Stack[services.PreviousFlow]{},
		}
	}
	if currentNode == "" {
		services.ClearDataDictionary(key)
	}
	resp, err := services.RenderWorkFlow(key, appId, flowId, currentNode, sessionId, inMemoryMap)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func CheckLastNode(c *gin.Context) {
	appToken := c.Query("token")
	var appId, flowId string

	appIdint, flowIdint, err := services.ValidateAppToken(appToken)
	if err != nil {
		appId = c.Query("appId")
		flowId = c.Query("flowId")
	} else {
		appId = fmt.Sprintf("%v", appIdint)
		flowId = fmt.Sprintf("%v", flowIdint)
	}
	if appId == "" || flowId == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "app id or flow id not found"})
		return
	}
	currentNode := c.Query("currentNode")
	key := fmt.Sprintf("%v_%v", appId, flowId)

	found, err := services.CheckLastNode(key, appId, flowId, currentNode)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	if found {
		c.JSON(http.StatusOK, "This is the last node")
	} else {
		c.JSON(http.StatusOK, "This is not the last node")
	}
}

func PreviousForm(c *gin.Context) {

	appToken := c.Query("token")
	sessionId := c.Query("sessionId")
	var appId, flowId string
	var err error

	appIdint, flowIdint, err := services.ValidateAppToken(appToken)
	if err != nil {
		appId = c.Query("appId")
		flowId = c.Query("flowId")
	} else {
		appId = fmt.Sprintf("%v", appIdint)
		flowId = fmt.Sprintf("%v", flowIdint)
	}

	if appId == "" || flowId == "" {
		c.JSON(http.StatusBadRequest, gin.H{"message": "appId or flowId not found"})
		return
	}

	//var newCurrentNode string
	key := appId + "_" + flowId
	nodetype := ""
	var prevNode, node services.FormHistory

	for nodetype != "form" {
		prevNode, node, err = services.PopNode(key)
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}

		if prevNode.FlowId != node.FlowId {
			services.CurrentFlow[key]["flowId"] = node.FlowId
			if prevNode.FlowId == flowId && node.FlowId != flowId {
				// we are going from main flow to subflow
				services.IncrementFlowLevel(key)
				services.AddPreviousFlow(key, prevNode.FlowId, prevNode.NodeId)
			}
			if prevNode.FlowId != flowId && node.FlowId == flowId {
				// we are going from subflow to main flow
				services.DecrementFlowLevel(key)
				services.GetPreviousFlow(key)
			}
			if prevNode.FlowId != flowId && node.FlowId != flowId {
				// we are going from subflow to subflow
				fmt.Println("flow to flow")
			}
		}

		nodetype = node.NodeType
	}

	var themeJson string
	theme, err := db.GetThemeByAppIdAndFlowId(appId, node.FlowId, "FLOW")
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			themeJson = ""
		} else {
			c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
			return
		}
	} else {
		themeJsonByte, _ := json.Marshal(theme.Theme)
		themeJson = string(themeJsonByte)
	}

	dataDictionary := services.GetDataDictionary(sessionId)
	resp := gin.H{
		"nodeId":         node.NodeId,
		"dataDictionary": dataDictionary,
		"form":           node.Form,
		"theme":          themeJson,
	}

	c.JSON(http.StatusOK, gin.H{"data": resp})
}

func RenderPage(ctx *gin.Context) {
	path := ctx.Query("path")
	appToken := ctx.Query("token")
	sessionId := ctx.Query("sessionId")

	var userDataDictionary map[string]interface{}
	ctx.BindJSON(&userDataDictionary)

	resp, err := services.RenderPage(path, appToken, sessionId, userDataDictionary)
	if err != nil {
		ctx.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}
	ctx.JSON(http.StatusOK, gin.H{"data": resp})
}

func UpdateDataDictionaryBySessionId(c *gin.Context) {
	sessionId := c.Query("sessionId")
	token := c.Query("token")

	appId, _, err := services.ValidateAppToken(token)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	var dataDictionary map[string]interface{}
	c.BindJSON(&dataDictionary)
	services.UpdateDataDictionaryBySessionId(appId,sessionId, dataDictionary)
	c.JSON(http.StatusOK, "Data dictionary updated successfully")
}
