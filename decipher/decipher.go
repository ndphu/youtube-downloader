package decipher

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"net/url"
	"regexp"
	"strconv"
)

var basejsPattern = regexp.MustCompile(`(/s/player/\w+/player_ias.vflset/\w+/base.js)`)

func Decipher(videoId string, cipher string) (string, error) {
	operations, err := parseDecipherOps(videoId)
	if err != nil {
		return "", err
	}
	bs := []byte(cipher)
	for _, op := range operations {
		bs = op(bs)
	}
	return string(bs), err
}

func DecipherURL(videoID string, cipher string) (string, error) {
	queryParams, err := url.ParseQuery(cipher)
	if err != nil {
		return "", err
	}

	operations, err := parseDecipherOps(videoID)
	if err != nil {
		return "", err
	}

	// apply operations
	bs := []byte(queryParams.Get("s"))
	for _, op := range operations {
		bs = op(bs)
	}

	fmt.Println("deciphered sig", string(bs))

	decipheredURL := fmt.Sprintf("%s&%s=%s", queryParams.Get("url"), queryParams.Get("sp"), string(bs))
	return decipheredURL, nil
}

const (
	jsvarStr   = "[a-zA-Z_\\$][a-zA-Z_0-9]*"
	reverseStr = ":function\\(a\\)\\{" +
		"(?:return )?a\\.reverse\\(\\)" +
		"\\}"
	spliceStr = ":function\\(a,b\\)\\{" +
		"a\\.splice\\(0,b\\)" +
		"\\}"
	swapStr = ":function\\(a,b\\)\\{" +
		"var c=a\\[0\\];a\\[0\\]=a\\[b(?:%a\\.length)?\\];a\\[b(?:%a\\.length)?\\]=c(?:;return a)?" +
		"\\}"
)

var (
	actionsObjRegexp = regexp.MustCompile(fmt.Sprintf(
		"var (%s)=\\{((?:(?:%s%s|%s%s|%s%s),?\\n?)+)\\};", jsvarStr, jsvarStr, swapStr, jsvarStr, spliceStr, jsvarStr, reverseStr))

	actionsFuncRegexp = regexp.MustCompile(fmt.Sprintf(
		"function(?: %s)?\\(a\\)\\{"+
			"a=a\\.split\\(\"\"\\);\\s*"+
			"((?:(?:a=)?%s\\.%s\\(a,\\d+\\);)+)"+
			"return a\\.join\\(\"\"\\)"+
			"\\}", jsvarStr, jsvarStr, jsvarStr))

	reverseRegexp = regexp.MustCompile(fmt.Sprintf("(?m)(?:^|,)(%s)%s", jsvarStr, reverseStr))
	spliceRegexp  = regexp.MustCompile(fmt.Sprintf("(?m)(?:^|,)(%s)%s", jsvarStr, spliceStr))
	swapRegexp    = regexp.MustCompile(fmt.Sprintf("(?m)(?:^|,)(%s)%s", jsvarStr, swapStr))
)

func parseDecipherOps(videoId string) ([]DecipherOperation, error) {
	actions, err := getDecipherActionsFromCache()
	if err != nil {
		log.Println("Fail to get decipher actions from cache")
		actions, err = parseDecipherActions(videoId)
		if err != nil {
			log.Println("Fail to get decipher actions from Youtube")
			return nil, err
		}
		if err := saveDecipherActionsToCache(actions); err != nil {
			log.Println("Fail to save decipher ops to cache")
			return nil, err
		}
	}

	var ops []DecipherOperation
	for _, action := range actions {
		switch action.Action {
		case "reverse":
			{
				ops = append(ops, reverseFunc)
			}
		case "swap":
			{
				ops = append(ops, newSwapFunc(action.Arg))
			}
		case "splice":
			{
				ops = append(ops, newSpliceFunc(action.Arg))
			}
		}
	}
	return ops, nil
}

func getDecipherActionsFromCache() ([]DecipherAction, error) {
	file, err := ioutil.ReadFile(".cache.decipherOps")
	if err != nil {
		return nil, err
	}
	var actions []DecipherAction
	err = json.Unmarshal(file, &actions)
	return actions, err
}

func saveDecipherActionsToCache(actions []DecipherAction) error {
	marshal, err := json.Marshal(actions)
	if err != nil {
		return err
	}
	return ioutil.WriteFile(".cache.decipherOps", marshal, 0644)
}

func parseDecipherActions(videoID string) ([]DecipherAction, error) {
	basejsBody, err := fetchPlayerConfig(videoID)
	if err != nil {
		return nil, err
	}

	objResult := actionsObjRegexp.FindSubmatch(basejsBody)
	funcResult := actionsFuncRegexp.FindSubmatch(basejsBody)
	if len(objResult) < 3 || len(funcResult) < 2 {
		return nil, fmt.Errorf("error parsing signature tokens (#obj=%d, #func=%d)", len(objResult), len(funcResult))
	}

	obj := objResult[1]
	objBody := objResult[2]
	funcBody := funcResult[1]

	var reverseKey, spliceKey, swapKey string

	if result := reverseRegexp.FindSubmatch(objBody); len(result) > 1 {
		reverseKey = string(result[1])
	}
	if result := spliceRegexp.FindSubmatch(objBody); len(result) > 1 {
		spliceKey = string(result[1])
	}
	if result := swapRegexp.FindSubmatch(objBody); len(result) > 1 {
		swapKey = string(result[1])
	}

	regex, err := regexp.Compile(fmt.Sprintf("(?:a=)?%s\\.(%s|%s|%s)\\(a,(\\d+)\\)", obj, reverseKey, spliceKey, swapKey))
	if err != nil {
		return nil, err
	}

	actions := make([]DecipherAction, 0)
	for _, s := range regex.FindAllSubmatch(funcBody, -1) {
		switch string(s[1]) {
		case reverseKey:
			actions = append(actions, DecipherAction{
				Action: "reverse",
				Arg:    0,
			})
		case swapKey:
			arg, _ := strconv.Atoi(string(s[2]))
			actions = append(actions, DecipherAction{
				Action: "swap",
				Arg:    arg,
			})
		case spliceKey:
			arg, _ := strconv.Atoi(string(s[2]))
			actions = append(actions, DecipherAction{
				Action: "splice",
				Arg:    arg,
			})
		}
	}
	return actions, nil
}

type DecipherAction struct {
	Action string `json:"action"`
	Arg    int    `json:"arg"`
}

func fetchPlayerConfig(videoID string) ([]byte, error) {
	embedURL := fmt.Sprintf("https://youtube.com/embed/%s?hl=en", videoID)
	embedBody, err := httpGetBodyBytes(embedURL)
	if err != nil {
		return nil, err
	}

	escapedBasejsURL := string(basejsPattern.Find(embedBody))
	if escapedBasejsURL == "" {
		log.Println("playerConfig:", string(embedBody))
		return nil, errors.New("unable to find basejs URL in playerConfig")
	}

	return httpGetBodyBytes("https://youtube.com" + escapedBasejsURL)
}

func httpGetBodyBytes(url string) ([]byte, error) {
	resp, err := http.Get(url)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	return ioutil.ReadAll(resp.Body)
}

func GetSignatureTimestamp(videoId string) (string, error) {
	cached, err := getSignatureFromCache()
	if err != nil {
		newSignature, err := getFromYoutube(videoId)
		if err != nil {
			return "", err
		}
		if err := saveSignatureToCache(newSignature); err != nil {
			log.Println("Fail to save signature to cache by error", err.Error())
		}
		return newSignature, nil
	}
	return cached, nil
}

func getSignatureFromCache() (string, error) {
	file, err := ioutil.ReadFile(".cache.signature")
	if err != nil {
		return "", err
	}
	return string(file), nil
}
func saveSignatureToCache(signature string) error {
	return ioutil.WriteFile(".cache.signature", []byte(signature), 0644)
}

func getFromYoutube(videoId string) (string, error) {
	basejsBody, err := fetchPlayerConfig(videoId)
	if err != nil {
		return "", err
	}

	result := signatureRegexp.FindSubmatch(basejsBody)
	if result == nil {
		return "", errors.New("SignatureNotFound")
	}

	return string(result[1]), nil
}

var signatureRegexp = regexp.MustCompile(`(?m)(?:^|,)(?:signatureTimestamp:)(\d+)`)
