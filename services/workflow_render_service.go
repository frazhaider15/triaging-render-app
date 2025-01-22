package services

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"regexp"
	"strconv"
	"strings"

	"31g.co.uk/triaging/constants"
	"31g.co.uk/triaging/db"
	"31g.co.uk/triaging/models"
	"31g.co.uk/triaging/util"
	"github.com/gin-gonic/gin"
	"github.com/spf13/cast"
	"gonih.org/stack"
	"gorm.io/gorm"
)

var dataDictionary = make(map[string]map[string]interface{})
var FormHistoryMap = make(map[string]stack.Stack[FormHistory])
var CurrentFlow = make(map[string]map[string]interface{})

func RenderWorkFlow(key, appId, flowId, currentNode string, inMemoryMap map[string]interface{}) (map[string]interface{}, error) {
	var (
		dataDictionary map[string]interface{}
		err            error
	)
	resp := make(map[string]interface{})
	if appId == "" || flowId == "" {
		return nil, fmt.Errorf("app id or flowId not found")
	}

	if !CheckIfDataDictionaryInitialised(key) {
		dataDictionary, err = CreateEmptyDataDictionary(appId, key)
		if err != nil {
			return nil, err
		}
	} else {
		dataDictionary = GetDataDictionary(key)
	}
	updateNestedMaps(dataDictionary, inMemoryMap)

	targetNodeType := "api"
	for targetNodeType != "form" {
		_, ok := CurrentFlow[key]["flowId"]
		if ok {
			flowId = CurrentFlow[key]["flowId"].(string)
		}
		var builderJson models.Workflow

		var flow models.Flow

		flow, err = db.GetFlowById(cast.ToInt64(flowId), appId)
		if err != nil {
			return nil, err
		}
		builderJson = flow.BuilderJson

		currentNodeId, nodeType := GetCurrentNode(currentNode, builderJson)
		if nodeType == "end" {
			ClearDataDictionary(key)
			nodeData, ok := builderJson.FlowWorkspace.NodesData[currentNodeId]
			if !ok {
				resp["type"] = "end"
				return resp, nil
			}
			resp = nodeData.(map[string]interface{})
			resp["type"] = "end"
			return resp, nil
		}

		targetNodeId := getTargetNodeByCurrentNodeId(currentNodeId, builderJson)
		targetNode, err := GetNodeById(targetNodeId, builderJson)
		if err != nil {
			return resp, err
		}
		targetNodeType = targetNode.Type
		nodeID := targetNode.ID
		if targetNodeType == "form" {
			resp, err = getFormByFormId(key, nodeID, appId, flowId, builderJson, dataDictionary)
			if err != nil {
				return nil, err
			}
			currentNode = nodeID
		}
		if targetNodeType == "api" {
			dataDictionary, err = handleApiNode(key, flowId, nodeID, dataDictionary, builderJson)
			if err != nil {
				return nil, err
			}
			currentNode = nodeID
		}
		if targetNodeType == "condition" {
			resp, err = getFormOnCondition(key, appId, flowId, nodeID, builderJson, dataDictionary)
			if err != nil {
				return nil, err
			}
			return resp, err
		}
		if targetNodeType == "subFlow" {
			IncrementFlowLevel(key)
			AddPreviousFlow(key, flowId, nodeID)
			resp, err = handleFlowNode(key, appId, nodeID, builderJson, dataDictionary)
			if err != nil {
				return nil, err
			}
			if resp != nil {
				return resp, nil
			}
			currentNode = nodeID
		}
		if targetNodeType == "end" {
			if getFlowLevel(key) == 0 {
				pushNode(key, flowId, nodeID, "", targetNodeType)
				nodeData, ok := builderJson.FlowWorkspace.NodesData[nodeID]
				if !ok {
					resp["type"] = "end"
					return resp, nil
				}
				resp = nodeData.(map[string]interface{})
				resp["type"] = "end"
				break
			}
			// going back from subflow to main flow
			prevFlow := GetPreviousFlow(key)
			CurrentFlow[key]["flowId"] = prevFlow.FlowId
			currentNode = prevFlow.NodeId
			DecrementFlowLevel(key)
			subflowNodeId, _ := getNodeIdByFlowId(appId, prevFlow.FlowId, flowId)
			if subflowNodeId != "" {
				pushNode(key, prevFlow.FlowId, subflowNodeId, "", "subFlow")
			}

		}

	}
	return resp, nil
}

func CheckLastNode(key, appId, flowId, currentNodeId string) (bool, error) {
	currentFlowId := CurrentFlow[key]["flowId"].(string)
	var builderJson models.Workflow
	flow, err := db.GetFlowById(cast.ToInt64(flowId), appId)
	if err != nil {
		return false, err
	}

	// if err := json.Unmarshal([]byte(flow.BuilderJson.(string)), &builderJson); err != nil {
	// 	return false, err
	// }
	builderJson = flow.BuilderJson
	if currentFlowId == flowId {
		currentNodeId, nodeType := GetCurrentNode(currentNodeId, builderJson)
		if nodeType == "end" {
			return true, nil
		}

		targetNodeId := getTargetNodeByCurrentNodeId(currentNodeId, builderJson)
		targetNode, err := GetNodeById(targetNodeId, builderJson)
		if err != nil {
			return false, err
		}
		targetNodeType := targetNode.Type
		if targetNodeType == "end" {
			return true, nil
		}

	} else {
		mainNode := GetMainFlowNode(key, flowId)
		targetNodeId := getTargetNodeByCurrentNodeId(mainNode, builderJson)
		targetNode, err := GetNodeById(targetNodeId, builderJson)
		if err != nil {
			return false, err
		}
		targetNodeType := targetNode.Type
		if targetNodeType == "end" {
			return true, nil
		}
	}
	return false, nil
}
func CreateEmptyDataDictionary(appId, key string) (map[string]interface{}, error) {
	appIdInt, err := strconv.ParseUint(appId, 10, 32)
	if err != nil {
		return nil, err
	}
	dataTypes, err := db.GetAllDataTypesByAppId(uint(appIdInt))
	if err != nil {
		return nil, err
	}
	dataDictionary[key] = make(map[string]interface{})
	for _, dt := range dataTypes {
		dataFields := dt.Fields

		// Create a map for the current data type
		dataTypeMap := make(map[string]interface{})
		dataDictionary[key][dt.Name] = dataTypeMap

		// Process each data field
		for _, df := range dataFields {
			if df.TypeGroup == constants.DataFieldGroupCustom.String() {
				// Handle nested custom data types
				handleCustomDataTypes(df, dataTypeMap)
			} else if df.TypeGroup == constants.DataFieldGroupPrimitive.String() {
				// Handle primitive data types
				handlePrimitiveDataTypes(df, dataTypeMap)
			}
		}
	}
	dataDictionary[key]["errors"] = []models.DataDictionaryError{}

	return dataDictionary[key], nil
}

func CheckIfDataDictionaryInitialised(key string) bool {
	_, found := dataDictionary[key]
	return found
}

func ClearDataDictionary(key string) {
	delete(dataDictionary, key)
}

func UpdateDataDictionaryAfterApiCall(apiRes map[string]interface{}, key, mapType string) map[string]interface{} {
	dictionaryData, found := dataDictionary[key][mapType]
	if !found {
		return dataDictionary[key]
	}

	// Type assertion to convert interface{} to map[string]interface{}
	dictionaryMap, ok := dictionaryData.(map[string]interface{})
	if !ok {
		return dataDictionary[key]
	}
	for k, val := range apiRes {
		_, ok := dictionaryMap[k]
		if ok {
			dictionaryMap[k] = val
		}
	}
	dataDictionary[key][mapType] = dictionaryMap
	return dataDictionary[key]
}

func AddErrorInDataDictionary(errMsg, code, key string) map[string]interface{} {
	dictionaryData, found := dataDictionary[key]["errors"]
	if !found {
		return dataDictionary[key]
	}

	// Type assertion to convert interface{} to map[string]interface{}
	errorsArray, ok := dictionaryData.([]models.DataDictionaryError)
	if !ok {
		return dataDictionary[key]
	}
	errorsArray = append(errorsArray, models.DataDictionaryError{
		ErrorMessage: errMsg,
		ErrorCode:    code,
	})

	dataDictionary[key]["errors"] = errorsArray
	return dataDictionary[key]
}

func GetDataDictionary(key string) map[string]interface{} {
	return dataDictionary[key]
}

func PopNode(key string) (FormHistory, FormHistory, error) {
	formHistory, exists := FormHistoryMap[key]
	if !exists || len(formHistory) <= 1 {
		return FormHistory{}, FormHistory{}, errors.New("no previous node in the history")
	}
	// Pop the last element
	prevNode := formHistory.Pop()
	FormHistoryMap[key] = formHistory
	return prevNode, formHistory.Top(), nil
}

func GetLatestNode(key string) (FormHistory, error) {
	formHistory, exists := FormHistoryMap[key]
	if !exists || len(formHistory) <= 1 {
		return FormHistory{}, errors.New("no previous node in the history")
	}
	return formHistory.Top(), nil
}

//////////////////////////////////////////////////////////////////////////////////////////////////
////////////////////////// Helper Functions //////////////////////////////////////////////////////
///////////////////////////////////////////////////////////////////////////////////////////////`///

func handlePrimitiveDataTypes(df models.DataTypeField, dataTypeMap map[string]interface{}) {
	if df.IsArray {
		dataTypeMap[df.Name] = getPrimitiveArrayType(df.Type)
	} else {
		dataTypeMap[df.Name] = getPrimitiveType(df.Type)
	}
}

func getPrimitiveType(typeName string) interface{} {
	switch typeName {
	case constants.DataFieldTypeString.String():
		return ""
	case constants.DataFieldTypeInteger.String():
		return 0
	case constants.DataFieldTypeFloat.String():
		return 0.0
	case constants.DataFieldTypeBool.String():
		return false
	default:
		return nil
	}
}

func getPrimitiveArrayType(typeName string) interface{} {
	switch typeName {
	case constants.DataFieldTypeString.String():
		return []string{}
	case constants.DataFieldTypeInteger.String():
		return []int64{}
	case constants.DataFieldTypeFloat.String():
		return []float64{}
	case constants.DataFieldTypeBool.String():
		return []bool{}
	default:
		return nil
	}
}

func handleCustomDataTypes(df models.DataTypeField, dataTypeMap map[string]interface{}) {
	subDataTypeMap := make(map[string]interface{})
	dataTypeMap[df.Name] = subDataTypeMap

	// Recursively process nested custom data types
	subFields, err := db.GetAllDataFieldsBydataTypeID(cast.ToUint(df.Type))
	if err != nil {
		// Handle error
		return
	}

	for _, subField := range subFields {
		if subField.TypeGroup == constants.DataFieldGroupCustom.String() {
			handleCustomDataTypes(subField, subDataTypeMap)
		} else if subField.TypeGroup == constants.DataFieldGroupPrimitive.String() {
			handlePrimitiveDataTypes(subField, subDataTypeMap)
		}
	}
}

func GetCurrentNode(currentNodeId string, builderJson models.Workflow) (string, string) {
	var nodeId, nodeType string
	if currentNodeId == "" {
		for _, node := range builderJson.FlowWorkspace.Nodes {
			if node.Type == "start" {
				nodeId = node.ID
			}
		}
	} else {
		for _, node := range builderJson.FlowWorkspace.Nodes {
			if node.ID == currentNodeId {
				nodeId = node.ID
				nodeType = node.Type
			}
		}
	}
	return nodeId, nodeType
}

func getTargetNodeByCurrentNodeId(currentNodeId string, builderJson models.Workflow) string {
	var targetNodeId string
	for _, edge := range builderJson.FlowWorkspace.Edges {
		if edge.Source == currentNodeId {
			targetNodeId = edge.Target
		}
	}
	return targetNodeId
}

func GetNodeById(targetNode string, builderJson models.Workflow) (models.Node, error) {
	for _, node := range builderJson.FlowWorkspace.Nodes {
		if node.ID == targetNode {
			return node, nil
		}
	}
	return models.Node{}, fmt.Errorf("no node found by given id")
}

func getFormByFormId(key, nodeID, appId, flowId string, builderJson models.Workflow, dataDictionary map[string]interface{}) (gin.H, error) {
	var responseData gin.H

	nodeData := builderJson.FlowWorkspace.NodesData[nodeID]

	formIdStruct := nodeData.(map[string]interface{})

	formId := formIdStruct["formId"].(float64)

	var form models.Form
	form, err := db.GetFormById(formId)
	if err != nil {
		return nil, err
	}

	var themeJson string
	theme, err := db.GetThemeByAppIdAndFlowId(appId, flowId, "FLOW")
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			themeJson = ""
		} else {
			return nil, err
		}
	} else {
		themeJsonData, err := json.Marshal(theme.Theme)
		if err != nil {
			return nil, err
		}
		themeJson = string(themeJsonData)
	}
	formJson, _ := json.Marshal(form.BuilderJson)
	responseData = gin.H{
		"nodeId":         nodeID,
		"dataDictionary": dataDictionary,
		"form":           string(formJson),
		"theme":          themeJson,
	}

	pushNode(key, flowId, nodeID, string(formJson), "form")
	return responseData, nil
}

func handleApiNode(key, flowId, nodeID string, dataDictionary map[string]interface{}, builderJson models.Workflow) (map[string]interface{}, error) {
	apiNodeDataInterface := builderJson.FlowWorkspace.NodesData[nodeID]
	if apiNodeDataInterface == nil {
		return nil, fmt.Errorf("api node data not found")
	}
	apiNodeDataMap := apiNodeDataInterface.(map[string]interface{})

	byteData, err := json.Marshal(apiNodeDataMap)
	if err != nil {
		return nil, err
	}
	var apiNodeData models.ApiNodeData
	err = json.Unmarshal(byteData, &apiNodeData)
	if err != nil {
		return nil, err
	}
	headers := make(map[string]string)
	reqUrl := apiNodeData.API.URL
	placeHolders := getPlaceholders(reqUrl)
	for _, placeHolder := range placeHolders {
		placeHolderValue, _ := getValueFromNestedMap(dataDictionary, placeHolder)
		value := fmt.Sprintf("%v", placeHolderValue)
		if value != "" {
			reqUrl = replacePlaceholder(reqUrl, placeHolder, value)
		}
	}
	reqUrl = reqUrl + "?"
	for _, header := range apiNodeData.API.Headers {
		if header.KeyItem != "" {
			if stringsContainCurlyBraces(header.ValueItem) {
				headerValue, _ := getValueFromNestedMap(dataDictionary, header.ValueItem)
				headers[header.KeyItem] = fmt.Sprintf("%v", headerValue)
			} else {
				headers[header.KeyItem] = header.ValueItem
			}

		}
	}
	for _, param := range apiNodeData.API.QueryParams {
		if stringsContainCurlyBraces(param.ValueItem) {
			paramValue, _ := getValueFromNestedMap(dataDictionary, param.ValueItem)
			reqUrl = reqUrl + param.KeyItem + "=" + fmt.Sprintf("%v", paramValue) + "&"
		} else {
			reqUrl = reqUrl + param.KeyItem + "=" + param.ValueItem + "&"
		}
	}
	bodyplaceHolders := getPlaceholders(apiNodeData.API.Body)
	for _, bodyplaceHolder := range bodyplaceHolders {
		placeHolderValue, _ := getValueFromNestedMap(dataDictionary, bodyplaceHolder)
		value := fmt.Sprintf("%v", placeHolderValue)

		apiNodeData.API.Body = replacePlaceholder(apiNodeData.API.Body, bodyplaceHolder, value)

	}
	res, code, err := util.MakeHttpRequestWithRawJsonBody(apiNodeData.API.ReqMethod, reqUrl, apiNodeData.API.Body, headers)
	if err != nil {
		dd := AddErrorInDataDictionary(err.Error(), fmt.Sprintf("%v", code), key)
		return dd, nil
	}

	// Create a map to store the JSON data
	var resultMap map[string]interface{}

	// Unmarshal the JSON data into the map
	err = json.Unmarshal(res, &resultMap)
	if err != nil {
		return nil, err
	}
	mapType := strings.Trim(apiNodeData.API.MapType, "{}")
	dd := UpdateDataDictionaryAfterApiCall(resultMap, key, mapType)
	pushNode(key, flowId, nodeID, "", "api")
	return dd, nil
}

func handleFlowNode(key, appId, nodeID string, builderJson models.Workflow, dataDictionary map[string]interface{}) (map[string]interface{}, error) {
	nodeData := builderJson.FlowWorkspace.NodesData[nodeID]

	formIdStruct := nodeData.(map[string]interface{})

	flowId := formIdStruct["flowId"].(float64)
	CurrentFlow[key]["flowId"] = fmt.Sprintf("%v", flowId)
	resp, err := RenderWorkFlow(key, appId, fmt.Sprintf("%v", flowId), "", dataDictionary)
	if err != nil {
		return nil, err
	}
	return resp, nil
}

func getNodeIdByFlowId(appId, flowId, expectedFlowID string) (string, error) {
	var builderJson models.Workflow

	flow, err := db.GetFlowById(cast.ToInt64(flowId), appId)
	if err != nil {
		return "", err
	}

	// if err := json.Unmarshal([]byte(flow.BuilderJson.(string)), &builderJson); err != nil {
	// 	return "", err
	// }
	builderJson = flow.BuilderJson
	nodesData := builderJson.FlowWorkspace.NodesData

	for k, val := range nodesData {
		formIdStruct := val.(map[string]interface{})
		if formIdStruct["flowId"] != nil {
			flowId := formIdStruct["flowId"].(float64)
			if expectedFlowID == fmt.Sprintf("%v", flowId) {
				return k, nil
			}
		}
	}

	return "", nil
}
func resolveCondition(nodeID string, dataDictionary map[string]interface{}, builderJson models.Workflow) (string, string, error) {
	var conditionName string
	basePath := builderJson.FlowWorkspace.NodesData[nodeID].(map[string]interface{})
	byteData, err := json.Marshal(basePath)
	if err != nil {
		return "", "", err
	}
	var conditionodeData models.ConditionNodeData
	err = json.Unmarshal(byteData, &conditionodeData)
	if err != nil {
		return "", "", err
	}

	for _, node := range conditionodeData.Logic.Conditions {
		for _, condition := range node.Rules {
			var operandA, operandB interface{}
			if stringsContainCurlyBraces(condition.OperandA.(string)) {
				operandA, err = getValueFromNestedMap(dataDictionary, condition.OperandA.(string))
				if err != nil {
					return "", "", err
				}
			} else {
				operandA = condition.OperandA
			}
			if stringsContainCurlyBraces(condition.OperandB.(string)) {
				operandB, err = getValueFromNestedMap(dataDictionary, condition.OperandB.(string))
				if err != nil {
					return "", "", err
				}
			} else {
				operandB = condition.OperandB
			}

			operandADataType := reflect.TypeOf(operandA).Kind()
			operandBDataType := reflect.TypeOf(operandB).Kind()
			if operandADataType == operandBDataType {
				switch operandADataType {
				case reflect.Int:
					if compareInt(operandA.(int), operandB.(int), condition.Operator) {
						conditionName = node.Name
					}
				case reflect.Float64:
					if compareFloat64(operandA.(float64), operandB.(float64), condition.Operator) {
						conditionName = node.Name
					}
				case reflect.String:
					if compareString(operandA.(string), operandB.(string), condition.Operator) {
						conditionName = node.Name
					}
				case reflect.Bool:
					if compareBool(operandA.(bool), operandB.(bool), condition.Operator) {
						conditionName = node.Name
					}
				}
			}

		}
	}
	return conditionName, nodeID, nil
}

func getFormOnCondition(key, appId, flowId, nodeID string, builderJson models.Workflow, dataDictionary map[string]interface{}) (map[string]interface{}, error) {
	var currentEdge, targetNode string
	var err error
	var resp map[string]interface{}
	conditionName, currentNodeId, _ := resolveCondition(nodeID, dataDictionary, builderJson)
	if conditionName == "" {
		err := fmt.Errorf("condition not met")
		return nil, err
	}
	pushNode(key, flowId, nodeID, "", "condition")
	edges := builderJson.FlowWorkspace.Edges

	var sourceNodes []string
	for _, edge := range edges {
		if edge.Source == currentNodeId {
			sourceNodes = append(sourceNodes, edge.ID)
		}
	}

	for _, edgeId := range sourceNodes {
		selectedCondition := builderJson.FlowWorkspace.EdgesData[edgeId]["selectedCondition"]
		if selectedCondition == conditionName {
			currentEdge = edgeId
		}
	}

	for _, edgeTarget := range edges {
		if edgeTarget.ID == currentEdge {
			targetNode = edgeTarget.Target
		}
	}

	_, nodeType := GetCurrentNode(targetNode, builderJson)
	if nodeType == "end" {
		ClearDataDictionary(key)
		nodeData, ok := builderJson.FlowWorkspace.NodesData[targetNode]
		if !ok {
			resp = gin.H{"type": "end"}
			return resp, nil
		}
		resp = nodeData.(map[string]interface{})
		resp["type"] = "end"
		return resp, nil
	}
	if nodeType == "form" {
		nodeData := builderJson.FlowWorkspace.NodesData[targetNode]
		formIdStruct := nodeData.(map[string]interface{})

		form, err := db.GetFormById(formIdStruct["formId"].(float64))
		if err != nil {
			return nil, err
		}

		var themeJson string
		theme, err := db.GetThemeByAppIdAndFlowId(appId, flowId, "FLOW")
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				themeJson = ""
			} else {
				return nil, err
			}
		} else {
			themeJsonByte, _ := json.Marshal(theme.Theme)
			themeJson = string(themeJsonByte)
		}
		formJsonByte, _ := json.Marshal(form.BuilderJson)
		resp = gin.H{"dataDictionary": dataDictionary, "form": string(formJsonByte), "nodeId": targetNode, "theme": themeJson}
		pushNode(key, flowId, targetNode, string(formJsonByte), "form")
		return resp, nil
		//c.JSON(http.StatusOK, gin.H{"data": gin.H{"dataDictionary": dataDictionary, "form": form.BuilderJson, "nodeId": targetNode}})
	}
	if nodeType == "api" {
		dataDictionary, err = handleApiNode(key, flowId, targetNode, dataDictionary, builderJson)
		if err != nil {
			return nil, err
		}
	}
	if nodeType == "condition" {
		resp, err = getFormOnCondition(key, appId, flowId, nodeID, builderJson, dataDictionary)
		if err != nil {
			return nil, err
		}

	}
	if nodeType == "subFlow" {
		IncrementFlowLevel(key)
		AddPreviousFlow(key, flowId, targetNode)
		resp, err = handleFlowNode(key, appId, targetNode, builderJson, dataDictionary)
		if err != nil {
			return nil, err
		}
	}
	return resp, nil
}

func compareBool(operandA, operandB bool, operator string) bool {

	switch operator {
	case "==":
		return operandA == operandB
	case "!=":
		return operandA != operandB
	default:
		return false
	}
}

func compareString(operandA, operandB string, operator string) bool {

	switch operator {
	case "==":
		return operandA == operandB
	case "!=":
		return operandA != operandB
	default:
		return false
	}
}

func compareFloat64(operandA, operandB float64, operator string) bool {

	switch operator {
	case "==":
		return operandA == operandB
	case ">":
		return operandA > operandB
	case "<":
		return operandA < operandB
	case "!=":
		return operandA != operandB
	case ">=":
		return operandA >= operandB
	case "<=":
		return operandA <= operandB
	default:
		return false
	}
}

func compareInt(operandA, operandB int, operator string) bool {
	switch operator {
	case "==":
		return operandA == operandB
	case ">":
		return operandA > operandB
	case "<":
		return operandA < operandB
	case "!=":
		return operandA != operandB
	case ">=":
		return operandA >= operandB
	case "<=":
		return operandA <= operandB
	default:
		return false
	}
}

func updateNestedMaps(existingMap, newValueMap map[string]interface{}) {
	for nestedKey, nestedValue := range newValueMap {
		if existingNestedValue, exists := existingMap[nestedKey]; exists {
			// Check if the nested value is a map
			if existingNestedMap, ok := existingNestedValue.(map[string]interface{}); ok {
				// Check if the new nested value is also a map
				if newNestedMap, ok := nestedValue.(map[string]interface{}); ok {
					// Recursively update nested maps
					updateNestedMaps(existingNestedMap, newNestedMap)
				} else {
					// If the new nested value is not a map, update the existing nested value
					existingMap[nestedKey] = nestedValue
				}
			} else {
				// If the existing nested value is not a map, update it directly
				existingMap[nestedKey] = nestedValue
			}
		}
	}
}

func pushNode(key, flowId, nodeId, form, nodeType string) {
	formHistory, exists := FormHistoryMap[key]
	if !exists {
		formHistory = stack.Stack[FormHistory]{}
	}

	formHistory.Push(FormHistory{
		NodeId:   nodeId,
		FlowId:   flowId,
		Form:     form,
		NodeType: nodeType,
	})
	FormHistoryMap[key] = formHistory
}

func getValueFromNestedMap(data map[string]interface{}, key string) (interface{}, error) {
	key = strings.Trim(key, "{}")
	keysList := strings.Split(key, ".")
	currentMap := data

	for _, key := range keysList {
		value, exists := currentMap[key]
		if !exists {
			return nil, fmt.Errorf("key not found: %s", key)
		}

		// Check if the value is a nested map
		if nestedMap, ok := value.(map[string]interface{}); ok {
			// Update the current map for the next iteration
			currentMap = nestedMap
		} else {
			// If the value is not a map, return the value
			return value, nil
		}
	}

	return nil, fmt.Errorf("invalid keys format: %s", key)
}

func stringsContainCurlyBraces(input string) bool {
	return strings.Contains(input, "{") && strings.Contains(input, "}")
}

func getPlaceholders(url string) []string {
	// Use a regular expression to find all occurrences of the pattern "{<some text>}"
	regexPattern := `\{([a-zA-Z0-9_.]+)\}`
	re := regexp.MustCompile(regexPattern)

	// Find all matches in the URL
	matches := re.FindAllString(url, -1)

	// Extract placeholder names from the matches
	var placeholders []string

	placeholders = append(placeholders, matches...)

	return placeholders
}

// replacePlaceholder replaces a specific placeholder in a URL with its value
func replacePlaceholder(url, placeholder, value string) string {
	return strings.ReplaceAll(url, placeholder, value)
}

func IncrementFlowLevel(key string) {
	level := CurrentFlow[key]["flowLevel"]
	CurrentFlow[key]["flowLevel"] = level.(int) + 1
}
func DecrementFlowLevel(key string) {
	level := CurrentFlow[key]["flowLevel"]
	CurrentFlow[key]["flowLevel"] = level.(int) - 1
}
func getFlowLevel(key string) int {
	level := CurrentFlow[key]["flowLevel"]
	return level.(int)
}

func AddPreviousFlow(key, flowId, nodeId string) {
	flows := CurrentFlow[key]["previousFlows"]
	flowsstack := flows.(stack.Stack[PreviousFlow])
	flowsstack.Push(PreviousFlow{NodeId: nodeId, FlowId: flowId})
	CurrentFlow[key]["previousFlows"] = flowsstack
}

func GetPreviousFlow(key string) PreviousFlow {
	flows := CurrentFlow[key]["previousFlows"]
	flowsstack := flows.(stack.Stack[PreviousFlow])
	f := flowsstack.Pop()
	CurrentFlow[key]["previousFlows"] = flowsstack
	return f
}

func GetMainFlowNode(key, mainFlowId string) string {
	flows := CurrentFlow[key]["previousFlows"]
	flowsstack := flows.(stack.Stack[PreviousFlow])
	var f PreviousFlow
	for f.FlowId != mainFlowId {
		f = flowsstack.Pop()
	}
	return f.NodeId
}

func PeekPreviousFlow(key string) (PreviousFlow, error) {
	flows := CurrentFlow[key]["previousFlows"]
	flowsstack := flows.(stack.Stack[PreviousFlow])
	if flowsstack.Len() > 0 {
		return flowsstack.Top(), nil
	}
	return PreviousFlow{}, fmt.Errorf("no previous flow")
}

type PreviousFlow struct {
	NodeId string
	FlowId string
}

type FormHistory struct {
	NodeId   string
	FlowId   string
	Form     string
	NodeType string
}

func ValidateAppToken(token string) (uint, uint, error) {

	appToken, err := db.GetAppTokenByToken(util.EncodeStringToBase64(token))
	if err != nil {
		return 0, 0, err
	}
	return appToken.AppId, appToken.FlowId, nil
}
