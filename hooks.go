package hal

import (
	"strings"
)

func init() {
	temp1 := &createSessionHook{}
	HOOKS.registerHookInstance(temp1.Name(), temp1)

	temp2 := &listSessionHook{}
	HOOKS.registerHookInstance(temp2.Name(), temp2)

	temp3 := &selectSessionHook{}
	HOOKS.registerHookInstance(temp3.Name(), temp3)

	temp4 := &configSessionHook{}
	HOOKS.registerHookInstance(temp4.Name(), temp4)
}

type defaultHook struct {
	keyword string
	name    string
}

func (h *defaultHook) Check(s string) bool {
	return h.keyword == strings.ToLower(s)
}

func (h *defaultHook) SetKeyword(keyword string) {
	h.keyword = keyword
}

type createSessionHook struct {
	defaultHook
}

func (h *createSessionHook) New() Hook {
	res := &createSessionHook{}
	res.name = h.name
	res.keyword = h.keyword

	return res
}

func (h *createSessionHook) Name() string {
	if h.name == "" {
		h.name = "createSession"
	}
	return h.name
}

func (h *createSessionHook) Exec() error {
	CreateASession()
	return nil
}

type listSessionHook struct {
	defaultHook
}

func (h *listSessionHook) New() Hook {
	res := &listSessionHook{}
	res.name = h.name
	res.keyword = h.keyword

	return res
}

func (h *listSessionHook) Name() string {
	if h.name == "" {
		h.name = "listSession"
	}
	return h.name
}

func (h *listSessionHook) Exec() error {
	ListSessions()
	return nil
}

type selectSessionHook struct {
	defaultHook
}

func (h *selectSessionHook) New() Hook {
	res := &selectSessionHook{}
	res.name = h.name
	res.keyword = h.keyword

	return res
}

func (h *selectSessionHook) Name() string {
	if h.name == "" {
		h.name = "selectSession"
	}
	return h.name
}

func (h *selectSessionHook) Exec() error {
	SelectSession()
	return nil
}

type configSessionHook struct {
	defaultHook
}

func (h *configSessionHook) New() Hook {
	res := &configSessionHook{}
	res.name = h.name
	res.keyword = h.keyword

	return res
}

func (h *configSessionHook) Name() string {
	if h.name == "" {
		h.name = "configSession"
	}
	return h.name
}

func (h *configSessionHook) Exec() error {
	ConfigSession()
	return nil
}
