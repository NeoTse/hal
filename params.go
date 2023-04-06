package hal

import (
	"bufio"
	"encoding/json"
	"os"
)

type Params struct {
	Initialized  bool
	OpenaiKey    string
	ChatgptModel string
	MaxHistory   int

	SpeechKey       string
	SpeechRegion    string
	Language        string // BCP-47 code
	Voice           string
	StopWord        string
	Keyword         string
	KeywordModel    string
	KeywordLanguage string
}

func (p Params) String() string {
	json, err := json.MarshalIndent(p, "", " ")
	if err != nil {
		return ""
	}

	return string(json)
}

var PARAMS = Params{MaxHistory: 4, Language: "en-US", Voice: "en-US-ElizabethNeural"}

func (p Params) SaveParams(file string) error {
	json, err := json.MarshalIndent(p, "", " ")
	if err != nil {
		return err
	}

	f, err := os.Create(file)
	if err != nil {
		return err
	}

	defer f.Close()
	writer := bufio.NewWriter(f)
	_, err = writer.Write(json)
	if err != nil {
		return err
	}

	tlog.Debugf("save params.")

	return writer.Flush()
}

func (p *Params) LoadParams(file string) error {
	f, err := os.Open(file)
	if err != nil {
		return err
	}

	defer f.Close()
	reader := bufio.NewReader(f)
	var content []byte
	for res, _ := reader.ReadBytes('\n'); len(res) != 0; res, _ = reader.ReadBytes('\n') {
		content = append(content, res...)
	}

	tlog.Debugf("load params.")

	return json.Unmarshal(content, p)
}

// because azure stt go speech sdk only support 4 languages in auto detected by languages (not support full languages auto detected)
var AUTO_DETECTED_LANGUAGE = map[string]string{
	// "German":     {"de-AT", "de-CH", "de-DE"},
	// "English":    {"en-AU", "en-CA", "en-GB", "en-HK", "en-IE", "en-IN", "en-KE", "en-NG", "en-NZ", "en-PH", "en-SG", "en-TZ", "en-US", "en-ZA"},
	// "Spanish":    {"es-AR", "es-BO", "es-CL", "es-CO", "es-CR", "es-CU", "es-DO", "es-EC", "es-ES", "es-GQ", "es-GT", "es-HN", "es-MX", "es-NI", "es-PA", "es-PE", "es-PR", "es-PY", "es-SV", "es-US", "es-UY", "es-VE"},
	// "French":     {"fr-BE", "fr-CA", "fr-CH", "fr-FR"},
	// "Italian":    {"it-IT", "it-CH"},
	// "Japanese":   {"ja-JP"},
	// "Korean":     {"ko-KR"},
	// "Portuguese": {"pt-BR", "pt-PT"},
	// "Russian":    {"ru-RU"},
	// "Ukrainian":  {"uk-UA"},
	// "Chinese":    {"wuu-CN", "yue-CN", "zh-CN", "zh-CN-sichuan", "zh-HK", "zh-TW"},
	// "Arabic":     {"ar-AE", "ar-BH", "ar-DZ", "ar-EG", "ar-IQ", "ar-JO", "ar-KW", "ar-LB", "ar-LY", "ar-MA", "ar-OM", "ar-QA", "ar-SA", "ar-SY", "ar-TN", "ar-YE"},
	// "Hindi":      {"hi-IN"},

	"English": "en-US",
	"French":  "fr-FR",
	"Chinese": "zh-CN",
	"Spanish": "es-ES",
}

func GetAutoDetectedLanguages() []string {
	var res []string
	for _, v := range AUTO_DETECTED_LANGUAGE {
		res = append(res, v)
	}

	return res
}

func GetAutoDetectedLanguagesName() []string {
	var res []string
	for k := range AUTO_DETECTED_LANGUAGE {
		res = append(res, k)
	}

	return res
}
