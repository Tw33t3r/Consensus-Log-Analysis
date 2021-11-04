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

type averageAmount struct {
	amount      int64
	averageTime int64
}

var blocks []map[string]time.Time
var currentBlock = 0
var blockOffset = 0
var furthestBlock = -1
var initializedFirstBlock = false
var averages = make(map[string]*averageAmount)

//Go constant string array
//Vrf Generation, 95 Percent Committed, and 100 Percent committed don't appear at consistent times
//relative to other metrics, so these metrics are ignored for average values.
func getMeasuredMetrics() []string {
	return []string{"proposing", "receivedCommitSig", "crosslinkProposal", "commitSigReady", "newBlockProposal",
		"startingConsensus", "sentAnnounce", "firstPrepare", "enoughPrepared", "sentPrepare", "firstCommit",
		"enoughCommitted", "gracePeriodStart", "insertedNewBlock", "sentCommitted", "consensusReached", "gracePeriodEnd"}
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
		arrayBlockNumber := blockNum - blockOffset
		if arrayBlockNumber < len(blocks) && arrayBlockNumber >= 0 {
			blocks[arrayBlockNumber]["vrfGenerated"] = messageTime
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
		arrayBlockNumber := blockNum - blockOffset
		if arrayBlockNumber < len(blocks) && arrayBlockNumber >= 0 {
			blocks[arrayBlockNumber]["newBlockProposal"] = messageTime
		}

	case startingConsensusMessage:
		blockNum := int(logMap["myBlock"].(float64))
		arrayBlockNumber := blockNum - blockOffset
		if arrayBlockNumber < len(blocks) && arrayBlockNumber >= 0 {
			blocks[arrayBlockNumber]["startingConsensus"] = messageTime
		}

	case sentAnnounceMessage:
		blockNum := int(logMap["myBlock"].(float64))
		arrayBlockNumber := blockNum - blockOffset
		if arrayBlockNumber < len(blocks) && arrayBlockNumber >= 0 {
			blocks[arrayBlockNumber]["sentAnnounce"] = messageTime
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
		arrayBlockNumber := blockNum - blockOffset
		if arrayBlockNumber < len(blocks) && arrayBlockNumber >= 0 {
			blocks[arrayBlockNumber]["sentPrepare"] = messageTime
		}

	case twoThirdsMessage:
		blockNum := int(logMap["MsgBlockNum"].(float64))
		arrayBlockNumber := blockNum - blockOffset
		if arrayBlockNumber < len(blocks) && arrayBlockNumber >= 0 {
			blocks[arrayBlockNumber]["enoughCommitted"] = messageTime
		}

	case gracePeriodStartMessage:
		blockNum := int(logMap["myBlock"].(float64))
		arrayBlockNumber := blockNum - blockOffset
		if arrayBlockNumber < len(blocks) && arrayBlockNumber >= 0 {
			blocks[arrayBlockNumber]["gracePeriodStart"] = messageTime
		}

	case insertedNewBlockMessage:
		blockNum, _ := strconv.Atoi(logMap["number"].(string))
		arrayBlockNumber := blockNum - blockOffset
		if arrayBlockNumber < len(blocks) && arrayBlockNumber >= 0 {
			blocks[arrayBlockNumber]["insertedNewBlock"] = messageTime
		}

	case sentCommittedMessage:
		blockNum := int(logMap["blockNum"].(float64))
		arrayBlockNumber := blockNum - blockOffset
		if arrayBlockNumber < len(blocks) && arrayBlockNumber >= 0 {
			blocks[arrayBlockNumber]["sentCommitted"] = messageTime
		}

	case oneHundredPercentMessage:
		blockNum := int(logMap["MsgBlockNum"].(float64))
		arrayBlockNumber := blockNum - blockOffset
		if arrayBlockNumber < len(blocks) && arrayBlockNumber >= 0 {
			blocks[arrayBlockNumber]["100PercentCommitted"] = messageTime
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
		if blockNum-blockOffset >= 0 {
			blocks[blockNum-blockOffset]["consensusReached"] = messageTime
			if _, exists := blocks[currentBlock]["gracePeriodEnd"]; exists {
				submitBlockData(blockOffset + currentBlock)
				currentBlock++
			}
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
			timeDifference := nextMetricTime.Sub(metricTime)
			fmt.Printf("Time between %s and %s for block %v was %v\n", metric, nextMetric, block, timeDifference)
			compareNextMetric(nextMetric, nextMetricTime, metricTimes, metricArray[1:], block)
			addAverage(metric+nextMetric, int64(timeDifference))
		} else {
			compareNextMetric(metric, metricTime, metricTimes, metricArray[1:], block)
		}
	}
}

func addAverage(metric string, metricTime int64) {
	if oldAverage, exists := averages[metric]; exists {
		newAmount := oldAverage.amount + 1
		newAverage := oldAverage.averageTime + ((metricTime - oldAverage.averageTime) / newAmount)
		averages[metric] = &averageAmount{amount: newAmount, averageTime: newAverage}
	} else {
		averages[metric] = &averageAmount{amount: 1, averageTime: metricTime}
	}
}

func printAverages() {
	metrics := getMeasuredMetrics()
	firstMetric := metrics[0]
	recurseAverages(firstMetric, metrics[1:])
}

func recurseAverages(metric string, metricArray []string) {
	if len(metricArray) > 0 {
		nextMetric := metricArray[0]
		key := metric + nextMetric
		if averageMetric, exists := averages[key]; exists {
			fmt.Printf("Average Time between %s and %s was %v\n", metric, nextMetric, time.Duration(averageMetric.averageTime))
			recurseAverages(nextMetric, metricArray[1:])
		} else {
			recurseAverages(metric, metricArray[1:])
		}
	}
}
