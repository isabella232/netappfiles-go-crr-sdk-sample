// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

// This sample code creates an Azure Netapp Files Account, a Capacity Pool,
// and two volumes, one NFSv3 and one NFSv4.1, then it takes a snapshot
// of the first volume (NFSv3) and performs clean up if the variable
// shouldCleanUp is changed to true.
//
// This package uses go-haikunator package (https://github.com/yelinaung/go-haikunator)
// port from Python's haikunator module and therefore used here just for sample simplification,
// this doesn't mean that it is endorsed/thouroughly tested by any means, use at own risk.
// Feel free to provide your own names on variables using it.

package main

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/Azure-Samples/netappfiles-go-crr-sdk-sample/netappfiles-go-crr-sdk-sample/internal/sdkutils"
	"github.com/Azure-Samples/netappfiles-go-crr-sdk-sample/netappfiles-go-crr-sdk-sample/internal/utils"
	"github.com/Azure/azure-sdk-for-go/profiles/latest/netapp/mgmt/netapp"
	"github.com/Azure/go-autorest/autorest/to"
)

const (
	virtualNetworksApiVersion string = "2019-09-01"
)

type (
	// Properties - properties to be used when defining primary and secondary anf resources
	Properties struct {
		Location              string
		ResourceGroupName     string
		VnetResourceGroupName string
		VnetName              string
		SubnetName            string
		AnfAccountName        string
		CapacityPoolName      string
		VolumeName            string
		ServiceLevel          string // Valid service levels are Standard, Premium and Ultra
		VolumeID              string // This will be populated after resource is created
		CapacityPoolID        string // This will be populated after resource is created
		AcccountID            string // This will be populated after resource is created
	}
)

var (
	shouldCleanUp bool = false

	// Important - change ANF related variables below to appropriate values related to your environment
	// Share ANF properties related
	capacityPoolSizeBytes int64 = 4398046511104 // 4TiB (minimum capacity pool size)
	volumeSizeBytes       int64 = 107374182400  // 100GiB (minimum volume size)
	protocolTypes               = []string{"NFSv3"}
	sampleTags                  = map[string]*string{
		"Author":  to.StringPtr("ANF Go CRR SDK Sample"),
		"Service": to.StringPtr("Azure Netapp Files"),
	}

	// ANF Resource Properties
	anfResources = map[string]*Properties{
		"Primary": &Properties{
			Location:              "westus",
			ResourceGroupName:     "anf-primary-rg",
			VnetResourceGroupName: "anf-primary-rg",
			VnetName:              "westus-primary-vnet",
			SubnetName:            "anf-primary-sn",
			AnfAccountName:        "PrimaryANFAccount",
			CapacityPoolName:      "PrimaryPool",
			ServiceLevel:          "Premium",
			VolumeName:            "PrimaryVolume",
		},
		"Secondary": &Properties{
			Location:              "eastus",
			ResourceGroupName:     "anf-secondary-rg",
			VnetResourceGroupName: "anf-secondary-rg",
			VnetName:              "eastus-secondary-vnet",
			SubnetName:            "anf-secondary-sn",
			AnfAccountName:        "SecondaryANFAccount",
			CapacityPoolName:      "SecondaryPool",
			ServiceLevel:          "Standard",
			VolumeName:            "SecondaryVolume",
		},
	}

	// Some other variables used throughout the course of the code execution - no need to change it
	exitCode int
)

func main() {

	cntx := context.Background()

	// Cleanup and exit handling
	defer func() { exit(cntx); os.Exit(exitCode) }()

	utils.PrintHeader("Azure NetAppFiles Go CRR SDK Sample - Sample application that enables cross-region replication on an NFSv3 volume.")

	// Getting subscription ID from authentication file
	config, err := utils.ReadAzureBasicInfoJSON(os.Getenv("AZURE_AUTH_LOCATION"))
	if err != nil {
		utils.ConsoleOutput(fmt.Sprintf("an error ocurred getting non-sensitive info from AzureAuthFile: %v", err))
		exitCode = 1
		shouldCleanUp = false
		return
	}

	// Primary and Secondary ANF operations
	sideIndex := []string{"Primary", "Secondary"}
	for _, side := range sideIndex {
		utils.ConsoleOutput(fmt.Sprintf("Working on %v ANF Resources...", side))

		// Checking if subnet exists before any other operation starts
		subnetID := fmt.Sprintf("/subscriptions/%v/resourceGroups/%v/providers/Microsoft.Network/virtualNetworks/%v/subnets/%v",
			*config.SubscriptionID,
			anfResources[side].VnetResourceGroupName,
			anfResources[side].VnetName,
			anfResources[side].SubnetName,
		)

		utils.ConsoleOutput(fmt.Sprintf("Checking if vnet/subnet %v exists.", subnetID))

		_, err = sdkutils.GetResourceByID(cntx, subnetID, virtualNetworksApiVersion)
		if err != nil {
			if string(err.Error()) == "NotFound" {
				utils.ConsoleOutput(fmt.Sprintf("error: %v subnet %v not found: %v", side, subnetID, err))
			} else {
				utils.ConsoleOutput(fmt.Sprintf("error: an error ocurred trying to check if %v %v subnet exists: %v", side, subnetID, err))
			}
			exitCode = 1
			shouldCleanUp = false
			return
		}

		// Account creation
		utils.ConsoleOutput(fmt.Sprintf("Creating %v Azure NetApp Files account...", side))

		account, err := sdkutils.CreateAnfAccount(cntx, anfResources[side].Location, anfResources[side].ResourceGroupName, anfResources[side].AnfAccountName, nil, sampleTags)
		if err != nil {
			utils.ConsoleOutput(fmt.Sprintf("an error ocurred while creating account: %v", err))
			exitCode = 1
			shouldCleanUp = false
			return
		}
		anfResources[side].AcccountID = *account.ID
		utils.ConsoleOutput(fmt.Sprintf("Account successfully created, resource id: %v", anfResources[side].AcccountID))

		// Capacity pool creation
		utils.ConsoleOutput(fmt.Sprintf("Creating %v Capacity Pool...", side))
		capacityPool, err := sdkutils.CreateAnfCapacityPool(
			cntx,
			anfResources[side].Location,
			anfResources[side].ResourceGroupName,
			anfResources[side].AnfAccountName,
			anfResources[side].CapacityPoolName,
			anfResources[side].ServiceLevel,
			capacityPoolSizeBytes,
			sampleTags,
		)
		if err != nil {
			utils.ConsoleOutput(fmt.Sprintf("an error ocurred while creating %v capacity pool: %v", side, err))
			exitCode = 1
			shouldCleanUp = false
			return
		}
		anfResources[side].CapacityPoolID = *capacityPool.ID
		utils.ConsoleOutput(fmt.Sprintf("Capacity Pool successfully created, resource id: %v", anfResources[side].CapacityPoolID))

		// Volume creation
		utils.ConsoleOutput(fmt.Sprintf("Creating %v NFSv3 Volume...", side))

		// Build data protection object if Secondary side.
		dataProtectionObject := netapp.VolumePropertiesDataProtection{}
		if side == "Secondary" {
			utils.ConsoleOutput(fmt.Sprintf("\tCreating data protection object since this is %v volume...", side))
			utils.ConsoleOutput(fmt.Sprintf("\tRemote volume id is %v...", anfResources["Primary"].VolumeID))
			replicationObject := netapp.ReplicationObject{
				EndpointType:           "dst",
				RemoteVolumeRegion:     to.StringPtr(anfResources["Primary"].Location),
				RemoteVolumeResourceID: to.StringPtr(anfResources["Primary"].VolumeID),
				ReplicationSchedule:    "hourly",
			}

			dataProtectionObject = netapp.VolumePropertiesDataProtection{
				Replication: &replicationObject,
			}
		}

		volume, err := sdkutils.CreateAnfVolume(
			cntx,
			anfResources[side].Location,
			anfResources[side].ResourceGroupName,
			anfResources[side].AnfAccountName,
			anfResources[side].CapacityPoolName,
			anfResources[side].VolumeName,
			anfResources[side].ServiceLevel,
			subnetID,
			"",
			protocolTypes,
			volumeSizeBytes,
			false,
			true,
			sampleTags,
			dataProtectionObject,
		)

		if err != nil {
			utils.ConsoleOutput(fmt.Sprintf("an error ocurred while creating %v volume: %v", side, err))
			exitCode = 1
			shouldCleanUp = false
			return
		}

		anfResources[side].VolumeID = *volume.ID
		utils.ConsoleOutput(fmt.Sprintf("Volume successfully created, resource id: %v", anfResources[side].VolumeID))

		utils.ConsoleOutput("Waiting for volume to be ready...")
		err = sdkutils.WaitForANFResource(cntx, anfResources[side].VolumeID, 60, 50, false)
		if err != nil {
			utils.ConsoleOutput(fmt.Sprintf("an error ocurred while waiting for %v volume: %v", side, err))
			exitCode = 1
			shouldCleanUp = false
			return
		}
	}

	// Authorizing replication
	utils.ConsoleOutput("Authorizing replication...")
	err = sdkutils.AuthorizeReplication(
		cntx,
		anfResources["Primary"].ResourceGroupName,
		anfResources["Primary"].AnfAccountName,
		anfResources["Primary"].CapacityPoolName,
		anfResources["Primary"].VolumeName,
		anfResources["Secondary"].VolumeID,
	)
	if err != nil {
		utils.ConsoleOutput(fmt.Sprintf("an error ocurred while authorizing replication: %v", err))
		exitCode = 1
		shouldCleanUp = false
		return
	}

	utils.ConsoleOutput("Waiting for primary volume replication be ready...")
	err = sdkutils.WaitForANFResource(cntx, anfResources["Primary"].VolumeID, 60, 50, true)
	if err != nil {
		utils.ConsoleOutput(fmt.Sprintf("an error ocurred while waiting for Primary volume be replication ready: %v", err))
		exitCode = 1
		shouldCleanUp = false
		return
	}
}

func exit(cntx context.Context) {
	utils.ConsoleOutput("Exiting")

	if shouldCleanUp {
		utils.ConsoleOutput("\tPerforming clean up")

		// Clean up must be executed in reverse order, mainly because replication must be deleted on secondary volume first
		sideIndex := []string{"Secondary", "Primary"}

		for _, side := range sideIndex {
			resourceGroupName := anfResources[side].ResourceGroupName
			accountName := anfResources[side].AnfAccountName
			poolName := anfResources[side].CapacityPoolName
			volumeName := anfResources[side].VolumeName

			// Delete replication
			utils.ConsoleOutput(fmt.Sprintf("\tRemoving data protection object from %v volume...", anfResources[side].VolumeName))
			err := sdkutils.DeleteAnfVolumeReplication(
				cntx,
				resourceGroupName,
				accountName,
				poolName,
				volumeName,
			)
			if err != nil && !strings.Contains(err.Error(), "VolumeReplicationMissing") {
				utils.ConsoleOutput(fmt.Sprintf("an error ocurred while deleting data replication: %v", err))
				exitCode = 1
				return
			}
			sdkutils.WaitForNoANFResource(cntx, anfResources[side].VolumeID, 60, 60, true)
			utils.ConsoleOutput("\tData replication successfully deleted")

			// Volume deletion
			utils.ConsoleOutput(fmt.Sprintf("\tRemoving %v volume...", anfResources[side].VolumeID))
			err = sdkutils.DeleteAnfVolume(
				cntx,
				resourceGroupName,
				accountName,
				poolName,
				volumeName,
			)
			if err != nil {
				utils.ConsoleOutput(fmt.Sprintf("an error ocurred while deleting volume: %v", err))
				exitCode = 1
				return
			}
			sdkutils.WaitForNoANFResource(cntx, anfResources[side].VolumeID, 60, 60, false)
			utils.ConsoleOutput("\tVolume successfully deleted")

			// Pool Cleanup
			utils.ConsoleOutput(fmt.Sprintf("\tCleaning up capacity pool %v...", anfResources[side].CapacityPoolID))
			err = sdkutils.DeleteAnfCapacityPool(
				cntx,
				resourceGroupName,
				accountName,
				poolName,
			)
			if err != nil {
				utils.ConsoleOutput(fmt.Sprintf("an error ocurred while deleting capacity pool: %v", err))
				exitCode = 1
				return
			}
			sdkutils.WaitForNoANFResource(cntx, anfResources[side].CapacityPoolID, 60, 60, false)
			utils.ConsoleOutput("\tCapacity pool successfully deleted")

			// Account Cleanup
			utils.ConsoleOutput(fmt.Sprintf("\tCleaning up account %v...", anfResources[side].AcccountID))
			err = sdkutils.DeleteAnfAccount(
				cntx,
				resourceGroupName,
				accountName,
			)
			if err != nil {
				utils.ConsoleOutput(fmt.Sprintf("an error ocurred while deleting account: %v", err))
				exitCode = 1
				return
			}
			utils.ConsoleOutput("\tAccount successfully deleted")
		}
		utils.ConsoleOutput("\tCleanup completed!")
	}
}
