package hal

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

var params Params

func init() {
	params = Params{MaxHistory: 4, Language: "en-US", Voice: "en-US-ElizabethNeural"}
}

func TestSaveParams(t *testing.T) {
	f, err := os.CreateTemp("./test_data", "params*.json")
	assert.Nil(t, err)

	defer f.Close()
	defer os.Remove(f.Name())
	err = params.SaveParams(f.Name())
	assert.Nil(t, err)
}

func TestLoadParams(t *testing.T) {
	f, err := os.CreateTemp("./test_data", "params*.json")
	assert.Nil(t, err)

	defer f.Close()
	defer os.Remove(f.Name())
	err = params.SaveParams(f.Name())
	assert.Nil(t, err)

	cParams := Params{MaxHistory: 4, Language: "en-US", Voice: "en-US-ElizabethNeural"}
	err = params.LoadParams(f.Name())
	assert.Nil(t, err)

	assert.Equal(t, cParams, params)
}
