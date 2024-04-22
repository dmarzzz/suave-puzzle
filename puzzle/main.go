package main

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/flashbots/suapp-examples/framework"
)

const (
	PRIV_KEY      = "0xPK"  // FILL IN TO RUN EXAMPLE
	DesiredPrefix = "0x777" // Desired starting bytes of the address
)

func main() {
	privKey := framework.NewPrivKeyFromHex(PRIV_KEY)
	fmt.Printf("SUAVE Signer Address: %s\n", privKey.Address())

	contractAddr, ok, suaveTxHash := deploySuaveContract(privKey)
	if !ok {
		panic("Secret setting or attempt failed")
	}

	fmt.Printf("Contract deployed at: %s\n", contractAddr.Hex())
	fmt.Printf("Transaction Hash: %s\n", suaveTxHash.Hex())
}

func deploySuaveContract(privKey *framework.PrivKey) (common.Address, bool, common.Hash) {
	fr := framework.New()
	contract := fr.DeployContract("Puzzle.sol/ChillRobotPuzzle.json")
	var contractAddr *framework.Contract

	for {
		// grind for desired prefix
		contract = fr.DeployContract("Puzzle.sol/ChillRobotPuzzle.json")
		contractAddr = contract.Ref(privKey)
		if strings.HasPrefix(contractAddr.Address().Hex(), DesiredPrefix) {
			break // Desired prefix is matched
		}
		fmt.Printf("Mismatch, have: %s, want: %s\n", contractAddr.Address().Hex(), DesiredPrefix)
	}

	// TODO: if addr does not equal desired prefix deploy again
	skHex := hex.EncodeToString(crypto.FromECDSA(privKey.Priv))

	// Offchain set secret message
	teamNumber1 := big.NewInt(1)
	receipt := contractAddr.SendTransaction("offchain_setSecretMessage", []interface{}{teamNumber1, "test"}, []byte(skHex))
	if len(receipt.Logs) == 0 {
		panic("No logs from setting secret")
	}
	var secretSetEvent SecretSetEvent
	if err := secretSetEvent.Unpack(receipt.Logs[0]); err != nil {
		panic(err)
	}
	fmt.Printf("Secret was set with Data ID: %s\n", secretSetEvent.DataID.String())

	teamNumber2 := big.NewInt(2)
	receipt = contractAddr.SendTransaction("offchain_setSecretMessage", []interface{}{teamNumber2, "123test"}, []byte(skHex))
	if len(receipt.Logs) == 0 {
		panic("No logs from setting secret")
	}
	var secretSetEventTeam2 SecretSetEvent
	if err := secretSetEventTeam2.Unpack(receipt.Logs[0]); err != nil {
		panic(err)
	}
	fmt.Printf("Secret was set with Data ID: %s\n", secretSetEventTeam2.DataID.String())

	// Offchain attempt secret message
	receipt = contractAddr.SendTransaction("scrt", []interface{}{"test"}, []byte(skHex))
	if len(receipt.Logs) == 0 {
		panic("No logs from attempt")
	}
	var team1AttemptResultEvent AttemptResultEvent
	if err := team1AttemptResultEvent.Unpack(receipt.Logs[0]); err != nil {
		panic(err)
	}
	fmt.Printf("Correct Attempt result: %t\n", team1AttemptResultEvent.Success)

	// Offchain attempt secret message
	receipt = contractAddr.SendTransaction("scrt", []interface{}{"123test"}, []byte(skHex))
	if len(receipt.Logs) == 0 {
		panic("No logs from attempt")
	}
	var team2AttemptResultEvent AttemptResultEvent
	if err := team2AttemptResultEvent.Unpack(receipt.Logs[0]); err != nil {
		panic(err)
	}
	fmt.Printf("Correct Attempt result: %t\n", team2AttemptResultEvent.Success)

	// Offchain attempt secret message
	receipt = contractAddr.SendTransaction("scrt", []interface{}{"test123"}, []byte(skHex))
	if len(receipt.Logs) == 0 {
		panic("No logs from attempt")
	}
	var failResultEvent AttemptResultEvent
	if err := failResultEvent.Unpack(receipt.Logs[0]); err != nil {
		panic(err)
	}
	fmt.Printf("Wrong Attempt result: %t\n", failResultEvent.Success)

	return contractAddr.Address(), true, receipt.TxHash
}

// SecretSetEvent represents the SecretSet event.
var secretSetEventABI = `[{"anonymous":false,"inputs":[{"indexed":false,"name":"teamNumber","type":"uint256"}, {"indexed":false,"name":"dataID","type":"bytes32"}],"name":"SecretSet","type":"event"}]`

type SecretSetEvent struct {
	TeamNumber *big.Int // Use *big.Int instead of int
	DataID     common.Hash
}

func (sse *SecretSetEvent) Unpack(log *types.Log) error {
	eventABI, err := abi.JSON(strings.NewReader(secretSetEventABI))
	if err != nil {
		return err
	}
	return eventABI.UnpackIntoInterface(sse, "SecretSet", log.Data)
}

// AttemptResultEvent represents the AttemptResult event.
var attemptResultEventABI = `[{"anonymous":false,"inputs":[{"indexed":false,"name":"teamNumber","type":"uint256"},{"indexed":false,"name":"success","type":"bool"}],"name":"AttemptResult","type":"event"}]`

type AttemptResultEvent struct {
	TeamNumber *big.Int // Use *big.Int instead of int
	Success    bool
}

func (are *AttemptResultEvent) Unpack(log *types.Log) error {
	eventABI, err := abi.JSON(strings.NewReader(attemptResultEventABI))
	if err != nil {
		return err
	}
	return eventABI.UnpackIntoInterface(are, "AttemptResult", log.Data)
}
