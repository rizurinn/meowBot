package api

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"strings"
)

const (
	stickerApiBaseUrl = "https://sticker-api.openwa.dev/"
)

func ConvertViaApi(mediaData []byte, mediaType, packName, authorName string) ([]byte, error) {
	var path, requestBody string

	metadata := fmt.Sprintf(`"stickerMetadata": { "author": "%s", "pack": "%s", "keepScale": true, "removebg": "false", "circle": false },`, authorName, packName)
	
	if mediaType == "image" {
		path = "prepareWebp"
		base64Image := base64.StdEncoding.EncodeToString(mediaData)
		imageBody := `"image": "data:` + http.DetectContentType(mediaData) + `;base64,` + base64Image + `"`
		requestBody = `{` + metadata + imageBody + `}`
	} else if mediaType == "video" {
		path = "convertMp4BufferToWebpDataUrl"
		base64Video := base64.StdEncoding.EncodeToString(mediaData)
		videoBody := `"file": "data:` + http.DetectContentType(mediaData) + `;base64,` + base64Video + `"`
		requestBody = `{` + metadata + videoBody + `}`
	} else {
		return nil, errors.New("tipe media tidak didukung")
	}

	client := &http.Client{}
	req, err := http.NewRequest("POST", stickerApiBaseUrl+path, bytes.NewBuffer([]byte(requestBody)))
	if err != nil {
		return nil, fmt.Errorf("gagal membuat request: %v", err)
	}
	req.Header.Set("Content-Type", "application/json;charset=utf-8")

	resp, err := client.Do(req)
	if err != nil {
		return nil, fmt.Errorf("gagal melakukan request ke API stiker: %v", err)
	}
	defer resp.Body.Close()

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, fmt.Errorf("gagal membaca respons body: %v", err)
	}
	
	if resp.StatusCode != 200 {
		return nil, fmt.Errorf("API stiker merespons dengan error %d: %s", resp.StatusCode, string(body))
	}

	var responseData string
	if mediaType == "image" {
		var result map[string]interface{}
		if err := json.Unmarshal(body, &result); err != nil {
			return nil, fmt.Errorf("gagal unmarshal respons JSON: %v", err)
		}
		if val, ok := result["webpBase64"].(string); ok && val != "" {
			responseData = val
		} else {
			return nil, fmt.Errorf("respons API tidak berisi 'webpBase64'")
		}
	} else {
		parts := strings.Split(string(body), ";base64,")
		if len(parts) < 2 {
			return nil, fmt.Errorf("respons API video tidak valid: %s", string(body))
		}
		responseData = parts[1]
		responseData = strings.Trim(responseData, `"`)
	}

	stickerBytes, err := base64.StdEncoding.DecodeString(responseData)
	if err != nil {
		return nil, fmt.Errorf("gagal decode hasil base64: %v", err)
	}

	return stickerBytes, nil
}
