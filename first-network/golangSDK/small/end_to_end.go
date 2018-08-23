/*
Copyright SecureKey Technologies Inc. All Rights Reserved.

SPDX-License-Identifier: Apache-2.0
*/

package main

import (
	"fmt"

	"github.com/hyperledger/fabric-sdk-go/pkg/client/channel"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/errors/retry"
	"github.com/hyperledger/fabric-sdk-go/pkg/common/providers/core"
	"github.com/hyperledger/fabric-sdk-go/pkg/core/config"
	"github.com/hyperledger/fabric-sdk-go/pkg/fabsdk"
)

const (
	channelID = "chpatinets"
	//channelID      = "mychannel"
	orgName        = "org1"
	orgAdmin       = "Admin"
	ordererOrgName = "OrdererOrg"
	ccID           = "patientssys"
)

func queryCC(client *channel.Client) []byte {
	var queryArgs = [][]byte{[]byte("getPatientbyID"), []byte("1")}

	response, err := client.Query(channel.Request{ChaincodeID: ccID, Fcn: "query", Args: queryArgs},
		channel.WithRetry(retry.DefaultChannelOpts),
	)
	if err != nil {
		fmt.Printf("Failed to query funds: %s\n", err)
		return nil
	}
	fmt.Println("Quary done !!!")
	return response.Payload
}

// setupAndRun enables testing an end-to-end scenario against the supplied SDK options
// the createChannel flag will be used to either create a channel and the example CC or not(ie run the tests with existing ch and CC)
func setupAndRun(configOpt core.ConfigProvider, sdkOpts ...fabsdk.Option) {

	sdk, err := fabsdk.New(configOpt, sdkOpts...)
	if err != nil {
		fmt.Printf("Failed to create new SDK: %s\n", err)
		return
	}
	defer sdk.Close()

	fmt.Println("=== Connection done .")

	//prepare channel client context using client context
	clientChannelContext := sdk.ChannelContext(channelID, fabsdk.WithUser("Admin"), fabsdk.WithOrg(orgName))
	// Channel client is used to query and execute transactions (Org1 is default org)
	client, err := channel.New(clientChannelContext)
	if err != nil {
		fmt.Printf("Failed to get client: %s\n", err)
		return
	}

	res := queryCC(client)
	fmt.Println(res)
}

func main() {
	//configOpt := config.FromFile("configtx.yaml")
	// var client = Client.loadFromConfig("config.yaml")

	configOpt := config.FromFile("config.yaml")
	setupAndRun(configOpt)
}

// UTC - peer.(*peerEndorser).sendProposal -> ERRO process proposal failed [rpc error: code = Unknown desc = access denied: channel [ ] creator org [org1.example.com]]
