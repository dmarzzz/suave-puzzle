// SPDX-License-Identifier: UNLICENSED
pragma solidity ^0.8.13;

import "../../../../lib/suave-std/src/suavelib/Suave.sol";

contract ChillRobotPuzzle {
    string private constant SECRET_KEY_BASE = "PuzzleSecret:v";
    address public owner;

    mapping(uint256 => Suave.DataId) private secretMessageDataIDs;
    mapping(uint256 => bool) private isSecretSet; // New mapping to track if a secret is set for a team

    event SecretSet(uint256 teamNumber, Suave.DataId dataID);
    event AttemptResult(uint256 teamNumber, bool success);
    event AllAttemptsResult(bool success, uint256 successfulTeam);

    constructor() {
        owner = msg.sender;
    }

    // Set a secret message for a specific team
    function offchain_setSecretMessage(uint256 teamNumber, string memory secret) external returns (bytes memory) {
        // TODO : use confidential input!
        require(msg.sender == owner, "Only the owner can set the secret message.");
        require(teamNumber >= 1 && teamNumber <= 5, "Team number must be between 1 and 5.");

        bytes memory secretBytes = bytes(secret);
        address[] memory peekers = new address[](1);
        peekers[0] = address(this);
        string memory key = string(abi.encodePacked(SECRET_KEY_BASE, uint2str(teamNumber)));
        address[] memory allowedStores = new address[](0);
        Suave.DataRecord memory record = Suave.newDataRecord(0, peekers, allowedStores, key);

        Suave.confidentialStore(record.id, key, secretBytes);
        // secretMessageDataIDs[teamNumber] = record.id;
        // isSecretSet[teamNumber] = true; // Mark this team's secret as set

        emit SecretSet(teamNumber, record.id);
        return bytes.concat(this.onchain_setSecretMessage.selector, abi.encode(teamNumber, record.id));
    }

    function onchain_setSecretMessage(uint256 teamNumber, Suave.DataId dataID) public {
        require(msg.sender == owner, "Unauthorized call");
        secretMessageDataIDs[teamNumber] = dataID;
        isSecretSet[teamNumber] = true;
        emit SecretSet(teamNumber, dataID);
    }

    // Attempt to guess the secret message for any team
    function scrt(string memory guess) external returns (bytes memory) {
        bool success = false;
        uint256 successfulTeam = 0;

        for (uint256 teamNumber = 1; teamNumber <= 5; teamNumber++) {
            if (isSecretSet[teamNumber]) {
                string memory key = string(abi.encodePacked(SECRET_KEY_BASE, uint2str(teamNumber)));
                bytes memory secretMessageBytes = Suave.confidentialRetrieve(secretMessageDataIDs[teamNumber], key);
                string memory secretMessage = string(secretMessageBytes);
                if (keccak256(bytes(guess)) == keccak256(bytes(secretMessage))) {
                    success = true;
                    successfulTeam = teamNumber;
                    break;
                }
            }
        }

        emit AllAttemptsResult(success, successfulTeam);
        return bytes.concat(this.onchain_attemptSecretMessage.selector, abi.encode(success, successfulTeam));
    }

    function onchain_attemptSecretMessage(bool success, uint256 teamNumber) public {
        emit AttemptResult(teamNumber, success);
    }

    function uint2str(uint256 _i) internal pure returns (string memory) {
        if (_i == 0) {
            return "0";
        }
        uint256 j = _i;
        uint256 len;
        while (j != 0) {
            len++;
            j /= 10;
        }
        bytes memory bstr = new bytes(len);
        uint256 k = len;
        while (_i != 0) {
            k = k - 1;
            uint8 temp = (48 + uint8(_i - _i / 10 * 10));
            bytes1 b1 = bytes1(temp);
            bstr[k] = b1;
            _i /= 10;
        }
        return string(bstr);
    }
}
