package download

import (
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"strings"
)

type Video struct {
	Id              string
	Formats         []Format
	AdaptiveFormats []Format
	Info            VideoDetails
}

func NewVideo(id string) *Video {
	return &Video{
		Id: id,
	}
}

func (v *Video) Load() error {
	info, err := GetDownloadInfo(v.Id)
	if err != nil {
		return err
	}
	v.Info = info.VideoDetails
	v.Formats = info.StreamingData.Formats
	v.AdaptiveFormats = info.StreamingData.AdaptiveFormats
	return nil
}

func (v *Video) GetStream(mimeType string) []Format {
	formats := make([]Format, 0)
	for _, format := range v.AdaptiveFormats {
		if strings.Contains(format.MimeType, mimeType) {
			formats = append(formats, format)
		}
	}
	return formats
}

func (v *Video) GetAudioStream() (string, error) {
	formats := v.GetStream("audio/mp4")
	for _, format := range formats {
		if format.AudioQuality == "AUDIO_QUALITY_MEDIUM" {
			return GetLink(v.Id, format)
		}
	}
	return "", errors.New("NoAudioStream")
}

func (v *Video) GetVideoStream() (string, error) {
	formats := v.GetStream("video/mp4")
	maxWidth := 0
	var highestQuality Format
	for _, format := range formats {
		if format.Width > maxWidth {
			highestQuality = format
		}
	}
	if highestQuality.Width == 0 {
		return "", errors.New("NoVideoStream")
	}
	return GetLink(v.Id, highestQuality)
}

func (v *Video) NextList() ([]VideoItem, error) {
	request := NewInfoRequest(v.Id)
	postBody, err := request.ToBytesReader()
	if err != nil {
		return nil, err
	}

	req, err := http.NewRequest("POST", NextUrl, postBody)
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

	var next M
	if err := json.Unmarshal(body, &next); err != nil {
		return nil, err
	}
	results := next.GetObject("contents").GetObject("twoColumnWatchNextResults").GetObject("secondaryResults").GetObject("secondaryResults").GetArray("results")

	videos := make([]VideoItem, 0)
	for _, res := range results {
		renderer := res.GetObject("compactVideoRenderer")
		if renderer == nil {
			continue
		}
		videos = append(videos, VideoItem{
			Id:    renderer.GetString("videoId"),
			Title: renderer.GetObject("title").GetString("simpleText"),
		})
	}
	return videos, nil
}

func (v *Video) Next() (string, error) {
	request := NewInfoRequest(v.Id)
	postBody, err := request.ToBytesReader()
	if err != nil {
		return "", err
	}

	req, err := http.NewRequest("POST", NextUrl, postBody)
	if err != nil {
		return "", err
	}
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("User-Agent", "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/93.0.4577.63 Safari/537.36,gzip(gfe)")
	resp, err := http.DefaultClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return "", err
	}

	var next M
	if err := json.Unmarshal(body, &next); err != nil {
		return "", err
	}
	return next.GetObject("responseContext").GetObject("webResponseContextExtensionData").GetObject("webPrefetchData").
		GetArray("navigationEndpoints")[0].GetObject("watchEndpoint").GetString("videoId"), nil
}
