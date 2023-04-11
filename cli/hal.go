package main

import (
	"errors"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strings"
	"syscall"

	hal "github.com/neotse/hal"
	"github.com/sashabaranov/go-openai"
)

func init() {
	hal.PARAMS.LoadParams("params.json")
	hal.CHATGPTS.LoadChatGPTs("sessions.json")
	hal.HOOKS.LoadHooks("hooks.json")
}

var (
	forceInit     bool
	verbose       bool
	slient        bool
	listSession   bool
	selectSession bool
	deleteSession bool
	createSession bool
	configSession bool
	maxHistory    int
	language      string
	voice         string
	chatGPTModel  string
	stopWord      string

	showKeyword     bool
	akeyword        string
	keywordModel    string
	keywordLanguage string
)

func parseArgs() bool {
	flag.BoolVar(&forceInit, "init", false, "following a process to setup HAL (recommend for the first use).")
	flag.BoolVar(&verbose, "verbose", false, "show more details.")
	flag.BoolVar(&slient, "slient", false, "keep HAL slient.")
	flag.IntVar(&maxHistory, "history", hal.PARAMS.MaxHistory, "the max history you want to keep when talk to chatgpt (Warning: the more history you have, the more tokens you use).")
	flag.StringVar(&language, "language", "", "the language you want to talking with HAL. (see https://learn.microsoft.com/en-us/azure/cognitive-services/speech-service/language-support?tabs=stt)")
	flag.StringVar(&voice, "voice", "", "the voice you want to HAL speaking if you allow it to speak. (see https://learn.microsoft.com/en-us/azure/cognitive-services/speech-service/language-support?tabs=tts).")
	flag.StringVar(&chatGPTModel, "model", "", "the model you want to use int chatgpt (gpt-4-32k-0314, gpt-4-32k, gpt-4-0314, gpt-4, gpt-3.5-turbo-0301, gpt-3.5-turbo).")
	flag.StringVar(&stopWord, "stopWord", "", "the keyword used to deactivate HAL.")
	session := flag.NewFlagSet("session", flag.ExitOnError)
	session.BoolVar(&listSession, "list", false, "list current chatgpt sessions.")
	session.BoolVar(&deleteSession, "delete", false, "delete the 'session' for talk. ")
	session.BoolVar(&selectSession, "select", false, "select the 'session' for start to talk. If not set, it will select the session recently used.")
	session.BoolVar(&createSession, "create", false, "create the 'session' for talk. ")
	session.BoolVar(&configSession, "config", false, "config the 'session' for talk. ")
	keyword := flag.NewFlagSet("keyword", flag.ExitOnError)
	keyword.BoolVar(&showKeyword, "show", false, "show the current config of keyword for activate.")
	keyword.StringVar(&akeyword, "keyword", "", "set the keyword for activate (case insensitive), path and lang must be set at same time.")
	keyword.StringVar(&keywordModel, "path", "", "set the path of model file of keyword.")
	keyword.StringVar(&keywordLanguage, "lang", "", "set the language of keyword. (see https://learn.microsoft.com/en-us/azure/cognitive-services/speech-service/language-support?tabs=stt)")

	flag.Parse()
	if len(os.Args) > 1 {
		if os.Args[1] == "session" {
			session.Parse(os.Args[2:])
		} else if os.Args[1] == "keyword" {
			keyword.Parse(os.Args[2:])
		}
	}

	if listSession {
		hal.ListSessions()
		return true
	} else if deleteSession {
		hal.DeleteSession()
		return true
	} else if selectSession {
		hal.SelectSession()
		return true
	} else if createSession {
		hal.CreateASession()
		return true
	} else if configSession {
		hal.ConfigSession()
		return true
	} else if showKeyword {
		hal.Showkeyword()
		return true
	}

	if akeyword != "" {
		if keywordModel == "" || keywordLanguage == "" {
			panic("path and lang must be set at the same time")
		}

		_, err := os.Stat(keywordModel)
		if err != nil {
			panic(err)
		}

		hal.PARAMS.Keyword = akeyword
		hal.PARAMS.KeywordModel = keywordModel
		hal.PARAMS.KeywordLanguage = keywordLanguage

		return true
	}

	// the priorty of value from CLI params is higher
	hal.PARAMS.MaxHistory = maxHistory
	if language != "" {
		hal.PARAMS.Language = language
	}

	if chatGPTModel != "" {
		hal.PARAMS.ChatgptModel = chatGPTModel
	}

	if voice != "" {
		hal.PARAMS.Voice = voice
	}

	if stopWord != "" {
		hal.PARAMS.StopWord = stopWord
	}

	if verbose {
		hal.SetLevel(hal.DEBUG)
	}

	return false
}

func registerSignalHandler() {
	schan := make(chan os.Signal, 2)
	signal.Notify(schan, os.Interrupt, syscall.SIGTERM)
	go func() {
		<-schan
		hal.PARAMS.SaveParams("params.json")
		fmt.Println("Params saved.")

		hal.CHATGPTS.SaveChatGPTs("sessions.json")
		fmt.Println("Sessions saved.")

		hal.HOOKS.SaveHooks("hooks.json")
		fmt.Println("Hooks saved.")
		os.Exit(1)
	}()

	runtime.Gosched()
}

func main() {
	registerSignalHandler()
	if parseArgs() {
		return
	}

	// first init or force init
	if forceInit || !hal.PARAMS.Initialized {
		hal.Initialize(chatGPTModel, language, voice, stopWord)
	}

	var p = hal.PARAMS
	fmt.Println("Params:")
	fmt.Println(p)

	var sr *hal.SpeechRecognitionStandalone
	var err error
	l := p.Language
	if l != "" {
		sr, err = hal.NewSpeechRecognitionStandalone(p.SpeechKey, p.SpeechRegion, []string{l})
	} else {
		sr, err = hal.NewAutoDetectedSpeechRecognitionStandalone(p.SpeechKey, p.SpeechRegion)
		l = "auto detected"
	}

	if err != nil {
		panic(err)
	}

	sk, err := hal.NewKeywordRecognitionStandalone(p.SpeechKey, p.SpeechRegion, []string{p.KeywordLanguage}, p.KeywordModel, p.Keyword)
	if err != nil {
		panic(err)
	}

	err = sk.Start()
	if err != nil {
		panic(err)
	}

	defer sk.Close()
	fmt.Printf("Speech Recognition Initialized. Language: %s\n", l)
	defer sr.Close()

	var ss *hal.SpeechSynthesisStandalone
	// slient without Speech Synthesis
	if !slient {
		var err error
		if p.Voice != "" {
			ss, err = hal.NewSpeechSynthesisStandalone(p.SpeechKey, p.SpeechRegion, p.Voice)
		} else {
			ss, err = hal.NewAutoDetectedSpeechSynthesisStandalone(p.SpeechKey, p.SpeechRegion)
			p.Voice = "auto detected"
		}

		if err != nil {
			panic(err)
		}

		defer ss.Close()

		fmt.Printf("Speech Synthesis Initialized. Voice: %s\n", p.Voice)
	}

	name, cg := hal.CHATGPTS.GetDefaultGPT()
	if cg == nil { // no default chatGPT or not specified session, create default session
		cg = hal.CHATGPTS.NewDefaultSession(hal.PARAMS.OpenaiKey, hal.PARAMS.ChatgptModel)
		cg.IsDefault = true
	}

	initHooksChatGPT()

	hal.CHATGPTS.SaveChatGPTs("sessions.json")

	fmt.Printf("ChatGPT Initialized. Name: %s, Model: %s\n", name, cg.Model)

	for {
		fmt.Printf("Say %s activate and Say %s deactivate\n", sk.KeyWord, hal.PARAMS.StopWord)
		r, err := sk.Result()
		if err != nil {
			panic(err)
		}

		fmt.Printf("%s here. \n", r)
		var blank int
		for {
			if blank >= 3 {
				fmt.Println("long time no speak, deactivated.")
				break
			}

			fmt.Println("Please speaking")
			sr.Start()
			text, err := sr.Result()
			if err != nil {
				if !errors.Is(err, hal.ErrSpeechRecognitionTimeout) {
					panic(err)
				} else {
					fmt.Println(err)
					continue
				}
			}

			if text == "" {
				blank++
				continue
			}

			if strings.HasPrefix(text, hal.PARAMS.StopWord) {
				fmt.Println("See you later :)")
				break
			}

			if hook(text) {
				continue
			}

			fmt.Println("Prompt:\n", text)

			res, err := cg.PromptStream(text)
			if err != nil {
				fmt.Println(err)
				continue
			}

			fmt.Println("ChatGPT:")
			streamSpitter := hal.NewStreamSplitter(res)
			for content := streamSpitter.Segment(true); content != ""; content = streamSpitter.Segment(true) {
				if !slient {
					err = ss.TextToSpeech(content)
					if err != nil {
						panic(err)
					}
				}

				for _, _, err := ss.Result(); err == nil; _, _, err = ss.Result() {
					if err != nil {
						fmt.Println(err)
						break
					}
				}

				if ss.Error() != nil {
					panic(err)
				}
			}

			fmt.Println()
		}
	}
}

func initHooksChatGPT() {
	hooks := hal.CHATGPTS.NewSessionWithName("hooks", hal.PARAMS.OpenaiKey, openai.GPT3Dot5Turbo) // fix model
	if hooks.System != nil {
		return
	}

	var b strings.Builder
	b.WriteString("keep these items in mind, I'll need them later:\n")
	for _, config := range hal.HOOKS.Configs {
		b.WriteString(fmt.Sprintf("keyword: %s\n", config.Keyword))
		b.WriteString(fmt.Sprintf("hook: %s\n", config.HookName))
	}

	_, _, err := hooks.Prompt(b.String())
	if err != nil {
		panic("init hooks: " + err.Error())
	}

	_, _, err = hooks.Prompt("I will send you some keywords, and you only need to output just the value of the corresponding hook. If the keywords is not in the provided item, just output unknown.")
	if err != nil {
		panic("init hooks: " + err.Error())
	}

	_, _, err = hooks.Prompt("and please remember that I only have the value of the hook, no other output is required")
	if err != nil {
		panic("init hooks: " + err.Error())
	}
	// freeze history
	hooks.SetMaxHistory(0)
}

func hook(text string) bool {
	hooks := hal.CHATGPTS.Clients["hooks"]
	resp, _, err := hooks.Prompt(text)
	if err != nil {
		panic(err)
	}

	hook := strings.Trim(resp, " \n")

	if strings.ToLower(hook) == "unknown" {
		return false
	}

	if !hal.HOOKS.IsExist(hook) {
		return false
	}

	err = hal.HOOKS.Exec(hook)
	if err != nil {
		panic(err)
	}

	return true
}
