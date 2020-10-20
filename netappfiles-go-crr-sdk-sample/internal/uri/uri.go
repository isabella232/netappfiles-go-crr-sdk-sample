// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

// This package provides some functions to parse a resource
// id and return their names based on their type.
// It also validates if a resource is of an specific type based
// on provided id and finally to validate if it is an ANF related
// resource.

package uri

import (
	"fmt"
	"strings"
)

const (
	netAppResourceProviderName string = "Microsoft.NetApp"
)

// GetResourceValue returns the name of a resource from resource id/uri based on resource type name.
func GetResourceValue(resourceURI string, resourceName string) string {

	if len(strings.TrimSpace(resourceURI)) == 0 {
		return ""
	}

	if len(strings.TrimSpace(resourceName)) == 0 {
		return ""
	}

	if !strings.HasPrefix(resourceURI, "/") {
		resourceURI = fmt.Sprintf("/%v", resourceURI)
	}

	if !strings.HasPrefix(resourceName, "/") {
		resourceName = fmt.Sprintf("/%v", resourceName)
	}

	// Checks to see if the ResourceName and ResourceGroup is the same name and if so handles it specially.
	rgResourceName := fmt.Sprintf("/resourceGroups%v", resourceName)
	rgIndex := strings.Index(strings.ToLower(resourceURI), strings.ToLower(rgResourceName))

	// Dealing with case where resource name is the same as resource group
	if rgIndex > -1 {
		removedSameRgName := strings.Split(strings.ToLower(resourceURI), strings.ToLower(resourceName))
		return strings.Split(removedSameRgName[len(removedSameRgName)-1], "/")[1]
	}

	// Dealing with regular cases
	index := strings.Index(strings.ToLower(resourceURI), strings.ToLower(resourceName))
	if index > -1 {
		resource := strings.Split(resourceURI[index+len(resourceName):], "/")
		if len(resource) > 1 {
			return resource[1]
		}
	}

	return ""
}

// GetResourceName gets the resource name from resource id/uri
func GetResourceName(resourceURI string) string {

	if len(strings.TrimSpace(resourceURI)) == 0 {
		return ""
	}

	position := strings.LastIndex(resourceURI, "/")
	return resourceURI[position+1:]
}

// GetSubscription gets he subscription id from resource id/uri
func GetSubscription(resourceURI string) string {

	if len(strings.TrimSpace(resourceURI)) == 0 {
		return ""
	}

	subscriptionID := GetResourceValue(resourceURI, "/subscriptions")
	if subscriptionID == "" {
		return ""
	}

	return subscriptionID
}

// GetResourceGroup gets the resource group name from resource id/uri
func GetResourceGroup(resourceURI string) string {

	if len(strings.TrimSpace(resourceURI)) == 0 {
		return ""
	}

	resourceGroupName := GetResourceValue(resourceURI, "/resourceGroups")
	if resourceGroupName == "" {
		return ""
	}

	return resourceGroupName
}

// GetANFAccount gets an account name from resource id/uri
func GetANFAccount(resourceURI string) string {

	if len(strings.TrimSpace(resourceURI)) == 0 {
		return ""
	}

	accountName := GetResourceValue(resourceURI, "/netAppAccounts")
	if accountName == "" {
		return ""
	}

	return accountName
}

// GetANFCapacityPool gets pool name from resource id/uri
func GetANFCapacityPool(resourceURI string) string {

	if len(strings.TrimSpace(resourceURI)) == 0 {
		return ""
	}

	accountName := GetResourceValue(resourceURI, "/capacityPools")
	if accountName == "" {
		return ""
	}

	return accountName
}

// GetANFVolume gets volume name from resource id/uri
func GetANFVolume(resourceURI string) string {

	if len(strings.TrimSpace(resourceURI)) == 0 {
		return ""
	}

	volumeName := GetResourceValue(resourceURI, "/volumes")
	if volumeName == "" {
		return ""
	}

	return volumeName
}

// GetANFSnapshot gets snapshot name from resource id/uri
func GetANFSnapshot(resourceURI string) string {

	if len(strings.TrimSpace(resourceURI)) == 0 {
		return ""
	}

	snapshotName := GetResourceValue(resourceURI, "/snapshots")
	if snapshotName == "" {
		return ""
	}

	return snapshotName
}

// GetANFSnapshotPolicy gets snapshot policy name from resource id/uri
func GetANFSnapshotPolicy(resourceURI string) string {

	if len(strings.TrimSpace(resourceURI)) == 0 {
		return ""
	}

	snapshotPolicyName := GetResourceValue(resourceURI, "/snapshotPolicies")
	if snapshotPolicyName == "" {
		return ""
	}

	return snapshotPolicyName
}

// IsANFResource checks if resource is an ANF related resource
func IsANFResource(resourceURI string) bool {

	if len(strings.TrimSpace(resourceURI)) == 0 {
		return false
	}

	return strings.Index(resourceURI, netAppResourceProviderName) > -1
}

// IsANFSnapshot checks resource is a snapshot
func IsANFSnapshot(resourceURI string) bool {

	if len(strings.TrimSpace(resourceURI)) == 0 || !IsANFResource(resourceURI) {
		return false
	}

	return strings.LastIndex(resourceURI, "/snapshots/") > -1
}

// IsANFVolume checks resource is a volume
func IsANFVolume(resourceURI string) bool {

	if len(strings.TrimSpace(resourceURI)) == 0 || !IsANFResource(resourceURI) {
		return false
	}

	return !IsANFSnapshot(resourceURI) &&
		strings.LastIndex(resourceURI, "/volumes/") > -1
}

// IsANFCapacityPool checks resource is a capacity pool
func IsANFCapacityPool(resourceURI string) bool {

	if len(strings.TrimSpace(resourceURI)) == 0 || !IsANFResource(resourceURI) {
		return false
	}

	return !IsANFSnapshot(resourceURI) &&
		!IsANFVolume(resourceURI) &&
		strings.LastIndex(resourceURI, "/capacityPools/") > -1
}

// IsANFSnapshotPolicy checks resource is a snapshot policy
func IsANFSnapshotPolicy(resourceURI string) bool {

	if len(strings.TrimSpace(resourceURI)) == 0 || !IsANFResource(resourceURI) {
		return false
	}

	return !IsANFSnapshot(resourceURI) &&
		!IsANFVolume(resourceURI) &&
		!IsANFCapacityPool(resourceURI) &&
		strings.LastIndex(resourceURI, "/snapshotPolicies/") > -1
}

// IsANFAccount checks resource is an account
func IsANFAccount(resourceURI string) bool {

	if len(strings.TrimSpace(resourceURI)) == 0 || !IsANFResource(resourceURI) {
		return false
	}

	return !IsANFSnapshot(resourceURI) &&
		!IsANFVolume(resourceURI) &&
		!IsANFCapacityPool(resourceURI) &&
		!IsANFSnapshotPolicy(resourceURI) &&
		strings.LastIndex(resourceURI, "/snapshotPolicies/") == -1 &&
		strings.LastIndex(resourceURI, "/backupPolicies/") == -1 &&
		strings.LastIndex(resourceURI, "/netAppAccounts/") > -1
}
