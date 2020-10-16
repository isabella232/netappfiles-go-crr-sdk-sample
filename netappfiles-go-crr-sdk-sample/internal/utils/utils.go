// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

// Package that provides some general functions.

package utils

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"strings"
	"syscall"

	"github.com/Azure-Samples/netappfiles-go-crr-sdk-sample/netappfiles-go-crr-sdk-sample/internal/models"
	"golang.org/x/crypto/ssh/terminal"
)

// PrintHeader prints a header message
func PrintHeader(header string) {
	fmt.Println(header)
	fmt.Println(strings.Repeat("-", len(header)))
}

// ConsoleOutput writes to stdout.
func ConsoleOutput(message string) {
	log.Println(message)
}

// Contains checks if there is a string already in an existing splice of strings
func Contains(array []string, element string) bool {
	for _, e := range array {
		if e == element {
			return true
		}
	}
	return false
}

// GetBytesInTiB converts a value from bytes to tebibytes (TiB)
func GetBytesInTiB(size uint64) uint32 {
	return uint32(size / 1024 / 1024 / 1024 / 1024)
}

// GetTiBInBytes converts a value from tebibytes (TiB) to bytes
func GetTiBInBytes(size uint32) uint64 {
	return uint64(size * 1024 * 1024 * 1024 * 1024)
}

// ReadAzureBasicInfoJSON reads the Azure Authentication json file json file and unmarshals it.
func ReadAzureBasicInfoJSON(path string) (*models.AzureBasicInfo, error) {
	infoJSON, err := ioutil.ReadFile(path)
	if err != nil {
		fmt.Sprintf("failed to read file: %v", err)
		return &models.AzureBasicInfo{}, err
	}
	var info models.AzureBasicInfo
	json.Unmarshal(infoJSON, &info)
	return &info, nil
}

// FindInSlice returns index greater than -1 and true if item is found
// Code from https://golangcode.com/check-if-element-exists-in-slice/
func FindInSlice(slice []string, val string) (int, bool) {
	for i, item := range slice {
		if item == val {
			return i, true
		}
	}
	return -1, false
}

// GetPassword gets a password
func GetPassword(prompt string) string {
	fmt.Print(prompt)
	bytePassword, _ := terminal.ReadPassword(int(syscall.Stdin))
	fmt.Println()
	return strings.TrimSpace(string(bytePassword))
}
