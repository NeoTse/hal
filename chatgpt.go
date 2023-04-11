package hal

import (
	"bufio"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"

	openai "github.com/sashabaranov/go-openai"
)

type ChatGPT struct {
	client     *openai.Client                  `json:"-"` //unnecessary serialize to json
	System     *openai.ChatCompletionMessage   `json:"system"`
	History    []*openai.ChatCompletionMessage `json:"history"`
	MaxHistory int                             `json:"maxHistory"`
	Model      string                          `json:"model"`
	Key        string                          `json:"key"`
	IsDefault  bool                            `json:"default"`
}

type streamResultCallBack func(content string)

type StreamResult struct {
	stream   *openai.ChatCompletionStream
	Err      error
	b        strings.Builder
	callback streamResultCallBack
}

func newStreamResult(stream *openai.ChatCompletionStream, callback streamResultCallBack) *StreamResult {
	return &StreamResult{
		stream:   stream,
		callback: callback,
	}
}

func (s *StreamResult) Next() string {
	if s.Err != nil {
		return ""
	}

	resp, err := s.stream.Recv()
	if err != nil {
		if !errors.Is(err, io.EOF) {
			s.callback("")
		} else {
			s.callback(s.b.String())
		}

		s.Err = err
		s.stream.Close()
		return ""
	}

	curr := resp.Choices[0].Delta.Content
	s.b.WriteString(curr)

	return curr
}

func NewChatGPT(key string, model string) *ChatGPT {
	return &ChatGPT{
		client:     openai.NewClient(key),
		Model:      model,
		MaxHistory: 4,
		Key:        key,
	}
}

func (c *ChatGPT) SetRole(text string) {
	c.System = &openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleSystem,
		Content: text,
	}
}

func (c *ChatGPT) DisableHistory() {
	c.MaxHistory = 0
}

func (c *ChatGPT) addPromptToHistory(text string) {
	if c.MaxHistory <= 0 {
		tlog.Debugf("prompt [%s] not add to history.", text)
		return
	}

	tlog.Debugf("prompt [%s] add to history.", text)
	c.History = append(c.History, &openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: text,
	})
}

func (c *ChatGPT) addResponseToHistory(text string) {
	if c.MaxHistory <= 0 {
		tlog.Debugf("response [%s] not add to history.", text)
		return
	}

	// if chatgpt with an error response ,delete the prompt from history
	if text == "" {
		if len(c.History) > 0 {
			tlog.Errorf("chatGPT with Broken response, delete the prompt from history: %s", *c.History[len(c.History)-1])
			c.History = c.History[:len(c.History)-1]
		}

		return
	}

	tlog.Debugf("response [%s] add to history.", text)
	c.History = append(c.History, &openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleAssistant,
		Content: text,
	})

	var i int
	for k := len(c.History) / 2; k > c.MaxHistory; i += 2 {
		k--
		tlog.Debugf("reach the max, delete item: [Prompt] %s, [ChatGPT] %s", c.History[i], c.History[i+1])
	}

	c.History = c.History[i:]
}

func (c *ChatGPT) Prompt(text string) (string, int, error) {
	resp, err := c.client.CreateChatCompletion(
		context.Background(),
		openai.ChatCompletionRequest{
			Model:    c.Model,
			Messages: c.buildMessages(text),
		},
	)

	if err != nil {
		return "", 0, err
	}

	c.addPromptToHistory(text)
	content := resp.Choices[0].Message.Content
	c.addResponseToHistory(content)
	return content, resp.Usage.TotalTokens, nil
}

func (c *ChatGPT) buildMessages(text string) []openai.ChatCompletionMessage {
	var res []openai.ChatCompletionMessage
	if c.System != nil {
		res = append(res, *c.System)
	}

	if len(c.History) != 0 {
		for _, h := range c.History {
			res = append(res, *h)
		}
	}

	res = append(res, openai.ChatCompletionMessage{
		Role:    openai.ChatMessageRoleUser,
		Content: text,
	})

	return res
}

func (c *ChatGPT) PromptStream(text string) (*StreamResult, error) {
	ctx := context.Background()
	req := openai.ChatCompletionRequest{
		Model:     c.Model,
		MaxTokens: 2048,
		Messages:  c.buildMessages(text),
		Stream:    true,
	}

	stream, err := c.client.CreateChatCompletionStream(ctx, req)
	if err != nil {
		return nil, err
	}

	c.addPromptToHistory(text)

	return newStreamResult(stream, c.addResponseToHistory), nil
}

func (c *ChatGPT) SetMaxHistory(n int) {
	c.MaxHistory = n
}

func (c *ChatGPT) SetModel(model string) {
	c.Model = model
}

func (c *ChatGPT) String() string {
	json, err := json.MarshalIndent(c, "", " ")
	if err != nil {
		return ""
	}

	return string(json)
}

type ChatGPTs struct {
	Clients map[string]*ChatGPT `json:"clients"`
}

func newChatGPTs() ChatGPTs {
	return ChatGPTs{Clients: map[string]*ChatGPT{}}
}

func (c ChatGPTs) NewSessionWithName(sessionName string, key string, model string) *ChatGPT {
	sessionName = strings.ToLower(sessionName)
	if client, ok := c.Clients[sessionName]; ok {
		return client
	}

	client := NewChatGPT(key, model)
	c.Clients[sessionName] = client

	return client
}

func (c ChatGPTs) NewDefaultSession(key string, model string) *ChatGPT {
	return c.NewSessionWithName("default", key, model)
}

func (c ChatGPTs) DelSession(sessionName string) {
	delete(c.Clients, sessionName)
}

func (c ChatGPTs) RenameSession(oldName string, newName string) *ChatGPT {
	if client, ok := c.Clients[oldName]; !ok {
		return client
	}

	temp := c.Clients[oldName]
	c.DelSession(oldName)
	c.Clients[newName] = temp

	return temp
}

func (c ChatGPTs) Sessions() []string {
	var res []string
	for k := range c.Clients {
		res = append(res, k)
	}

	return res
}

func (c ChatGPTs) SessionsWithout(name string) []string {
	var res []string
	for k := range c.Clients {
		if k == name {
			continue
		}

		res = append(res, k)
	}

	return res
}

func (c ChatGPTs) SetDefaultGPT(sessionName string) {
	for name, curr := range c.Clients {
		if name == sessionName {
			curr.IsDefault = true
		} else if curr.IsDefault {
			curr.IsDefault = false
		}
	}
}

func (c ChatGPTs) GetDefaultGPT() (string, *ChatGPT) {
	for name, curr := range c.Clients {
		if curr.IsDefault {
			return name, curr
		}
	}

	return "", nil
}

var CHATGPTS = newChatGPTs()

func (c ChatGPTs) SaveChatGPTs(file string) error {
	json, err := json.MarshalIndent(c, "", " ")
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

	tlog.Debugf("save chatgpts succeeded.")

	return writer.Flush()
}

func (c *ChatGPTs) LoadChatGPTs(file string) error {
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

	err = json.Unmarshal(content, c)
	if err != nil {
		return err
	}

	// create clients for each session
	for _, chatgpt := range c.Clients {
		chatgpt.client = openai.NewClient(chatgpt.Key)
	}

	tlog.Debugf("load chatgpts succeeded.")

	return nil
}

type StreamSplitter struct {
	segment strings.Builder
	stream  *StreamResult
}

func NewStreamSplitter(stream *StreamResult) StreamSplitter {
	return StreamSplitter{
		segment: strings.Builder{},
		stream:  stream,
	}
}

func (ss StreamSplitter) Segment(echo bool) string {
	var content string
	for content = ss.stream.Next(); ss.stream.Err == nil && !strings.Contains(content, "\n"); content = ss.stream.Next() {
		ss.segment.WriteString(content)
		if echo {
			fmt.Print(content)
		}
	}

	if strings.Contains(content, "\n") {
		ss.segment.WriteString(content)
		if echo {
			fmt.Print(content)
		}
	}

	res := ss.segment.String()
	ss.segment.Reset()

	return res
}
