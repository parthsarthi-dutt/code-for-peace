package queue

// Stream names and consumer group constants.
// Extracted into a separate package to avoid import cycles
// between database/ and queue/producer/.

const (
	// Submission streams in priority order (contest > practice > misc)
	StreamContest  = "contestSubmission"
	StreamPractice = "practiceSubmission"
	StreamMisc     = "miscellaneousSubmission"

	// Dead-letter queue for messages that fail processing repeatedly
	StreamDLQ = "deadLetterQueue"

	// ConsumerGroup is the name all workers join.
	// Redis guarantees each message goes to exactly ONE consumer in the group.
	ConsumerGroup = "judges"
)

// AllStreams lists every submission stream in priority order.
var AllStreams = []string{StreamContest, StreamPractice, StreamMisc}
