package download

import (
	"bytes"
	"encoding/json"
)

type InfoResponse struct {
	StreamingData StreamingData `json:"streamingData"`
	VideoDetails  VideoDetails  `json:"videoDetails"`
}

type StreamingData struct {
	Formats         []Format `json:"formats"`
	AdaptiveFormats []Format `json:"adaptiveFormats"`
}

type Format struct {
	Itag             int    `json:"itag"`
	Url              string `json:"url"`
	MimeType         string `json:"mimeType"`
	Bitrate          int    `json:"bitrate"`
	Width            int    `json:"width"`
	Height           int    `json:"height"`
	ContentLength    string `json:"contentLength"`
	Quality          string `json:"quality"`
	Fps              int    `json:"fps"`
	QualityLabel     string `json:"qualityLabel"`
	ProjectionType   string `json:"projectionType"`
	AverageBitrate   int    `json:"averageBitrate"`
	AudioQuality     string `json:"audioQuality"`
	ApproxDurationMs string `json:"approxDurationMs"`
	AudioSampleRate  string `json:"audioSampleRate"`
	AudioChannels    int    `json:"audioChannels"`
	SignatureCipher  string `json:"signatureCipher"`
}

type M map[string]interface{}

func (m *M) ToBytesReader() (*bytes.Reader, error) {
	marshal, err := json.Marshal(m)
	if err != nil {
		return nil, err
	}
	return bytes.NewReader(marshal), nil
}

type DownloadSource struct {
	Video string `json:"video"`
	Audio string `json:"audio"`
}

type VideoDetails struct {
	VideoId          string    `json:"videoId"`
	Title            string    `json:"title"`
	LengthSeconds    string    `json:"lengthSeconds"`
	ChannelId        string    `json:"channelId"`
	ShortDescription string    `json:"shortDescription"`
	IsCrawlable      bool      `json:"isCrawlable"`
	Thumbnail        Thumbnail `json:"thumbnail"`
	AverageRating    float64   `json:"averageRating"`
	ViewCount        string    `json:"viewCount"`
	Author           string    `json:"author"`
	IsPrivate        bool      `json:"isPrivate"`
	IsLiveContent    bool      `json:"isLiveContent"`
}

type Thumbnail struct {
	Thumbnails []YoutubeImage `json:"thumbnails"`
}

type YoutubeImage struct {
	Url    string `json:"url"`
	Width  int    `json:"width"`
	Height int    `json:"height"`
}
