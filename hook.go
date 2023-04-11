package hal

import (
	"bufio"
	"crypto/sha1"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
)

type Hook interface {
	New() Hook
	Check(stt string) bool
	SetKeyword(text string)
	Name() string
	Exec() error
}

type HookConfig struct {
	id       string `json:"-"`
	Keyword  string `json:"keyword"`
	HookName string `json:"hook"`
	Enable   bool   `json:"enable"`
	instance Hook   `json:"-"`
}

var (
	ErrRepeatConfig = errors.New("found a same config")
	ErrNoSuchHook   = errors.New("hook not exists")
)

type Hooks struct {
	Configs   map[string]*HookConfig `json:"hookConfigs"`
	instances map[string]Hook        `json:"-"`
}

func newHooks() *Hooks {
	return &Hooks{
		Configs:   make(map[string]*HookConfig),
		instances: make(map[string]Hook),
	}
}

func (h *Hooks) IsExist(hookName string) bool {
	if _, ok := h.instances[hookName]; ok {
		return true
	}

	return false
}

func (h *Hooks) Exec(hookName string) error {
	hook := h.instances[hookName]
	if hook == nil {
		return ErrNoSuchHook
	}

	return hook.Exec()
}

func (h *Hooks) Get(id string) *HookConfig {
	return h.Configs[id]
}

func (h *Hooks) registerHookInstance(name string, instance Hook) bool {
	if _, ok := h.instances[name]; ok {
		return false
	}

	h.instances[name] = instance
	return true
}

func (h *Hooks) Add(keyword string, name string) error {
	id := hookId(keyword, name)
	if _, ok := h.Configs[id]; ok {
		return ErrRepeatConfig
	}

	config := &HookConfig{Keyword: keyword, HookName: name, Enable: true}
	config.id = id
	config.instance = h.instances[name].New()
	config.instance.SetKeyword(keyword)
	h.Configs[id] = config

	return nil
}

func (h *Hooks) Delete(id string) bool {
	if _, ok := h.Configs[id]; !ok {
		return false
	}

	delete(h.Configs, id)
	return true
}

func (h *Hooks) Disable(id string) {
	if config, ok := h.Configs[id]; ok {
		config.Enable = false
	}
}

func (h *Hooks) Enable(id string) {
	if config, ok := h.Configs[id]; ok {
		config.Enable = true
	}
}

func hookId(keyword string, name string) string {
	sha1 := sha1.New()
	io.WriteString(sha1, keyword)
	io.WriteString(sha1, "\t")
	io.WriteString(sha1, name)

	return fmt.Sprintf("%x", sha1.Sum(nil))
}

var HOOKS = newHooks()

func (h *Hooks) LoadHooks(file string) error {
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

	err = json.Unmarshal(content, h)
	if err != nil {
		return err
	}

	for _, config := range h.Configs {
		config.id = hookId(config.Keyword, config.HookName)
		config.instance = h.instances[config.HookName].New()
		config.instance.SetKeyword(config.Keyword)
	}

	tlog.Debugf("load hooks succeeded.")
	return nil
}

func (h *Hooks) SaveHooks(file string) error {
	json, err := json.MarshalIndent(h, "", " ")
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

	tlog.Debugf("save hooks succeeded.")

	return writer.Flush()
}
