package hal

import (
	"fmt"
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var testHooks = newHooks()

func init() {
	temp1 := &createSessionHook{}
	testHooks.registerHookInstance(temp1.Name(), temp1)

	temp2 := &listSessionHook{}
	testHooks.registerHookInstance(temp2.Name(), temp2)

	temp3 := &selectSessionHook{}
	testHooks.registerHookInstance(temp3.Name(), temp3)

	temp4 := &configSessionHook{}
	testHooks.registerHookInstance(temp4.Name(), temp4)

	testHooks.Add("create session", "createSession")
	testHooks.Add("list session", "listSession")
	testHooks.Add("select session", "selectSession")
	testHooks.Add("configure session", "configSession")
}

func TestGetHook(t *testing.T) {
	id := hookId("create session", "createSession")
	config := testHooks.Get(id)
	expected := &HookConfig{
		id:       id,
		Keyword:  "create session",
		HookName: "createSession",
		Enable:   true,
		instance: &createSessionHook{},
	}

	assert.Equal(t, expected.id, config.id)
	assert.Equal(t, expected.Keyword, config.Keyword)
	assert.Equal(t, expected.HookName, config.HookName)
	assert.Equal(t, expected.Enable, config.Enable)
	assert.NotEqual(t, expected.instance, config.instance)
}

func TestAddHook(t *testing.T) {
	assert.ErrorIs(t, HOOKS.Add("configure session", "configSession"), ErrRepeatConfig)
	assert.Nil(t, HOOKS.Add("configure Session", "configSession"))
}

func TestDeleteHook(t *testing.T) {
	testHooks.Add("test", "configSession")
	id := hookId("test", "configSession")
	assert.True(t, testHooks.Delete(id))

	id = hookId("Test", "configSession")
	assert.False(t, testHooks.Delete(id))
}

func TestDisableHook(t *testing.T) {
	id := hookId("list session", "listSession")
	assert.True(t, testHooks.Get(id).Enable)
	testHooks.Disable(id)
	assert.False(t, testHooks.Get(id).Enable)
	testHooks.Enable(id)
	assert.True(t, testHooks.Get(id).Enable)
}

func TestSaveHooks(t *testing.T) {
	f, err := os.CreateTemp("./test_data", "hooks*.json")
	assert.Nil(t, err)

	defer f.Close()
	defer os.Remove(f.Name())

	err = testHooks.SaveHooks(f.Name())
	assert.Nil(t, err)
}

func TestLoadHooks(t *testing.T) {
	f, err := os.CreateTemp("./test_data", "hooks*.json")
	assert.Nil(t, err)

	defer f.Close()
	defer os.Remove(f.Name())

	err = testHooks.SaveHooks(f.Name())
	assert.Nil(t, err)

	temp := newHooks()

	err = temp.LoadHooks(f.Name())
	fmt.Println(temp)
	assert.Nil(t, err)

	assert.Equal(t, testHooks, temp)
}
