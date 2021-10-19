package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	proposalMessage          = "PROPOSING NEW BLOCK ------------------------------------------------"
	vrfGeneratedMessage      = "[GenerateVrfAndProof] Leader generated a VRF"
	recievedCommitSigMessage = "[ProposeNewBlock] received commit sigs asynchronously"
	//Partial match, uses contains instead of ==
	crosslinkProposalMessage = " pending crosslinks"
	commitSigReadyMessage    = "Commit sigs are ready"
	newBlockProposalMessage  = "=========Successfully Proposed New Block=========="
	startingConsensusMessage = "[ConsensusMainLoop] STARTING CONSENSUS"
	sentAnnounceMessage      = "[Announce] Sent Announce Message!!"
	quorumMessage            = "Quorum details"
	addVoteMessage           = "[AddNewVote] New Vote Added!"
	prepareQuorumMessage     = "[OnPrepare] Received Enough Prepare Signatures"
	sentPrepareMessage       = "[OnPrepare] Sent Prepared Message!!"
	twoThirdsMessage         = "[OnCommit] 2/3 Enough commits received"
	gracePeriodStartMessage  = "[OnCommit] Starting Grace Period"
	insertedNewBlockMessage  = "Inserted new block"
	sentCommittedMessage     = "[preCommitAndPropose] Sent Committed Message"
	oneHundredPercentMessage = "[OnCommit] 100% Enough commits received"
	gracePeriodEndMessage    = "[OnCommit] Commit Grace Period Ended"
	consensusReachedMessage  = "HOORAY!!!!!!! CONSENSUS REACHED!!!!!!!"
)

var blocks []map[string]time.Time
var currentBlock = 0
var blockOffset = 0
var furthestBlock = -1
var initializedFirstBlock = false

//Go constant string array
func getMeasuredMetrics() []string {
	return []string{"proposing", "receivedCommitSig", "vrfGenerated", "crosslinkProposal", "commitSigReady", "newBlockProposal",
		"startingConsensus", "sentAnnounce", "firstPrepare", "enoughPrepared", "sentPrepare", "firstCommit",
		"enoughCommitted", "95PercentCommitted", "100PercentCommitted", "consensusReached", "gracePeriodEnd"}
}

func analyzeOutput(logMap map[string]interface{}) {
	messageTime, _ := time.Parse(time.RFC3339, logMap["time"].(string))

	switch logMap["message"] {
	case proposalMessage:
		if !initializedFirstBlock {
			blockOffset = int(logMap["blockNum"].(float64))
			initializedFirstBlock = true
		}
		newBlock := map[string]time.Time{"proposing": messageTime}
		blocks = append(blocks, newBlock)
		furthestBlock++

	case vrfGeneratedMessage:
		blockNum := int(logMap["BlockNum"].(float64))
		if blockNum-blockOffset < len(blocks) {
			blocks[blockNum-blockOffset]["vrfGenerated"] = messageTime
		}

	case recievedCommitSigMessage:
		if _, exists := blocks[furthestBlock]["commitSigReady"]; !exists {
			blocks[furthestBlock]["receivedCommitSig"] = messageTime
		}

	case commitSigReadyMessage:
		if _, exists := blocks[furthestBlock]["commitSigReady"]; !exists {
			blocks[furthestBlock]["commitSigReady"] = messageTime
		}

	case newBlockProposalMessage:
		blockNum := int(logMap["blockNum"].(float64))
		if blockNum-blockOffset < len(blocks) {

			blocks[blockNum-blockOffset]["newBlockProposal"] = messageTime
		}

	case startingConsensusMessage:
		blockNum := int(logMap["myBlock"].(float64))
		if blockNum-blockOffset < len(blocks) {
			blocks[blockNum-blockOffset]["startingConsensus"] = messageTime
		}

	case sentAnnounceMessage:
		blockNum := int(logMap["myBlock"].(float64))
		if blockNum-blockOffset < len(blocks) {
			blocks[blockNum-blockOffset]["sentAnnounce"] = messageTime
		}

	case quorumMessage:
		if logMap["phase"] == "Prepare" {
			if _, exists := blocks[currentBlock]["firstPrepare"]; !exists {
				blocks[currentBlock]["firstPrepare"] = messageTime
			}
		} else if logMap["phase"] == "Commit" {
			if _, exists := blocks[currentBlock]["firstCommit"]; !exists {
				blocks[currentBlock]["firstCommit"] = messageTime
			}
		}

	case addVoteMessage:
		if logMap["phase"] == "Prepare" {
			//Check for first instance of prepare after announce is sent out
			if _, exists := blocks[currentBlock]["sentAnnounce"]; exists {
				if _, exists := blocks[currentBlock]["firstPrepare"]; !exists {
					blocks[currentBlock]["firstPrepare"] = messageTime
				}
			}
		} else if logMap["phase"] == "Commit" {
			//Check for first instance of commit after announce is sent out
			if _, exists := blocks[currentBlock]["sentPrepare"]; exists {
				if _, exists := blocks[currentBlock]["firstCommit"]; !exists {
					blocks[currentBlock]["firstCommit"] = messageTime
				}
				if value, _ := strconv.ParseFloat(logMap["total-power-of-signers"].(string), 32); value > .95 {
					if _, exists := blocks[currentBlock]["95PercentCommitted"]; !exists {
						blocks[currentBlock]["95PercentCommitted"] = messageTime
					}
				}
			}
		}

	case prepareQuorumMessage:
		if _, exists := blocks[currentBlock]["enoughPrepared"]; !exists {
			blocks[currentBlock]["enoughPrepared"] = messageTime
		}

	case sentPrepareMessage:
		blockNum := int(logMap["blockNum"].(float64))
		if blockNum-blockOffset < len(blocks) {
			blocks[blockNum-blockOffset]["sentPrepare"] = messageTime
		}

	case twoThirdsMessage:
		blockNum := int(logMap["MsgBlockNum"].(float64))
		if blockNum-blockOffset < len(blocks) {
			blocks[blockNum-blockOffset]["enoughCommitted"] = messageTime
		}

	case gracePeriodStartMessage:
		blockNum := int(logMap["myBlock"].(float64))
		if blockNum-blockOffset < len(blocks) {
			blocks[blockNum-blockOffset]["gracePeriodStart"] = messageTime
		}

	case insertedNewBlockMessage:
		blockNum, _ := strconv.Atoi(logMap["number"].(string))
		if blockNum-blockOffset < len(blocks) {
			blocks[blockNum-blockOffset]["insertedNewBlock"] = messageTime
		}

	case sentCommittedMessage:
		blockNum := int(logMap["blockNum"].(float64))
		if blockNum-blockOffset < len(blocks) {
			blocks[blockNum-blockOffset]["sentCommitted"] = messageTime
		}

	case oneHundredPercentMessage:
		blockNum := int(logMap["MsgBlockNum"].(float64))
		if blockNum-blockOffset < len(blocks) {
			blocks[blockNum-blockOffset]["100PercentCommitted"] = messageTime
		}

	case gracePeriodEndMessage:
		blockNum := int(logMap["MsgBlockNum"].(float64))
		if blockNum-blockOffset >= 0 {
			blocks[blockNum-blockOffset]["gracePeriodEnd"] = messageTime
			if _, exists := blocks[blockNum-blockOffset]["consensusReached"]; exists {
				submitBlockData(blockOffset + currentBlock)
				currentBlock++
			}
		}

	case consensusReachedMessage:
		blockNum := int(logMap["blockNum"].(float64))
		blocks[blockNum-blockOffset]["consensusReached"] = messageTime
		if _, exists := blocks[currentBlock]["gracePeriodEnd"]; exists {
			submitBlockData(blockOffset + currentBlock)
			currentBlock++
		}
	default:
		if strings.Contains(logMap["message"].(string), " pending crosslinks") {
			if _, exists := blocks[currentBlock]["crosslinkProposal"]; !exists {
				blocks[currentBlock]["crosslinkProposal"] = messageTime
			} else {
				blocks[furthestBlock]["crosslinkProposal"] = messageTime
			}
		}
	}
}

func submitBlockData(block int) {
	blockTime := blocks[currentBlock]
	metrics := getMeasuredMetrics()
	firstMetric := metrics[0]
	firstMetricTime := blockTime[firstMetric]
	compareNextMetric(firstMetric, firstMetricTime, blockTime, metrics[1:], block)
}

func compareNextMetric(metric string, metricTime time.Time, metricTimes map[string]time.Time, metricArray []string, block int) {
	if len(metricArray) > 0 {
		nextMetric := metricArray[0]
		if nextMetricTime, exists := metricTimes[nextMetric]; exists {
			fmt.Printf("Time between %s and %s for block %v was %v\n", metric, nextMetric, block, nextMetricTime.Sub(metricTime))
			compareNextMetric(nextMetric, nextMetricTime, metricTimes, metricArray[1:], block)
		} else {
			compareNextMetric(metric, metricTime, metricTimes, metricArray[1:], block)
		}
	}
}
