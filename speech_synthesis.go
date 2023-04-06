package hal

import (
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/Microsoft/cognitive-services-speech-sdk-go/audio"
	"github.com/Microsoft/cognitive-services-speech-sdk-go/common"
	"github.com/Microsoft/cognitive-services-speech-sdk-go/speech"
)

const AUDIO_FORMT = common.Audio16Khz32KBitRateMonoMp3

var (
	ErrSpeechSynthesisTimeout = errors.New("speech synthesis get result timeout")
	MaxSpeechSynthesisDelay   = time.Duration(10)
)

type SpeechSynthesis interface {
	TextToSpeech(text string) error
	Result() (*WordBoundery, []byte, error)
	Error() error
	Close() error
}

type SpeechSynthesisStream struct {
	audioConfig       *audio.AudioConfig
	speechConfig      *speech.SpeechConfig
	languageConfig    *speech.AutoDetectSourceLanguageConfig
	speechSynthesizer *speech.SpeechSynthesizer
	result            *speechSynthesisResult
}

type speechSynthesisResult struct {
	audio     chan []byte
	subtitles chan *WordBoundery
	finished  chan bool
	cancelled chan error
	outcome   chan speech.SpeechSynthesisOutcome
}

func newSpeechSynthesisResult() *speechSynthesisResult {
	return &speechSynthesisResult{audio: make(chan []byte), subtitles: make(chan *WordBoundery), finished: make(chan bool), cancelled: make(chan error)}
}

func (r *speechSynthesisResult) Write(buffer []byte) int {
	n := len(buffer)
	r.audio <- buffer

	return n
}

func (r *speechSynthesisResult) CloseStream() {
	close(r.audio)
	close(r.subtitles)
	close(r.finished)
	close(r.cancelled)
	if r.outcome != nil {
		close(r.outcome)
	}
}

func (r *speechSynthesisResult) Error() error {
	select {
	case res := <-r.outcome:
		defer res.Close()
		return res.Error
	case <-time.After(MaxSpeechSynthesisDelay * time.Second):
		return ErrSpeechSynthesisTimeout
	}
}

func (r *speechSynthesisResult) Result() (sub *WordBoundery, audio []byte, err error) {
	select {
	case audio = <-r.audio:
		return nil, audio, nil
	case sub = <-r.subtitles:
		return sub, nil, nil
	case <-r.finished:
		return nil, nil, io.EOF
	case err = <-r.cancelled:
		return nil, nil, err
	case <-time.After(MaxSpeechSynthesisDelay * time.Second):
		return nil, nil, ErrSpeechSynthesisTimeout
	}
}

// func (r *speechSynthesisResult) GetAudio() ([]byte, error) {
// 	select {
// 	case res := <-r.audio:
// 		return res, r.err
// 	case <-r.finished:
// 		return nil, r.err
// 	case <-time.After(MaxSpeechSynthesisDelay * time.Second):
// 		return nil, ErrSpeechSynthesisTimeout
// 	}
// }

// func (r *speechSynthesisResult) GetSubtitle() (*WordBoundery, error) {
// 	select {
// 	case res := <-r.subtitles:
// 		return res, nil
// 	case <-r.endSubtitles:
// 		return nil, io.EOF
// 	case r.err = <-r.cancelled:
// 		return nil, r.err
// 	case <-time.After(MaxSpeechSynthesisDelay * time.Second):
// 		return nil, ErrSpeechSynthesisTimeout
// 	}
// }

// type noResult struct {
// }

// func (r noResult) Write(buffer []byte) int {
// 	return len(buffer)
// }

// func (r noResult) CloseStream() {

// }

func NewSpeechSynthesisStream(key, region string, voice string) (*SpeechSynthesisStream, error) {
	result := newSpeechSynthesisResult()
	audioOutputStream, err := audio.CreatePushAudioOutputStream(result)
	if err != nil {
		return nil, err
	}

	audioConfig, err := audio.NewAudioConfigFromStreamOutput(audioOutputStream)
	if err != nil {
		defer audioOutputStream.Close()
		return nil, err
	}

	speechConfig, err := speech.NewSpeechConfigFromSubscription(key, region)
	if err != nil {
		defer audioConfig.Close()
		return nil, err
	}

	speechConfig.SetSpeechSynthesisOutputFormat(AUDIO_FORMT)
	speechConfig.SetSpeechSynthesisVoiceName(voice)

	speechSynthesizer, err := speech.NewSpeechSynthesizerFromConfig(speechConfig, audioConfig)
	if err != nil {
		defer speechConfig.Close()
		return nil, err
	}

	res := &SpeechSynthesisStream{
		audioConfig:       audioConfig,
		speechConfig:      speechConfig,
		speechSynthesizer: speechSynthesizer,
		result:            result,
	}

	speechSynthesizer.SynthesisStarted(res.synthesizeStartedHandler)
	speechSynthesizer.Synthesizing(res.synthesizingHandler)
	speechSynthesizer.SynthesisCompleted(res.synthesizedHandler)
	speechSynthesizer.SynthesisCanceled(res.cancelledHandler)
	speechSynthesizer.WordBoundary(res.wordBoundaryHandler)

	return res, nil
}

func NewAutoDetectedSpeechSynthesisStream(key, region string) (*SpeechSynthesisStream, error) {
	result := newSpeechSynthesisResult()
	audioOutputStream, err := audio.CreatePushAudioOutputStream(result)
	if err != nil {
		return nil, err
	}

	audioConfig, err := audio.NewAudioConfigFromStreamOutput(audioOutputStream)
	if err != nil {
		defer audioOutputStream.Close()
		return nil, err
	}

	speechConfig, err := speech.NewSpeechConfigFromSubscription(key, region)
	if err != nil {
		defer audioConfig.Close()
		return nil, err
	}

	languageConfig, err := speech.NewAutoDetectSourceLanguageConfigFromOpenRange()
	if err != nil {
		defer speechConfig.Close()
		return nil, err
	}

	speechConfig.SetSpeechSynthesisOutputFormat(AUDIO_FORMT)

	speechSynthesizer, err := speech.NewSpeechSynthesizerFomAutoDetectSourceLangConfig(speechConfig, languageConfig, audioConfig)
	if err != nil {
		defer speechConfig.Close()
		return nil, err
	}

	res := &SpeechSynthesisStream{
		audioConfig:       audioConfig,
		speechConfig:      speechConfig,
		languageConfig:    languageConfig,
		speechSynthesizer: speechSynthesizer,
		result:            result,
	}

	speechSynthesizer.SynthesisStarted(res.synthesizeStartedHandler)
	speechSynthesizer.Synthesizing(res.synthesizingHandler)
	speechSynthesizer.SynthesisCompleted(res.synthesizedHandler)
	speechSynthesizer.SynthesisCanceled(res.cancelledHandler)
	speechSynthesizer.WordBoundary(res.wordBoundaryHandler)

	return res, nil
}

func (s *SpeechSynthesisStream) TextToSpeech(text string) error {
	s.result.outcome = s.speechSynthesizer.StartSpeakingTextAsync(text)
	return nil
}

func (s *SpeechSynthesisStream) Result() (*WordBoundery, []byte, error) {
	return s.result.Result()
}

func (s *SpeechSynthesisStream) Error() error {
	return s.result.Error()
}

func (s *SpeechSynthesisStream) Close() {
	s.speechSynthesizer.Close()
	s.audioConfig.Close()
	s.speechConfig.Close()
	if s.languageConfig != nil {
		s.languageConfig.Close()
	}
}

func (s *SpeechSynthesisStream) synthesizeStartedHandler(event speech.SpeechSynthesisEventArgs) {
	defer event.Close()
	tlog.Debugf("Synthesis started.")
}

func (s *SpeechSynthesisStream) synthesizingHandler(event speech.SpeechSynthesisEventArgs) {
	defer event.Close()
	tlog.Debugf("Synthesizing, audio chunk size %d.", len(event.Result.AudioData))
	// s.result.audio <- event.Result.AudioData
}

func (s *SpeechSynthesisStream) synthesizedHandler(event speech.SpeechSynthesisEventArgs) {
	defer event.Close()
	tlog.Debugf("Synthesized, audio length %d.", len(event.Result.AudioData))
	// s.result.audio <- event.Result.AudioData
	s.result.finished <- true
}

func (s *SpeechSynthesisStream) cancelledHandler(event speech.SpeechSynthesisEventArgs) {
	defer event.Close()
	c, _ := speech.NewCancellationDetailsFromSpeechSynthesisResult(&event.Result)
	err := fmt.Errorf("CANCELED:\n Reason=%d.\nErrorCode=%d\nErrorDetails=[%s]", c.Reason, c.ErrorCode, c.ErrorDetails)
	tlog.Errorf(err.Error())
	s.result.cancelled <- err
}

func (s *SpeechSynthesisStream) wordBoundaryHandler(event speech.SpeechSynthesisWordBoundaryEventArgs) {
	defer event.Close()
	w := WordBoundery{
		BounderyType: event.BoundaryType,
		AudioOffset:  (float64(event.AudioOffset) + 5000.0) / 10000.0,
		Duration:     event.Duration.Seconds() * 1000,
		TextOffset:   event.TextOffset,
		WordLength:   event.WordLength,
		Text:         event.Text,
	}

	tlog.Debugf(w.String())

	s.result.subtitles <- &w
}

type SpeechSynthesisStandalone struct {
	SpeechSynthesisStream
}

func NewSpeechSynthesisStandalone(key, region string, voice string) (*SpeechSynthesisStandalone, error) {
	audioConfig, err := audio.NewAudioConfigFromDefaultSpeakerOutput()
	if err != nil {
		return nil, err
	}

	speechConfig, err := speech.NewSpeechConfigFromSubscription(key, region)
	if err != nil {
		defer audioConfig.Close()
		return nil, err
	}

	speechConfig.SetSpeechSynthesisOutputFormat(AUDIO_FORMT)
	speechConfig.SetSpeechSynthesisVoiceName(voice)

	speechSynthesizer, err := speech.NewSpeechSynthesizerFromConfig(speechConfig, audioConfig)
	if err != nil {
		defer speechConfig.Close()
		return nil, err
	}

	res := &SpeechSynthesisStandalone{}

	res.audioConfig = audioConfig
	res.speechConfig = speechConfig
	res.speechSynthesizer = speechSynthesizer
	res.result = newSpeechSynthesisResult()

	speechSynthesizer.SynthesisStarted(res.synthesizeStartedHandler)
	speechSynthesizer.Synthesizing(res.synthesizingHandler)
	speechSynthesizer.SynthesisCompleted(res.synthesizedHandler)
	speechSynthesizer.SynthesisCanceled(res.cancelledHandler)
	speechSynthesizer.WordBoundary(res.wordBoundaryHandler)

	return res, nil
}

func NewAutoDetectedSpeechSynthesisStandalone(key, region string) (*SpeechSynthesisStandalone, error) {
	audioConfig, err := audio.NewAudioConfigFromDefaultSpeakerOutput()
	if err != nil {
		return nil, err
	}

	speechConfig, err := speech.NewSpeechConfigFromSubscription(key, region)
	if err != nil {
		defer audioConfig.Close()
		return nil, err
	}

	languageConfig, err := speech.NewAutoDetectSourceLanguageConfigFromOpenRange()
	if err != nil {
		defer speechConfig.Close()
		return nil, err
	}

	speechConfig.SetSpeechSynthesisOutputFormat(AUDIO_FORMT)

	speechSynthesizer, err := speech.NewSpeechSynthesizerFomAutoDetectSourceLangConfig(speechConfig, languageConfig, audioConfig)
	if err != nil {
		defer speechConfig.Close()
		return nil, err
	}

	res := &SpeechSynthesisStandalone{}

	res.audioConfig = audioConfig
	res.speechConfig = speechConfig
	res.languageConfig = languageConfig
	res.speechSynthesizer = speechSynthesizer
	res.result = newSpeechSynthesisResult()

	speechSynthesizer.SynthesisStarted(res.synthesizeStartedHandler)
	speechSynthesizer.Synthesizing(res.synthesizingHandler)
	speechSynthesizer.SynthesisCompleted(res.synthesizedHandler)
	speechSynthesizer.SynthesisCanceled(res.cancelledHandler)
	speechSynthesizer.WordBoundary(res.wordBoundaryHandler)

	return res, nil
}

type WordBoundery struct {
	BounderyType common.SpeechSynthesisBoundaryType
	AudioOffset  float64
	Duration     float64
	TextOffset   uint
	WordLength   uint
	Text         string
}

func (w WordBoundery) String() string {
	return fmt.Sprintf("BoundaryType: %d\nAudioOffset: %f (ms)\nDuration: %f (ms)\nTextOffset: %d\nWordLen: %d\nText: %s\n",
		w.BounderyType, w.AudioOffset, w.Duration, w.TextOffset, w.WordLength, w.Text)
}

func (s *SpeechSynthesisStream) GetAllSupportLanguage() map[string]bool {
	voices := s.speechSynthesizer.GetVoicesAsync("")
	vs := <-voices

	if vs.Result.Reason == common.VoicesListRetrieved {
		set := map[string]bool{}
		for _, v := range vs.Result.Voices {
			set[v.Locale] = true
		}

		return set
	}

	if vs.Error != nil {
		tlog.Errorf(vs.Result.ErrorDetails)
	}

	return nil
}

type Voice struct {
	Name      string
	LocalName string
	Gender    string
	StyleList []string
}

func (v *Voice) String() string {
	return fmt.Sprintf("Name: %s, LocalName: %s, Geneder: %s, Style: %s", v.Name, v.LocalName, v.Gender, strings.Join(v.StyleList, ","))
}

func (s *SpeechSynthesisStream) GetAllSupportVoicesForLanguage(language string) []*Voice {
	voices := s.speechSynthesizer.GetVoicesAsync(language)
	v := <-voices
	if v.Result.Reason == common.VoicesListRetrieved {
		res := make([]*Voice, 0, len(v.Result.Voices))
		for _, v := range v.Result.Voices {
			res = append(res, &Voice{
				Name:      v.ShortName,
				LocalName: v.LocalName,
				Gender:    gender(v.Gender),
				StyleList: v.StyleList,
			})
		}

		return res
	}

	if v.Error != nil {
		tlog.Errorf(v.Result.ErrorDetails)
	}

	return nil
}

func FormatVoices(voices []*Voice) string {
	var b strings.Builder
	b.WriteString("    Name\t\tGender\tSytle\n")

	for i, v := range voices {
		b.WriteString(fmt.Sprintf("%d. %s\t\t%s\t%s\n", i+1, v.LocalName, v.Gender, strings.Join(v.StyleList, ",")))
	}

	return b.String()
}

func gender(gender common.SynthesisVoiceGender) string {
	if gender == 1 {
		return "female"
	}

	return "male"
}
