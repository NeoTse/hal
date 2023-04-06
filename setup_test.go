package hal

import (
	"fmt"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	PARAMS.LoadParams("params.json")
}

func TestGetOpenaiKey(t *testing.T) {
	// use the key from params.json
	input = strings.NewReader("")
	PARAMS.OpenaiKey = getOpenaiKey()
	assert.NotEmpty(t, PARAMS.OpenaiKey)

	// use the key from input
	temp := PARAMS.OpenaiKey
	PARAMS.OpenaiKey = ""
	input = strings.NewReader(temp)
	PARAMS.OpenaiKey = getOpenaiKey()
	assert.Equal(t, temp, PARAMS.OpenaiKey)
}

func TestChooseModel(t *testing.T) {
	input = strings.NewReader("")
	PARAMS.ChatgptModel = chooseModel()
	assert.NotEmpty(t, PARAMS.ChatgptModel)

	PARAMS.ChatgptModel = ""
	input = strings.NewReader("3")
	PARAMS.ChatgptModel = chooseModel()
	assert.Equal(t, "gpt-4", PARAMS.ChatgptModel)
}

func TestGetSpeechKey(t *testing.T) {
	input = strings.NewReader("")
	PARAMS.SpeechKey = getSpeechKey()
	assert.NotEmpty(t, PARAMS.SpeechKey)

	temp := PARAMS.SpeechKey
	PARAMS.SpeechKey = ""
	input = strings.NewReader(temp)
	PARAMS.SpeechKey = getSpeechKey()
	assert.Equal(t, temp, PARAMS.SpeechKey)
}

func TestGetSpeechRegion(t *testing.T) {
	input = strings.NewReader("")
	PARAMS.SpeechRegion = getSpeechRegion()
	assert.NotEmpty(t, PARAMS.SpeechRegion)

	temp := PARAMS.SpeechRegion
	PARAMS.SpeechRegion = ""
	input = strings.NewReader(temp)
	PARAMS.SpeechRegion = getSpeechRegion()
	assert.Equal(t, temp, PARAMS.SpeechRegion)
}

func TestCheckSpeechKeyAndRegion(t *testing.T) {
	fmt.Printf("key: %s, Region: %s\n", PARAMS.SpeechKey, PARAMS.SpeechRegion)
	assert.True(t, checkSpeechKeyAndRegion(PARAMS.SpeechKey, PARAMS.SpeechRegion))
}

func TestChooseLanguage(t *testing.T) {
	input = strings.NewReader("")
	PARAMS.Language = chooseLanguage()
	assert.NotEmpty(t, PARAMS.Language)

	input = strings.NewReader("1")
	PARAMS.Language = chooseLanguage()
	assert.Empty(t, PARAMS.Language)
}

func TestSelectLanguage(t *testing.T) {
	input = strings.NewReader("en-US")
	PARAMS.Language = selectLanguage()
	assert.NotEmpty(t, PARAMS.Language)
}

func TestSelectVoiceLanguage(t *testing.T) {
	input = strings.NewReader("en-US")
	assert.Equal(t, "en-US", selectVoiceLanguage())
}

func TestSelectVoice(t *testing.T) {
	input = strings.NewReader("1")
	PARAMS.Voice = selectVoice("en-US")
	assert.NotEmpty(t, PARAMS.Voice)
}
