package hal

import (
	"fmt"
	"io"
	"testing"

	"github.com/stretchr/testify/assert"
)

func init() {
	PARAMS.LoadParams("params.json")
}

func TestTextToSpeechStream(t *testing.T) {
	ss, err := NewSpeechSynthesisStream(PARAMS.SpeechKey, PARAMS.SpeechRegion, "en-US-ElizabethNeural")
	assert.Nil(t, err)

	sendText("Does anyone hearing me?", ss, t)
	assert.Nil(t, ss.Error())
	sendText("yes, there is one man.", ss, t)
	assert.Nil(t, ss.Error())

	ss.Close()
}

func TestAutoDetectedTextToSpeechStream(t *testing.T) {
	ss, err := NewAutoDetectedSpeechSynthesisStream(PARAMS.SpeechKey, PARAMS.SpeechRegion)
	assert.Nil(t, err)

	sendText("Hello, text to speech.", ss, t)
	assert.Nil(t, ss.Error())
	sendText("你好，文本转语音.", ss, t)
	assert.Nil(t, ss.Error())

	ss.Close()
}

func TestTextToSpeechStandalone(t *testing.T) {
	ss, err := NewSpeechSynthesisStandalone(PARAMS.SpeechKey, PARAMS.SpeechRegion, "en-US-ElizabethNeural")
	assert.Nil(t, err)

	defer ss.Close()
	ss.TextToSpeech("Does anyone hearing me?")
	text, _, err := ss.Result()
	for ; err == nil; text, _, err = ss.Result() {
		fmt.Println("Text: ", text.Text)
	}
	assert.ErrorIs(t, err, io.EOF)
	assert.Nil(t, ss.Error())
}

func TestAutoDetectedTextToSpeechStandalone(t *testing.T) {
	ss, err := NewAutoDetectedSpeechSynthesisStandalone(PARAMS.SpeechKey, PARAMS.SpeechRegion)
	assert.Nil(t, err)

	defer ss.Close()
	ss.TextToSpeech("Does anyone hearing me?")
	text, _, err := ss.Result()
	for ; err == nil; text, _, err = ss.Result() {
		fmt.Println("Text: ", text.Text)
	}
	assert.ErrorIs(t, err, io.EOF)
	assert.Nil(t, ss.Error())

	ss.TextToSpeech("你能听见我说吗？")
	text, _, err = ss.Result()
	for ; err == nil; text, _, err = ss.Result() {
		fmt.Println("Text: ", text.Text)
	}
	assert.ErrorIs(t, err, io.EOF)
	assert.Nil(t, ss.Error())
}

func TestGetSupportLanguageAndVoices(t *testing.T) {
	ss, err := NewSpeechSynthesisStandalone(PARAMS.SpeechKey, PARAMS.SpeechRegion, "en-US-ElizabethNeural")
	assert.Nil(t, err)

	defer ss.Close()
	all := ss.GetAllSupportLanguage()
	assert.NotNil(t, all)

	for _, l := range all {
		fmt.Println(l)
	}

	voices := ss.GetAllSupportVoicesForLanguage("zh-CN")
	for _, v := range voices {
		fmt.Println(v)
	}
}

func sendText(text string, ss *SpeechSynthesisStream, t *testing.T) {
	err := ss.TextToSpeech(text)
	assert.Nil(t, err)

	word, audio, err := ss.Result()
	for ; err == nil; word, audio, err = ss.Result() {
		if word != nil {
			fmt.Println("Text: ", word.Text)
		}

		if audio != nil {
			fmt.Println("Audio: ", len(audio))
		}
	}

	assert.ErrorIs(t, err, io.EOF)
}
