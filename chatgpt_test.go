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

func TestPrompt(t *testing.T) {
	cg := NewChatGPT(PARAMS.OpenaiKey, PARAMS.ChatgptModel)
	cg.SetRole("你是我的翻译，负责将中文翻译为英文。")
	cg.SetMaxHistory(2)
	message, tokens, err := cg.Prompt("人工智能技术正在蓬勃发展。")
	assert.Nil(t, err)
	fmt.Println(message)
	fmt.Printf("total tokens: %d\n", tokens)

	assert.Equal(t, "你是我的翻译，负责将中文翻译为英文。", cg.System.Content)
	assert.Equal(t, 2, len(cg.History))

	message, tokens, err = cg.Prompt("特别是，自然语言处理方面，有了质的飞跃。")
	assert.Nil(t, err)
	fmt.Println(message)
	assert.Equal(t, "你是我的翻译，负责将中文翻译为英文。", cg.System.Content)
	assert.Equal(t, 4, len(cg.History))
	fmt.Printf("total tokens: %d\n", tokens)

	message, tokens, err = cg.Prompt("比如说以 chatGPT 为代表的产品。")
	assert.Nil(t, err)
	fmt.Println(message)
	assert.Equal(t, "你是我的翻译，负责将中文翻译为英文。", cg.System.Content)
	assert.Equal(t, 4, len(cg.History))
	fmt.Printf("total tokens: %d\n", tokens)
}

func TestPromptAsync(t *testing.T) {
	cg := NewChatGPT(PARAMS.OpenaiKey, PARAMS.ChatgptModel)
	cg.SetRole("你是我的助理，帮我回答问题。")
	stream, err := cg.PromptStream("目前人工智能技术发展如何？")
	assert.Nil(t, err)

	for content := stream.Next(); stream.Err == nil; {
		fmt.Print(content)
		content = stream.Next()
	}

	assert.ErrorIsf(t, stream.Err, io.EOF, "EOF")
	assert.Equal(t, "你是我的助理，帮我回答问题。", cg.System.Content)
	assert.Equal(t, 2, len(cg.History))

	stream, err = cg.PromptStream("挑其中一个分支，具体介绍一下。")
	assert.Nil(t, err)

	for content := stream.Next(); stream.Err == nil; {
		fmt.Print(content)
		content = stream.Next()
	}
	fmt.Println(stream.b.String())
	assert.ErrorIsf(t, stream.Err, io.EOF, "EOF")
	assert.Equal(t, "你是我的助理，帮我回答问题。", cg.System.Content)
	assert.Equal(t, 4, len(cg.History))
}
