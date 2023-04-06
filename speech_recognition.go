package hal

import (
	"errors"
	"fmt"
	"time"

	"github.com/Microsoft/cognitive-services-speech-sdk-go/audio"
	"github.com/Microsoft/cognitive-services-speech-sdk-go/common"
	"github.com/Microsoft/cognitive-services-speech-sdk-go/speech"
)

var (
	ErrSpeechRecognitionTimeout                    = errors.New("speech recognition get result timeout")
	MaxSpeechRecognitionDelay                      = time.Duration(30)
	SegmentationSilenceTimeoutMs                   = "1500" //1.5s
	SpeechServiceConnectionInitialSilenceTimeoutMs = "5000" //5s
)

type SpeechRecognition interface {
	Start() error
	SpeechToText(data []byte) error
	Result() (string, error)
	Close() error
}

type SpeechRecognitionStream struct {
	audioConfig      *audio.AudioConfig
	speechConfig     *speech.SpeechConfig
	languageConfig   *speech.AutoDetectSourceLanguageConfig
	speechRecognizer *speech.SpeechRecognizer
	audioInputStream *audio.PushAudioInputStream
	result           speechRecognitionResult
}

type speechRecognitionResult struct {
	text      chan string
	finished  chan bool
	cancelled chan error
	outcome   chan speech.SpeechRecognitionOutcome
}

func (s speechRecognitionResult) GetResult() (string, error) {
	select {
	case res := <-s.outcome:
		defer res.Close()
		return res.Result.Text, res.Error
	case <-time.After(MaxSpeechRecognitionDelay * time.Second):
		return "", ErrSpeechRecognitionTimeout
	}
}

func (s speechRecognitionResult) Close() {
	close(s.text)
	close(s.finished)
	close(s.cancelled)
	if s.outcome != nil {
		close(s.outcome)
	}
}

func newSpeechConfig(key, region string) (*speech.SpeechConfig, error) {
	speechConfig, err := speech.NewSpeechConfigFromSubscription(key, region)
	speechConfig.SetProperty(common.SegmentationSilenceTimeoutMs, SegmentationSilenceTimeoutMs)
	speechConfig.SetProperty(common.SpeechServiceConnectionInitialSilenceTimeoutMs, SpeechServiceConnectionInitialSilenceTimeoutMs)

	return speechConfig, err
}

func NewSpeechRecognitionStream(key, region string, languages []string) (*SpeechRecognitionStream, error) {
	audioInputStream, err := audio.CreatePushAudioInputStream()
	if err != nil {
		return nil, err
	}

	audioConfig, err := audio.NewAudioConfigFromStreamInput(audioInputStream)
	if err != nil {
		audioInputStream.Close()
		return nil, err
	}

	speechConfig, err := newSpeechConfig(key, region)
	if err != nil {
		audioConfig.Close()
		return nil, err
	}

	languageConfig, err := speech.NewAutoDetectSourceLanguageConfigFromLanguages(languages)
	if err != nil {
		speechConfig.Close()
		return nil, err
	}

	speechRecognizer, err := speech.NewSpeechRecognizerFomAutoDetectSourceLangConfig(speechConfig, languageConfig, audioConfig)
	if err != nil {
		languageConfig.Close()
		return nil, err
	}

	res := &SpeechRecognitionStream{
		audioConfig:      audioConfig,
		speechConfig:     speechConfig,
		languageConfig:   languageConfig,
		speechRecognizer: speechRecognizer,
		audioInputStream: audioInputStream,
		result:           speechRecognitionResult{text: make(chan string), finished: make(chan bool), cancelled: make(chan error)},
	}

	speechRecognizer.SessionStarted(res.sessionStartedHandler)
	speechRecognizer.SessionStopped(res.sessionStoppedHandler)
	speechRecognizer.Recognizing(res.recognizingHandler)
	speechRecognizer.Recognized(res.recognizedHandler)
	speechRecognizer.Canceled(res.cancelledHandler)
	speechRecognizer.SpeechStartDetected(res.speechStartHandler)
	speechRecognizer.SpeechEndDetected(res.speechEndHandler)

	return res, err
}

func NewAutoDetectedSpeechRecognitionStream(key, region string) (*SpeechRecognitionStream, error) {
	return NewSpeechRecognitionStream(key, region, GetAutoDetectedLanguages())
}

func (s *SpeechRecognitionStream) Start() error {
	s.result.outcome = s.speechRecognizer.RecognizeOnceAsync()
	return nil
}

func (s *SpeechRecognitionStream) SpeechToText(data []byte) error {
	return s.audioInputStream.Write(data)
}

func (s *SpeechRecognitionStream) Result() (string, error) {
	// select {
	// case res := <-s.result.text:
	// 	return res, nil
	// case <-s.result.finished:
	// 	return "", io.EOF
	// case err := <-s.result.cancelled:
	// 	return "", err
	// case <-time.After(MaxSpeechRecognitionDelay * time.Second):
	// 	return "", ErrSpeechRecognitionTimeout
	// }
	return s.result.GetResult()
}

func (s *SpeechRecognitionStream) Close() error {
	s.audioConfig.Close()
	s.speechConfig.Close()
	s.languageConfig.Close()
	s.speechRecognizer.Close()
	s.result.Close()

	return nil
}

func (s *SpeechRecognitionStream) sessionStartedHandler(event speech.SessionEventArgs) {
	defer event.Close()
	tlog.Debugf("Session Started (ID=%s)", event.SessionID)
}

func (s *SpeechRecognitionStream) sessionStoppedHandler(event speech.SessionEventArgs) {
	defer event.Close()
	tlog.Debugf("Session Stopped (ID=%s)", event.SessionID)
	// s.result.finished <- true
}

func (s *SpeechRecognitionStream) recognizingHandler(event speech.SpeechRecognitionEventArgs) {
	defer event.Close()
	tlog.Debugf("Recognizing: ", event.Result.Text)
	// s.result.text <- event.Result.Text
}

func (s *SpeechRecognitionStream) recognizedHandler(event speech.SpeechRecognitionEventArgs) {
	defer event.Close()
	tlog.Debugf("Recognized: ", event.Result.Text)
	// s.result.text <- event.Result.Text
}

func (s *SpeechRecognitionStream) cancelledHandler(event speech.SpeechRecognitionCanceledEventArgs) {
	defer event.Close()
	err := fmt.Errorf("CANCELED:\n Reason=%d.\nErrorCode=%d\nErrorDetails=[%s]", event.Reason, event.ErrorCode, event.ErrorDetails)
	tlog.Errorf(err.Error())
	// s.result.cancelled <- err
}

func (s *SpeechRecognitionStream) speechStartHandler(event speech.RecognitionEventArgs) {
	defer event.Close()
	tlog.Debugf("Speech Start at: %d", event.Offset)
}

func (s *SpeechRecognitionStream) speechEndHandler(event speech.RecognitionEventArgs) {
	defer event.Close()
	tlog.Debugf("Speech End at: %d", event.Offset)
}

type SpeechRecognitionStandalone struct {
	SpeechRecognitionStream
}

func NewSpeechRecognitionStandalone(key, region string, languages []string) (*SpeechRecognitionStandalone, error) {
	audioConfig, err := audio.NewAudioConfigFromDefaultMicrophoneInput()
	if err != nil {
		return nil, err
	}

	speechConfig, err := newSpeechConfig(key, region)
	if err != nil {
		audioConfig.Close()
		return nil, err
	}

	languageConfig, err := speech.NewAutoDetectSourceLanguageConfigFromLanguages(languages)
	if err != nil {
		speechConfig.Close()
		return nil, err
	}

	speechRecognizer, err := speech.NewSpeechRecognizerFomAutoDetectSourceLangConfig(speechConfig, languageConfig, audioConfig)
	if err != nil {
		languageConfig.Close()
		return nil, err
	}

	res := &SpeechRecognitionStandalone{}
	res.audioConfig = audioConfig
	res.speechConfig = speechConfig
	res.speechRecognizer = speechRecognizer
	res.languageConfig = languageConfig
	res.result = speechRecognitionResult{text: make(chan string), finished: make(chan bool), cancelled: make(chan error)}

	speechRecognizer.SessionStarted(res.sessionStartedHandler)
	speechRecognizer.SessionStopped(res.sessionStoppedHandler)
	speechRecognizer.Recognizing(res.recognizingHandler)
	speechRecognizer.Recognized(res.recognizedHandler)
	speechRecognizer.Canceled(res.cancelledHandler)
	speechRecognizer.SpeechStartDetected(res.speechStartHandler)
	speechRecognizer.SpeechEndDetected(res.speechEndHandler)

	return res, err
}

func NewAutoDetectedSpeechRecognitionStandalone(key, region string) (*SpeechRecognitionStandalone, error) {
	return NewSpeechRecognitionStandalone(key, region, GetAutoDetectedLanguages())
}

func (s *SpeechRecognitionStandalone) SpeechToText(data []byte) error {
	panic("not supperted")
}

type KeywordRecognitionStandalone struct {
	SpeechRecognitionStream
	model   *speech.KeywordRecognitionModel
	KeyWord string
}

func NewKeywordRecognitionStandalone(key, region string, languages []string, model, keyWord string) (*KeywordRecognitionStandalone, error) {
	audioConfig, err := audio.NewAudioConfigFromDefaultMicrophoneInput()
	if err != nil {
		return nil, err
	}

	speechConfig, err := speech.NewSpeechConfigFromSubscription(key, region)
	if err != nil {
		audioConfig.Close()
		return nil, err
	}

	languageConfig, err := speech.NewAutoDetectSourceLanguageConfigFromLanguages(languages)
	if err != nil {
		speechConfig.Close()
		return nil, err
	}

	m, err := speech.NewKeywordRecognitionModelFromFile(model)
	if err != nil {
		languageConfig.Close()
		return nil, err
	}

	speechRecognizer, err := speech.NewSpeechRecognizerFomAutoDetectSourceLangConfig(speechConfig, languageConfig, audioConfig)
	if err != nil {
		m.Close()
		return nil, err
	}

	res := &KeywordRecognitionStandalone{}
	res.audioConfig = audioConfig
	res.speechConfig = speechConfig
	res.languageConfig = languageConfig
	res.speechRecognizer = speechRecognizer
	res.model = m
	res.KeyWord = keyWord
	res.result = speechRecognitionResult{text: make(chan string), finished: make(chan bool), cancelled: make(chan error)}

	speechRecognizer.SessionStarted(res.sessionStartedHandler)
	speechRecognizer.SessionStopped(res.sessionStoppedHandler)
	speechRecognizer.Recognizing(res.recognizingHandler)
	speechRecognizer.Recognized(res.recognizedHandler)
	speechRecognizer.Canceled(res.cancelledHandler)
	speechRecognizer.SpeechStartDetected(res.speechStartHandler)
	speechRecognizer.SpeechEndDetected(res.speechEndHandler)

	return res, err
}

func (s *KeywordRecognitionStandalone) Start() error {
	errChan := s.speechRecognizer.StartKeywordRecognitionAsync(*s.model)
	defer close(errChan)

	err := ErrSpeechRecognitionTimeout
	select {
	case err = <-errChan:
	case <-time.After(5 * time.Second):
	}

	return err
}

func (s *KeywordRecognitionStandalone) SpeechToText(data []byte) error {
	panic("not supported")
}

func (s *KeywordRecognitionStandalone) Result() (string, error) {
	select {
	case res := <-s.result.text:
		return res, nil
	case err := <-s.result.cancelled:
		return "", err
	}
}

func (s *KeywordRecognitionStandalone) Close() error {
	s.audioConfig.Close()
	s.speechConfig.Close()
	s.speechRecognizer.Close()
	s.result.Close()
	s.model.Close()
	errChan := s.speechRecognizer.StopKeywordRecognitionAsync()
	defer close(errChan)

	err := ErrSpeechRecognitionTimeout
	select {
	case err = <-errChan:
	case <-time.After(5 * time.Second):
	}

	return err
}

func (s *KeywordRecognitionStandalone) recognizingHandler(event speech.SpeechRecognitionEventArgs) {
	defer event.Close()
	tlog.Debugf("Recognizing: ", event.Result.Text)
}

func (s *KeywordRecognitionStandalone) sessionStartedHandler(event speech.SessionEventArgs) {
	defer event.Close()
	tlog.Debugf("Session Started (ID=%s)", event.SessionID)
	s.result.text <- s.KeyWord
}

// not wait cloud result
// func (s *KeywordRecognitionStandalone) sessionStoppedHandler(event speech.SessionEventArgs) {
// 	defer event.Close()
// 	tlog.Debugf("Session Stopped (ID=%s)", event.SessionID)
// 	s.result.finished <- true
// }

func (s *KeywordRecognitionStandalone) recognizedHandler(event speech.SpeechRecognitionEventArgs) {
	defer event.Close()
	tlog.Debugf("Recognized: ", event.Result.Text)
}
