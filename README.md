# HAL

**HAL** is a voice-enabled version of *chatGPT* that can run on command-line terminals, and multilingual support.

> Note: the Azure Speech SDK for go only support linux, so this project only run on linux environment. The Speech SDK does not support OpenSSL 3.0, which is the default in Ubuntu 22.04. On Ubuntu 22.04 only, install the latest libssl1.1 either as a [binary package](http://security.ubuntu.com/ubuntu/pool/main/o/openssl/), or by compiling it from sources. Read more about [this](https://learn.microsoft.com/en-us/azure/cognitive-services/speech-service/quickstarts/setup-platform?tabs=windows%2Cubuntu%2Cdotnet%2Cjre%2Cmaven%2Cnodejs%2Cmac%2Cpypi&pivots=programming-language-go).

## Prerequisites

✅ [Get an OpenAI API key](https://platform.openai.com/account/api-keys) <br/>
✅ [Get an Azure Speech key](https://learn.microsoft.com/en-us/azure/cognitive-services/cognitive-services-apis-create-account) <br/>

## Quick Start

### 1. Install

```bash
git clone https://www.github.com/neotse/hal
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
  -list
        list current chatgpt sessions.
  -select
        select the 'session' for start to talk. If not set, it will select the session recently used.
```

### Configuration by voice

When running, you can speak the keyword to invoke corresponding the configure process. keywords can be configured in `hooks.json`.

```json
{
 "hookConfigs": {
  "33d1e35c6624f0a9e71b2da03b3edd09f47d90f0": {
   "keyword": "select session.",
   "hook": "selectSession",
   "enable": true
  },
  "4ac145c3fda39261b967a576d0a6c42f700f22d6": {
   "keyword": "list session.",
   "hook": "listSession",
   "enable": true
  },
  "a24b1dcfccea7a9ca7b4c47256329a3130531745": {
   "keyword": "create session.",
   "hook": "createSession",
   "enable": true
  },
  "ac22cf93de68068ed942ea5d6d59646d62328c29": {
   "keyword": "configure session.",
   "hook": "configSession",
   "enable": true
  }
 }
}
```

There are a few things to note here:

1. the `keyword` is case insensitive
2. the `hook` is case sensitive
3. a `hook` can have multi keywords, every keyword has one configurtion item. That means multi keywords can activate this `hook`
4. a `keyword` can also have multi hooks, every keyword has one configurtion item. That means multi hooks can activated by this `keyword`
5. when the keyword is exactly equal to the text of STT (speach to text), the hook execute.