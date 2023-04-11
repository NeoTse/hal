# HAL

**HAL** is a voice-enabled version of *chatGPT* that can run on command-line terminals, and multilingual support.

> Note: the Azure Speech SDK for go only support linux, so this project only run on linux environment. The Speech SDK does not support OpenSSL 3.0, which is the default in Ubuntu 22.04. On Ubuntu 22.04 only, install the latest libssl1.1 either as a [binary package](http://security.ubuntu.com/ubuntu/pool/main/o/openssl/), or by compiling it from sources. Read more about [this](https://learn.microsoft.com/en-us/azure/cognitive-services/speech-service/quickstarts/setup-platform?tabs=windows%2Cubuntu%2Cdotnet%2Cjre%2Cmaven%2Cnodejs%2Cmac%2Cpypi&pivots=programming-language-go).

## Prerequisites

✅ [Get an OpenAI API key](https://platform.openai.com/account/api-keys) <br/>
✅ [Get an Azure Speech key](https://learn.microsoft.com/en-us/azure/cognitive-services/cognitive-services-apis-create-account) <br/>

## Quick Start

### 1. Install

> require: go >= 1.18

```bash
git clone https://github.com/neotse/hal.git
cd hal
chmod +x install.sh
./install.sh
```

### 2. Configure

just run it, and following the guide to configure. 
```bash
hal
```

If you want to reconfigure, run below command:

```bash
hal --init
```
Or, you can see the help:

```bash
hal --help
```

### 3. Use

#### Activate and Deactivate

Try to say **Harold** to Activate, and *stopword (configured above)* to Deactivate (or automatically deactivated if there is no speech for a while). 

#### Talk

just talk to `HAL` in your language (configured above)

#### if you want custom keyword to activate

First, you need [generate a keyword model file from Azure](https://learn.microsoft.com/en-us/azure/cognitive-services/speech-service/custom-keyword-basics?pivots=programming-language-python), and download the model file. Then, use `hal keyword` command to configure it.

```bash
Usage of keyword:
  -keyword string
        set the keyword for activate (case insensitive), path and lang must be set at same time.
  -lang string
        set the language of keyword. (see https://learn.microsoft.com/en-us/azure/cognitive-services/speech-service/language-support?tabs=stt)
  -path string
        set the path of model file of keyword.
  -show
        show the current config of keyword for activate.
```

## Session

Most of the time, conversations have context. Therefore, retaining some context can improve the quality of chatGPT's responses. In addition, OpenAI also provides `system` type messages to reinforce chatGPT's attention to improve the quality of responses. Therefore, here, sessions are used to retain this information. You can manage sessions, including listing sessions, selecting sessions, editing sessions, and creating sessions. Session management can be done in two ways:

### Pre-run configuration

use `hal session` command to configure sessions.

```bash
Usage of session:
  -config
        config the 'session' for talk. 
  -create
        create the 'session' for talk. 
  -delete
        delete the 'session' for talk. 
  -list
        list current chatgpt sessions.
  -select
        select the 'session' for start to talk. If not set, it will select the session recently used.
```

### Configuration by voice

When running, you can say the keyword (or similar in meaning) to invoke corresponding the action.

| **action** | **keyword**       | **for example (other languages are also supported)**              |
|------------|-------------------|----------------------------|
| list       | list sessions     | please list the sessions   |
| select     | select session    | I want to select a session |
| delete     | delete session    | delete a session           |
| create     | create session    | help me create a session   |
| config     | configure session | configure the session      |