package download

import (
	"errors"
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
