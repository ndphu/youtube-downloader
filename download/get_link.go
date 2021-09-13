package download

import (
	"log"
	"net/url"
	"youtube-downloader/decipher"
)

func GetLink(id string, format Format) (string, error) {
	if format.Url == "" {
		log.Println("Decoding stream link...")
		query, err := url.ParseQuery(format.SignatureCipher)
		if err != nil {
			return "", err
		}
		sig, err := decipher.Decipher(id, query.Get("s"))
		if err != nil {
			return "", err
		}
		return query.Get("url") + "&sig=" + sig, nil
	} else {
		return format.Url, nil
	}
}
