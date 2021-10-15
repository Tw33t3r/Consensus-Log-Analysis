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
var furthestBlock = -1

//Go constant string array
func getMeasuredMetrics() []string {
	return []string{"proposing", "vrfGenerated", "crosslinkProposal", "receivedCommitSig", "commitSigReady", "newBlockProposal",
		"startingConsensus", "sentAnnounce", "firstPrepare", "enoughPrepared", "sentPrepare", "firstCommit",
		"enoughCommitted", "95PercentCommitted", "100PercentCommitted", "gracePeriodEnd", "consensusReached"}
}

func analyzeOutput(logMap map[string]interface{}) {
	messageTime, _ := time.Parse(time.RFC3339, logMap["time"].(string))

	switch logMap["message"] {
	case proposalMessage:
		newBlock := map[string]time.Time{"proposing": messageTime}
		blocks = append(blocks, newBlock)
		furthestBlock++

	case vrfGeneratedMessage:
		blockNum := int(logMap["BlockNum"].(float64)) - 1
		blocks[blockNum]["vrfGenerated"] = messageTime

	case recievedCommitSigMessage:
		if _, exists := blocks[currentBlock]["recievedCommitSig"]; !exists {
			blocks[currentBlock]["receivedCommitSig"] = messageTime
		}

	case commitSigReadyMessage:
		if _, exists := blocks[currentBlock]["commitSigReady"]; !exists {
			blocks[currentBlock]["commitSigReady"] = messageTime
		}

	case newBlockProposalMessage:
		blockNum := int(logMap["blockNum"].(float64)) - 1
		blocks[blockNum]["newBlockProposal"] = messageTime

	case startingConsensusMessage:
		blockNum := int(logMap["myBlock"].(float64)) - 1
		blocks[blockNum]["startingConsensus"] = messageTime

	case sentAnnounceMessage:
		blockNum := int(logMap["myBlock"].(float64)) - 1
		blocks[blockNum]["sentAnnounce"] = messageTime

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
		blockNum := int(logMap["blockNum"].(float64)) - 1
		blocks[blockNum]["sentPrepare"] = messageTime

	case twoThirdsMessage:
		blockNum := int(logMap["MsgBlockNum"].(float64)) - 1
		blocks[blockNum]["enoughCommitted"] = messageTime

	case gracePeriodStartMessage:
		blockNum := int(logMap["myBlock"].(float64)) - 1
		blocks[blockNum]["gracePeriodStart"] = messageTime

	case insertedNewBlockMessage:
		blockNum, _ := strconv.Atoi(logMap["number"].(string))
		blocks[blockNum-1]["insertedNewBlock"] = messageTime

	case sentCommittedMessage:
		blockNum := int(logMap["blockNum"].(float64)) - 1
		blocks[blockNum]["sentCommitted"] = messageTime

	case oneHundredPercentMessage:
		blockNum := int(logMap["MsgBlockNum"].(float64)) - 1
		blocks[blockNum]["100PercentCommitted"] = messageTime

	case gracePeriodEndMessage:
		blockNum := int(logMap["MsgBlockNum"].(float64)) - 1
		blocks[blockNum]["gracePeriodEnd"] = messageTime
		if _, exists := blocks[blockNum]["consensusReached"]; exists {
			submitBlockData(currentBlock)
			currentBlock++
		}

	case consensusReachedMessage:
		blockNum := int(logMap["blockNum"].(float64)) - 1
		blocks[blockNum]["consensusReached"] = messageTime
		if _, exists := blocks[currentBlock]["gracePeriodEnd"]; exists {
			submitBlockData(currentBlock)
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
	blockTime := blocks[block]
	metrics := getMeasuredMetrics()
	firstMetric := metrics[0]
	firstMetricTime := blockTime[firstMetric]
	compareNextMetric(firstMetric, firstMetricTime, blockTime, metrics[1:], block)
}

func compareNextMetric(metric string, metricTime time.Time, metricTimes map[string]time.Time, metricArray []string, block int) {
	if len(metricArray) > 0 {
		nextMetric := metricArray[0]
		if nextMetricTime, exists := metricTimes[nextMetric]; exists {
			fmt.Printf("Time between %s and %s for block %v was %v\n", metric, nextMetric, block+1, nextMetricTime.Sub(metricTime))
			compareNextMetric(nextMetric, nextMetricTime, metricTimes, metricArray[1:], block)
		} else {
			fmt.Printf("There is no metric %s for block %v\n", nextMetric, block+1)
			compareNextMetric(metric, metricTime, metricTimes, metricArray[1:], block)
		}
	}
}
