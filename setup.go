package hal

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"

	"github.com/sashabaranov/go-openai"
)

func Initialize(chatGPTModel, language, voice, stopWord string) {
	PARAMS.OpenaiKey = getOpenaiKey()
	if chatGPTModel == "" {
		PARAMS.ChatgptModel = chooseModel()
	}

	createSession()

	PARAMS.SpeechKey = getSpeechKey()
	PARAMS.SpeechRegion = getSpeechRegion()
	for !checkSpeechKeyAndRegion(PARAMS.SpeechKey, PARAMS.SpeechRegion) {
		PARAMS.SpeechKey = getSpeechKey()
		PARAMS.SpeechRegion = getSpeechRegion()
	}

	if language == "" {
		PARAMS.Language = chooseLanguage()
	}

	if voice == "" {
		PARAMS.Voice = chooseVoice()
	}

	if stopWord == "" {
		PARAMS.StopWord = setStopWord()
	}

	PARAMS.Initialized = true
	PARAMS.SaveParams("params.json")
}

func getOpenaiKey() string {
	err, key := fmt.Errorf("dummy"), ""
	for err != nil {
		fmt.Println("Input your openai API Key (create it on openai https://platform.openai.com/account/api-keys):")
		if PARAMS.OpenaiKey != "" {
			fmt.Printf("Current key: %s. Replace it by input new key or Press Enter keep it.\n", PARAMS.OpenaiKey)
		}

		key := readStringFromStdin()
		if key == "" && PARAMS.OpenaiKey != "" {
			return PARAMS.OpenaiKey
		}

		if !CheckOpenaiKey(key) {
			fmt.Printf("When attempt connect to openai use the provided key, an error occurred. Please check your key or network.\n ERROR: %s\n", err)
		}
	}

	return key
}

func CheckOpenaiKey(key string) bool {
	client := openai.NewClient(key)
	_, err := client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model: openai.GPT3Dot5Turbo,
			Messages: []openai.ChatCompletionMessage{
				{
					Role:    openai.ChatMessageRoleUser,
					Content: "Hello!",
				},
			},
		},
	)

	return err == nil
}

func chooseModel() string {
	curr := PARAMS.ChatgptModel
	if curr == "" {
		curr = "gpt-3.5-turbo"
	}

	var choose int
	for choose < 1 || choose > 6 {
		fmt.Printf("Please choose a chatGPT model (default %s):\n", curr)
		fmt.Println("1. gpt-3.5-turbo\n2. gpt-3.5-turbo-0301\n3. gpt-4\n4. gpt-4-0314\n5. gpt-4-32k\n6. gpt-4-32k-0314")
		choose := readIntFromStdin()
		if choose == -1 {
			return curr
		}
	}

	return models[choose]
}

var models = []string{"", "gpt-3.5-turbo", "gpt-3.5-turbo-0301", "gpt-4", "gpt-4-0314", "gpt-4-32k", "gpt-4-32k-0314"}

func getSpeechKey() string {
	fmt.Println("Input your azure speech key (create it on azure https://learn.microsoft.com/en-us/azure/cognitive-services/cognitive-services-apis-create-account):")
	if PARAMS.SpeechKey != "" {
		fmt.Printf("Current key: %s. Replace it by input new key or Press Enter keep it.\n", PARAMS.SpeechKey)
	}

	skey := readStringFromStdin()
	if skey == "" && PARAMS.SpeechKey != "" {
		return PARAMS.SpeechKey
	}

	return skey
}

func getSpeechRegion() string {
	fmt.Println("Input your azure speech region (create it on azure https://learn.microsoft.com/en-us/azure/cognitive-services/cognitive-services-apis-create-account):")
	if PARAMS.SpeechRegion != "" {
		fmt.Printf("Current region: %s. Replace it by input new key or Press Enter keep it.\n", PARAMS.SpeechRegion)
	}

	sregion := readStringFromStdin()
	if sregion == "" && PARAMS.SpeechRegion != "" {
		return PARAMS.SpeechRegion
	}

	return sregion
}

func checkSpeechKeyAndRegion(key, region string) bool {
	tempSpeechSynthesis, err := NewAutoDetectedSpeechSynthesisStream(key, region)
	if err != nil {
		fmt.Printf("When attempt connect to azure use the provided key and region, an error occurred. Please check your key/region or network.\n ERROR: %s\n", err)
		return false
	}
	tempSpeechSynthesis.Close()

	return true
}

func chooseLanguage() string {
	l := PARAMS.Language
	if l == "" {
		l = "Auto Detected"
	}

	var choose int
	for choose != 2 {
		fmt.Printf("Choose a language you want to talk with HAL (default %s):\n", l)
		fmt.Printf("1. Auto Detected. (support %s)\n", strings.Join(GetAutoDetectedLanguagesName(), ","))
		fmt.Println("2. More languages.")

		choose := readIntFromStdin()
		if choose == -1 {
			return l
		}

		// set language to empty for Auto Detected
		if choose == 1 {
			PARAMS.Language = ""
			return ""
		}
	}

	return selectLanguage()
}

func selectLanguage() string {
	var l string
	var tempSpeechRecognition *SpeechRecognitionStream
	err := fmt.Errorf("dummy")
	for err != nil {
		fmt.Println("Language name (BCP-47 format, more details see https://learn.microsoft.com/en-us/azure/cognitive-services/speech-service/language-support?tabs=stt):")

		l = readStringFromStdin()
		tempSpeechRecognition, err = NewSpeechRecognitionStream(PARAMS.SpeechKey, PARAMS.SpeechRegion, []string{l})
		if err != nil {
			fmt.Println(err)
		}

		tempSpeechRecognition.Close()
	}

	return l
}

func chooseVoice() string {
	v := PARAMS.Voice
	if v == "" {
		v = "Decide by azure"
	}

	var choose int
	for choose != 2 {
		fmt.Printf("The voice of HAL (default %s):\n", v)
		fmt.Println("1. Decide by azure.")
		fmt.Println("2. Choose a voice.")

		choose = readIntFromStdin()
		if choose == -1 {
			return v
		}

		// set voice to empty for default in azure
		if choose == 1 {
			return ""
		}
	}

	return selectVoice(selectVoiceLanguage())
}

func selectVoiceLanguage() string {
	tempSpeechSynthesis, _ := NewAutoDetectedSpeechSynthesisStream(PARAMS.SpeechKey, PARAMS.SpeechRegion)
	languages := tempSpeechSynthesis.GetAllSupportLanguage()
	l := "unknown"
	for !languages[l] {
		fmt.Println("First Inpput the language name HAL speaking (BCP-47 format, more details see https://learn.microsoft.com/en-us/azure/cognitive-services/speech-service/language-support?tabs=tts):")
		l = readStringFromStdin()
	}
	tempSpeechSynthesis.Close()

	return l
}

func selectVoice(l string) string {
	tempSpeechSynthesis, _ := NewAutoDetectedSpeechSynthesisStream(PARAMS.SpeechKey, PARAMS.SpeechRegion)
	voices := tempSpeechSynthesis.GetAllSupportVoicesForLanguage(l)
	fmt.Printf("The supported voices of (%s) are list below. Please choose one:\n", l)
	fmt.Println(FormatVoices(voices))
	idx := readIntFromStdin()
	for idx < 1 || idx > len(voices) {
		fmt.Println("Invalid choose, please choose again:")
		idx = readIntFromStdin()
	}
	tempSpeechSynthesis.Close()

	return voices[idx-1].Name
}

func setStopWord() string {
	w := PARAMS.StopWord
	if w == "" {
		w = "goodbye"
	}

	fmt.Printf("Please input a word or phrase (Default '%s'):\n", w)
	word := readStringFromStdin()
	if word == "" {
		word = w
	}

	return word
}

func CreateASession() {
	var name string
	for name == "" || CHATGPTS.Clients[name] != nil {
		fmt.Println("Please give a session name (case insensitive):")
		name = strings.ToLower(readStringFromStdin())
		if name == "" {
			fmt.Println("Can not be empty. Please input again.")
		} else if CHATGPTS.Clients[name] != nil {
			fmt.Println("Session exist. Please choose other name.")
		}

	}

	fmt.Println("What do you want chatgpt to do?")
	desc := readStringFromStdin()

	gpt := CHATGPTS.NewSessionWithName(name, PARAMS.OpenaiKey, PARAMS.ChatgptModel)
	if desc != "" {
		gpt.SetRole(desc)
	}

	CHATGPTS.SetDefaultGPT(name)
	CHATGPTS.SaveChatGPTs("sessions.json")
}

func createSession() {
	choice := "unknown"
	for choice != "n" && choice != "no" && choice != "y" && choice != "yes" {
		fmt.Println("Do you want to create a session to tell what you want chatgpt to do? (yes/no)")
		choice = strings.ToLower(readStringFromStdin())
	}

	if choice == "n" || choice == "no" {
		return
	}

	CreateASession()
}

func SelectSession() {
	sessions := CHATGPTS.Sessions()
	var b strings.Builder
	for i, session := range sessions {
		b.WriteString(fmt.Sprintf("%d. %s", i+1, session))
	}

	var idx int
	for idx < 1 || idx > len(sessions) {
		fmt.Println("Please select a session to start:")
		fmt.Println(b.String())
		idx = readIntFromStdin()
	}

	ListSessions()
	CHATGPTS.SetDefaultGPT(sessions[idx-1])
	CHATGPTS.SaveChatGPTs("sessions.json")
}

func ListSessions() {
	for name, c := range CHATGPTS.Clients {
		var content string
		if c.System != nil {
			content = c.System.Content
		}

		var flag string
		if c.IsDefault {
			flag = "✓"
		}

		fmt.Printf("[%s] %s[%s]  %s\n", flag, name, c.Model, content)
	}
}

func ConfigSession() {
	sessions := CHATGPTS.Sessions()
	var b strings.Builder
	for i, session := range sessions {
		b.WriteString(fmt.Sprintf("%d. %s", i+1, session))
	}

	var idx int
	for idx < 1 || idx > len(sessions) {
		fmt.Println("Please select a session to config:")
		fmt.Println(b.String())
		idx = readIntFromStdin()
	}

	name := sessions[idx-1]
	session := CHATGPTS.Clients[name]
	var key string
	for key != "q" {
		fmt.Println("Press enter the key to config. q for quit")
		var content string
		if session.System != nil {
			content = session.System.Content
		}
		fmt.Printf("(N)ame: %s, (M)odel: %s, (K)ey: %s, (D)escription: %s", name, session.Model, session.Key, content)
		key = readStringFromStdin()

		if key == "" {
			continue
		}

		key = strings.ToLower(key)
		if key == "n" {
			var newName string
			var exist bool
			for newName == "" || exist {
				fmt.Println("Enter the new name:")
				newName = readStringFromStdin()
				_, exist = CHATGPTS.Clients[newName]
				if exist {
					fmt.Printf("%s exist. Please choose new name.\n", newName)
				}
			}

			CHATGPTS.RenameSession(name, newName)
		} else if key == "m" {
			session.Model = chooseModel()
		} else if key == "k" {
			session.Key = getOpenaiKey()
		} else if key == "d" {
			fmt.Println("What do you want chatgpt to do?")
			session.System.Content = readStringFromStdin()
		}
	}

	CHATGPTS.SaveChatGPTs("sessions.json")
}

func Showkeyword() {
	fmt.Printf("Keyword: %s, Language: %s, Path: %s\n", PARAMS.Keyword, PARAMS.KeywordLanguage, PARAMS.KeywordModel)
}

var input = io.Reader(os.Stdin)

func readStringFromStdin() string {
	reader := bufio.NewReader(input)
	choose, _ := reader.ReadBytes('\n')
	return strings.Trim(string(choose), "\n")
}

func readIntFromStdin() int {
	reader := bufio.NewReader(input)
	code, err := -1, fmt.Errorf("dummy")
	for err != nil {
		choose, _ := reader.ReadBytes('\n')
		s := strings.Trim(string(choose), "\n")
		if s == "" {
			break
		}

		code, err = strconv.Atoi(s)
		if err != nil {
			fmt.Println("Invalid choice, please try again.")
		}
	}

	return code
}
