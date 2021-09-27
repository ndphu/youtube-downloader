package download

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"youtube-downloader/decipher"
)

const PlayerUrl = "https://www.youtube.com/youtubei/v1/player?key=AIzaSyAO_FJ2SlqU8Q4STEHLGCilw_Y9_11qcW8"
const NextUrl = "https://www.youtube.com/youtubei/v1/next?key=AIzaSyAO_FJ2SlqU8Q4STEHLGCilw_Y9_11qcW8"

func NewInfoRequest(id string) *M {
	signTimestamp, _ := decipher.GetSignatureTimestamp(id)
	return &M{
		"context": M{
			"client": M{
				"hl":            "en",
				"gl":            "US",
				"clientName":    "WEB",
				"clientVersion": "2.20210909.07.00",
			},
		},
		"playbackContext": M{
			"contentPlaybackContext": M{
				"signatureTimestamp": signTimestamp,
			},
		},
		"videoId": id,
	}
}

func GetDownloadInfo(id string) (*InfoResponse, error) {
	infoRequest := NewInfoRequest(id)
	postBody, err := infoRequest.ToBytesReader()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", PlayerUrl, postBody)
	if err != nil {
		return nil, err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/93.0.4577.63 Safari/537.36,gzip(gfe)")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return nil, err
	}

	var info InfoResponse
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, err
	}

	return &info, nil
}

func GetAdaptiveFormat(body []byte) ([]Format, error) {
	var info InfoResponse
	if err := json.Unmarshal(body, &info); err != nil {
		return nil, err
	}
	return info.StreamingData.AdaptiveFormats, nil
}
