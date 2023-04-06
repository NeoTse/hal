#!/bin/sh

set +x
set -e

SPEECHSDK_ROOT="$HOME/HAL/speechsdk"
installSpeechSDK () {
    echo "Install Azure Speech SDK into $SPEECHSDK_ROOT"
    sudo apt-get update
    sudo apt-get install build-essential libssl-dev libasound2 wget -y

    mkdir -p "$SPEECHSDK_ROOT"
    wget -O SpeechSDK-Linux.tar.gz https://aka.ms/csspeech/linuxbinary
    tar --strip 1 -xzf SpeechSDK-Linux.tar.gz -C "$SPEECHSDK_ROOT"
    
    arch=$(uname -m)
    if [ $arch = "aarch64" ]; then
        arch="arm64"
    elif [ $arch = "armv7l" ] || [ $arch = "armv6l" ]; then
        arch="arm32"
    elif [ $arch = "x86_64" ]; then
        arch="x64"
    elif [ $arch = "x86" ] || [ $arch = "i686" ]; then
        arch="x86"
    else
        echo "Unsupport platform."
        rm -f SpeechSDK-Linux.tar.gz
        rm -rf $SPEECHSDK_ROOT
        return 1
    fi

    echo >> $HOME/.profile
    printf "SPEECHSDK_ROOT=\"\$HOME/HAL/speechsdk\"\n"  >> $HOME/.profile
    printf "export CGO_CFLAGS=\"-I\$SPEECHSDK_ROOT/include/c_api\"\n"  >> $HOME/.profile
    printf "export CGO_LDFLAGS=\"-L\$SPEECHSDK_ROOT/lib/%s -lMicrosoft.CognitiveServices.Speech.core\"\n" $arch >> $HOME/.profile
    printf "export LD_LIBRARY_PATH=\"\$SPEECHSDK_ROOT/lib/%s:\$LD_LIBRARY_PATH\"\n" $arch >> $HOME/.profile
    . $HOME/.profile
    rm -f SpeechSDK-Linux.tar.gz
    echo "Install Azure Speech SDK Succeeded."
}

HAL_ROOT="$HOME/HAL/go"
installHAL () {
    echo "Install HAL into ${HAL_ROOT}"
    go build cli/hal.go
    if [ -d $HAL_ROOT ]; then
        echo "old HAL found, and remove it."
        rm -rf $HAL_ROOT
    fi

    mkdir -p ${HAL_ROOT}
    cp hal ${HAL_ROOT}
    cp params.json hooks.json ${HAL_ROOT}
    cp -R ./model ${HAL_ROOT}


    existExport=$(grep HAL_ROOT $HOME/.profile)
    if [ ${#existExport} -eq 0 ]; then
        echo >> $HOME/.profile
        printf "HAL_ROOT=\"\$HOME/HAL/go\"\n" >> $HOME/.profile
        printf "export PATH=\"\$HAL_ROOT\":\$PATH\n" >> $HOME/.profile
        . $HOME/.profile
    fi
    
    echo "Install HAL Succeeded."
}


[ ! -d "$SPEECHSDK_ROOT" ] && installSpeechSDK
[ $? -eq 0 ] && installHAL