#!/usr/bin/env bash
set -eo pipefail

# Hashes are from the table at https://golang.org/dl/
declare -r GO_VERSION=1.16.6
declare -rA GO_VERSION_HASHES=(
    [linux_amd64]=be333ef18b3016e9d7cb7b1ff1fdb0cac800ca0be4cf2290fe613b3d069dfe0d
    [linux_arm64]=9e38047463da6daecab9017cd0599f33f84991e68263752cfab49253bbc98c30
    [mac_amd64]=0b49b6cbe50b30aa0a5bb9f8ccdbb43f9cd3d9a3c36a769b8e46777d694539b5
    [windows_amd64]=c1132ba4e6263a1712355fb0745bf4f23e1602e1661c20f071e08bdcc5fe8db5
)

function install_go_linux () {
    local -r arch=$(dpkg --print-architecture)
    ARCHIVE=go${GO_VERSION}.linux-${arch}.tar.gz
    wget https://golang.org/dl/${ARCHIVE}
    echo "${GO_VERSION_HASHES[linux_${arch}]}  ${ARCHIVE}" | sha256sum --check --
    tar -C $1 -xzf ${ARCHIVE}
    rm ${ARCHIVE}
}

function install_go_mac () {
    ARCHIVE=go${GO_VERSION}.darwin-amd64.tar.gz
    wget https://golang.org/dl/${ARCHIVE}
    echo "${GO_VERSION_HASHES[mac_amd64]}  ${ARCHIVE}" | shasum -a 256 --check -
    tar -C $1 -xzf ${ARCHIVE}
    rm ${ARCHIVE}
}

function install_go_windows () {
    ARCHIVE=go${GO_VERSION}.windows-amd64.zip
    wget https://golang.org/dl/${ARCHIVE}
    echo "${GO_VERSION_HASHES[windows_amd64]}  ${ARCHIVE}" | sha256sum --check --
    unzip -qq -d $1 ${ARCHIVE}
    rm ${ARCHIVE}
}

function main () {
    if [[ $# != 1 ]]; then
        >&2 echo Usage: $0 '<install-dir>'
        exit 1
    fi
    local -r install_dir=$1

    rm -rf "$install_dir"
    mkdir -p "$install_dir"
    case $(uname) in
        Linux)
            install_go_linux "$install_dir"
            ;;
        Darwin)
            install_go_mac "$install_dir"
            ;;
        MSYS_NT*)
            install_go_windows "$install_dir"
            ;;
        *)
            >&2 echo Error: unknown OS $(uname)
            exit 1
            ;;
    esac

    "${install_dir}/go/bin/go" version
}

main ${@}
