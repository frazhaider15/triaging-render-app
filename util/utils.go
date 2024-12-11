package util

import (
	"bytes"
	"encoding/base64"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
)

func ReplaceAllInString(originalStr string, toBereplaced string, replaceWith string) string {
	re := regexp.MustCompile(toBereplaced)
	originalStr = re.ReplaceAllString(originalStr, replaceWith)
	return originalStr
}

func IsValidName(str string) bool {
	regex := `^[\w\-_]+$`
	isValid, _ := regexp.MatchString(regex, str)
	return isValid
}

func MakeHttpRequestWithRawJsonBody(method, url, jsonPayload string, headers map[string]string) ([]byte, int, error) {

	// Create a new HTTP request with the JSON payload
	req, err := http.NewRequest(method, url, bytes.NewBuffer([]byte(jsonPayload)))
	if err != nil {
		return nil, 500, err
	}

	// Set the headers
	for key, val := range headers {
		req.Header.Add(key, val)
	}

	// Create an HTTP client and send the request
	client := &http.Client{}
	resp, err := client.Do(req)
	if err != nil {
		return nil, 200, err
	}
	defer resp.Body.Close()

	// Check the response status and handle it as needed
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(resp.Body)
		return nil, resp.StatusCode, fmt.Errorf("request failed with status code: %d\n err: %v", resp.StatusCode, string(body))
		// Handle the error response here
		// You can read the response body using ioutil.ReadAll(resp.Body)
	}
	// Handle the successful response here
	// Read the response body
	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, resp.StatusCode, err
	}
	return body, resp.StatusCode, err
}

func DecodeBase64String(s string) (string, error) {
	data, err := base64.StdEncoding.DecodeString(s)
	if err != nil {
		return "", err
	}
	return string(data), nil
}

func EncodeStringToBase64(s string) string {
	return base64.StdEncoding.EncodeToString([]byte(s))
}

func MaskToken(token string, visiblePrefix int, visibleSuffix int) string {
	tokenLength := len(token)

	if visiblePrefix+visibleSuffix > tokenLength {
		return strings.Repeat("*", tokenLength)
	}

	masked := token[:visiblePrefix] + strings.Repeat("*", tokenLength-visiblePrefix-visibleSuffix) + token[tokenLength-visibleSuffix:]
	return masked
}

func RemoveIntDuplicates(arr []uint) []uint {
  seen := make(map[uint]bool)
  var result []uint
  for _, num := range arr {
    if !seen[num] {
      seen[num] = true
      result = append(result, num)
    }
  }
  return result
}

func CapitalizeFirstLetter(str string) string {
    if len(str) == 0 {
        return str
    }
    return strings.ToUpper(str[:1]) + strings.ToLower(str[1:])
}
