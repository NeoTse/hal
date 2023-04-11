package main

import (
	"testing"

	hal "github.com/neotse/hal"
	"github.com/stretchr/testify/assert"
)

func init() {
	hal.PARAMS.LoadParams("../params.json")
	hal.CHATGPTS.LoadChatGPTs("../sessions.json")
	hal.HOOKS.LoadHooks("../hooks.json")
}

func TestHookChatGPT(t *testing.T) {
	assert.NotPanics(t, func() { initHooksChatGPT() })
	hooks := hal.CHATGPTS.Clients["hooks"]

	assert.NotNil(t, hooks)
	assert.Equal(t, 6, len(hooks.History))

	res := hook("please end a session")
	assert.False(t, res)
	assert.Equal(t, 6, len(hooks.History))
	res = hook("i want to list the sessions")
	assert.True(t, res)
	res = hook("列出当前所有会话")
	assert.True(t, res)
}
