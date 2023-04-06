package hal

import (
	"fmt"
	"io"
	"strings"
	"testing"

	"github.com/Microsoft/cognitive-services-speech-sdk-go/speech"
	"github.com/stretchr/testify/assert"
)

func init() {
	PARAMS.LoadParams("params.json")
	tlog.level = DEBUG
	SegmentationSilenceTimeoutMs = "800"
}

func TestSpeechRecognitionStream(t *testing.T) {
	sr, err := NewSpeechRecognitionStream(PARAMS.SpeechKey, PARAMS.SpeechRegion, []string{"en-US"})
	assert.Nil(t, err)
	err = sr.Start()
	assert.Nil(t, err)

	sendAudio("./test_data/jfk.wav", sr, t)
	// sendAudio("./test_data/audio2.wav", sr, t)

	err = sr.Close()
	assert.Nil(t, err)
}

func TestAutoDetectedSpeechRecognitionStream(t *testing.T) {
	sr, err := NewAutoDetectedSpeechRecognitionStream(PARAMS.SpeechKey, PARAMS.SpeechRegion)
	assert.Nil(t, err)
	err = sr.Start()
	assert.Nil(t, err)

	sendAudio("./test_data/jfk.wav", sr, t)
	// sendAudio("./test_data/audio2.wav", sr, t)

	err = sr.Close()
	assert.Nil(t, err)
}

func TestSpeechRecognitionStandalone(t *testing.T) {
	sr, err := NewSpeechRecognitionStandalone(PARAMS.SpeechKey, PARAMS.SpeechRegion, []string{"en-US"})
	assert.Nil(t, err)
	fmt.Println("say something")
	err = sr.Start()
	assert.Nil(t, err)

	r, err := sr.Result()
	fmt.Println(r)
	assert.Nil(t, err)

	err = sr.Close()
	assert.Nil(t, err)
}

func TestAutoDetectedSpeechRecognitionStandalone(t *testing.T) {
	sr, err := NewAutoDetectedSpeechRecognitionStandalone(PARAMS.SpeechKey, PARAMS.SpeechRegion)
	assert.Nil(t, err)
	err = sr.Start()
	assert.Nil(t, err)
	fmt.Println("started. say something")

	r, err := sr.Result()
	fmt.Println(r)
	assert.Nil(t, err)

	err = sr.Close()
	assert.Nil(t, err)
}

func TestKeywordRecognitionStanalone(t *testing.T) {
	sk, err := NewKeywordRecognitionStandalone(PARAMS.SpeechKey, PARAMS.SpeechRegion, []string{PARAMS.KeywordLanguage}, PARAMS.KeywordModel, PARAMS.Keyword)
	assert.Nil(t, err)
	fmt.Printf("say %s\n", PARAMS.Keyword)
	err = sk.Start()
	assert.Nil(t, err)

	r, err := sk.Result()
	assert.Nil(t, err)
	fmt.Printf("%s here\n", r)

	assert.True(t, strings.HasPrefix(r, sk.KeyWord))

	err = sk.Close()
	assert.Nil(t, err)
}

func sendAudio(audioFile string, sr *SpeechRecognitionStream, t *testing.T) {
	stream, err := speech.NewAudioDataStreamFromWavFileInput(audioFile)
	assert.Nil(t, err)
	if err == nil {
		defer stream.Close()
	}

	data := make([]byte, 1024)
	n, err := stream.Read(data)
	for ; err == nil; n, err = stream.Read(data) {
		fmt.Printf("send %d bytes\n", n)
		assert.Nil(t, sr.SpeechToText(data))
	}
	assert.ErrorIs(t, err, io.EOF)

	fmt.Println("audio send finished")

	r, err := sr.Result()
	fmt.Println(r)
	assert.Nil(t, err)
	// for ; err == nil; r, err = sr.Result() {
	// 	fmt.Println("-----", r)
	// }

	// assert.ErrorIs(t, err, io.EOF)
}
