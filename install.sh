#!/bin/bash

set +x
set -e

SPEECHSDK_ROOT="$HOME/HAL/speechsdk"
function installSpeechSDK() {
    echo "Install Azure Speech SDK into $SPEECHSDK_ROOT"
    sudo apt-get update
    sudo apt-get install build-essential libssl-dev libasound2 wget

    
    mkdir -p "$SPEECHSDK_ROOT"
    wget -O SpeechSDK-Linux.tar.gz https://aka.ms/csspeech/linuxbinary
    tar --strip 1 -xzf SpeechSDK-Linux.tar.gz -C "$SPEECHSDK_ROOT"
    
    arch=$(uname -m)
    if [ "$arch" == "aarch64" ]; then
        arch="arm64"
    elif [ "$arch" == "armv7l" ] || [ "$arch" == "armv6l" ]; then
        arch="arm32"
    elif [ "$arch" == "x86_64" ]; then
        arch="x64"
    elif [ "$arch" == "x86" ] || [ "$arch" == "i686" ]; then
        arch="x86"
    else
        echo "Unsupport platform."
        exit 1
    fi

    echo -ne 'export CGO_CFLAGS="-I$SPEECHSDK_ROOT/include/c_api"\n' >> $HOME/.profile
    export CGO_CFLAGS="-I$SPEECHSDK_ROOT/include/c_api"

    echo -ne 'CGO_LDFLAGS="-L$SPEECHSDK_ROOT/lib/$arch -lMicrosoft.CognitiveServices.Speech.core"\n' >> $HOME/.profile
    export CGO_LDFLAGS="-L$SPEECHSDK_ROOT/lib/$arch -lMicrosoft.CognitiveServices.Speech.core"
    
    echo -ne 'LD_LIBRARY_PATH="$SPEECHSDK_ROOT/lib/$arch:$LD_LIBRARY_PATH"' >> $HOME/.profile
    export LD_LIBRARY_PATH="$SPEECHSDK_ROOT/lib/$arch:$LD_LIBRARY_PATH"

    rm -f SpeechSDK-Linux.tar.gz
    echo "Install Azure Speech SDK Succeeded。"
}


func installHAL() {
    echo "Install HAL into ${install_path}"
    git clone https://www.github.com/neotse/hal
    cd hal
    go build
    if [ -d $install_path ]; then
        echo "remove old HAL"
        rm -rf $install_path
    fi

    mkdir ${install_path}
    cp hal ${install_path}
    cp params.json sessions.json hooks.json ${install_path}
    cp -R ./model ${install_path}


    echo -ne 'PATH="$HAL_ROOT":$PATH' >> $HOME/.profile
    export PATH="$HAL_ROOT":$PATH
    echo "Install HAL Succeeded."
}


[ ! -d "SPEECHSDK_ROOT" ] && installSpeechSDK
installHAL