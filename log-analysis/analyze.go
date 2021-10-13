package main

import (
	"fmt"
	"strconv"
	"strings"
	"time"
)

const (
	proposalMessage          = "PROPOSING NEW BLOCK ------------------------------------------------"
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

//Go constant string
func getMeasuredMetrics() []string {
	return []string{"proposing", "crosslinkProposal", "receivedCommitSig", "commitSigReady", "newBlockProposal",
		"startingConsensus", "sentAnnounce", "firstPrepare", "enoughPrepared", "sentPrepare", "firstCommit",
		"enoughCommitted", "allCommitted", "gracePeriodEnd", "consensusReached"}
}

func analyzeOutput(logMap map[string]interface{}) {
	messageTime, _ := time.Parse(time.RFC3339, logMap["time"].(string))

	switch logMap["message"] {
	case proposalMessage:
		newBlock := map[string]time.Time{"proposing": messageTime}
		blocks = append(blocks, newBlock)
		furthestBlock++

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

	//Optimized for readability/future-proofing, can be faster by checking if signers-count==1 first
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

	//Optimized for readability/future-proofing, can be faster by checking if signers-count==1 first
	case addVoteMessage:
		if logMap["phase"] == "Prepare" {
			if _, exists := blocks[currentBlock]["firstPrepare"]; !exists {
				blocks[currentBlock]["firstPrepare"] = messageTime
			}
		} else if logMap["phase"] == "Commit" {
			if _, exists := blocks[currentBlock]["firstCommit"]; !exists {
				blocks[currentBlock]["firstCommit"] = messageTime
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
		blocks[blockNum]["allCommitted"] = messageTime

	case gracePeriodEndMessage:
		blockNum := int(logMap["MsgBlockNum"].(float64)) - 1
		blocks[blockNum]["gracePeriodEnd"] = messageTime
		if _, exists := blocks[blockNum]["consensusReached"]; exists {
			submitBlockData(blocks[blockNum], currentBlock)
			currentBlock++
		}

	case consensusReachedMessage:
		blockNum := int(logMap["blockNum"].(float64)) - 1
		blocks[blockNum]["consensusReached"] = messageTime
		if _, exists := blocks[currentBlock]["gracePeriodEnd"]; exists {
			submitBlockData(blocks[currentBlock], currentBlock)
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

func submitBlockData(blockTime map[string]time.Time, block int) {
	metrics := getMeasuredMetrics()
	for metricIndex, metric := range metrics {
		if metricTime, exists := blockTime[metric]; exists {
			if metricIndex < len(metrics)-1 {
				nextMetric := metrics[metricIndex+1]
				if nextMetricTime, exists := blockTime[nextMetric]; exists {
					fmt.Printf("Time between %s and %s for block %v was %v\n", metric, nextMetric, block+1, nextMetricTime.Sub(metricTime))
				}
			}
		} else {
			fmt.Printf("There is no metric %s for block %v\n", metric, block+1)
		}
	}
}
