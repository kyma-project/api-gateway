#!/bin/bash

function get_kyma_file_name () {

	local _OS_TYPE=$1
	local _OS_ARCH=$2

	[ "$_OS_TYPE" == "Linux"   ] && [ "$_OS_ARCH" == "x86_64" ] && echo "kyma-linux"     ||
	[ "$_OS_TYPE" == "Linux"   ] && [ "$_OS_ARCH" == "arm64"  ] && echo "kyma-linux-arm" ||
	[ "$_OS_TYPE" == "Windows" ] && [ "$_OS_ARCH" == "x86_64" ] && echo "kyma.exe"       ||
	[ "$_OS_TYPE" == "Windows" ] && [ "$_OS_ARCH" == "arm64"  ] && echo "kyma-arm.exe"   ||
	[ "$_OS_TYPE" == "Darwin"  ] && [ "$_OS_ARCH" == "x86_64" ] && echo "kyma-darwin"    ||
	[ "$_OS_TYPE" == "Darwin"  ] && [ "$_OS_ARCH" == "arm64"  ] && echo "kyma-darwin-arm"
}

get_kyma_file_name "$@"
