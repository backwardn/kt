package main

import (
	"os"
	"reflect"
	"sort"
	"testing"
	"time"

	"github.com/Shopify/sarama"
)

func TestParseOffsets(t *testing.T) {

	data := []struct {
		input       string
		expected    map[int32]interval
		expectedErr error
	}{
		{
			input: "",
			expected: map[int32]interval{
				-1: interval{
					start: offset{typ: relOffset, start: sarama.OffsetOldest},
					end:   offset{typ: absOffset, start: 1<<63 - 1},
				},
			},
			expectedErr: nil,
		},
		{
			input: "all",
			expected: map[int32]interval{
				-1: interval{
					start: offset{typ: relOffset, start: sarama.OffsetOldest},
					end:   offset{typ: absOffset, start: 1<<63 - 1},
				},
			},
			expectedErr: nil,
		},
		{
			input: "	all ",
			expected: map[int32]interval{
				-1: interval{
					start: offset{typ: relOffset, start: sarama.OffsetOldest},
					end:   offset{typ: absOffset, start: 1<<63 - 1},
				},
			},
			expectedErr: nil,
		},
		{
			input: "all=+0:",
			expected: map[int32]interval{
				-1: interval{
					start: offset{typ: relOffset, start: sarama.OffsetOldest, diff: 0},
					end:   offset{typ: absOffset, start: 1<<63 - 1},
				},
			},
			expectedErr: nil,
		},
		{
			input: "1,2,4",
			expected: map[int32]interval{
				1: interval{
					start: offset{typ: relOffset, start: sarama.OffsetOldest},
					end:   offset{typ: absOffset, start: 1<<63 - 1},
				},
				2: interval{
					start: offset{typ: relOffset, start: sarama.OffsetOldest},
					end:   offset{typ: absOffset, start: 1<<63 - 1},
				},
				4: interval{
					start: offset{typ: relOffset, start: sarama.OffsetOldest},
					end:   offset{typ: absOffset, start: 1<<63 - 1},
				},
			},
			expectedErr: nil,
		},
		{
			input: "0=",
			expected: map[int32]interval{
				0: interval{
					start: offset{typ: relOffset, start: sarama.OffsetOldest},
					end:   offset{typ: absOffset, start: 1<<63 - 1},
				},
			},
			expectedErr: nil,
		},
		{
			input: "0=1",
			expected: map[int32]interval{
				0: interval{
					start: offset{typ: absOffset, start: 1},
					end:   offset{typ: absOffset, start: 1<<63 - 1},
				},
			},
			expectedErr: nil,
		},
		{
			input: "0=1:",
			expected: map[int32]interval{
				0: interval{
					start: offset{typ: absOffset, start: 1},
					end:   offset{typ: absOffset, start: 1<<63 - 1},
				},
			},
			expectedErr: nil,
		},
		{
			input: "0=4:,2=1:10,6",
			expected: map[int32]interval{
				0: interval{
					start: offset{typ: absOffset, start: 4},
					end:   offset{typ: absOffset, start: 1<<63 - 1},
				},
				2: interval{
					start: offset{typ: absOffset, start: 1},
					end:   offset{typ: absOffset, start: 10},
				},
				6: interval{
					start: offset{typ: relOffset, start: sarama.OffsetOldest},
					end:   offset{typ: absOffset, start: 1<<63 - 1},
				},
			},
			expectedErr: nil,
		},
		{
			input: "0=-1",
			expected: map[int32]interval{
				0: interval{
					start: offset{typ: relOffset, start: sarama.OffsetNewest, diff: -1},
					end:   offset{typ: absOffset, start: 1<<63 - 1},
				},
			},
			expectedErr: nil,
		},
		{
			input: "0=-1:",
			expected: map[int32]interval{
				0: interval{
					start: offset{typ: relOffset, start: sarama.OffsetNewest, diff: -1},
					end:   offset{typ: absOffset, start: 1<<63 - 1},
				},
			},
			expectedErr: nil,
		},
		{
			input: "0=+1",
			expected: map[int32]interval{
				0: interval{
					start: offset{typ: relOffset, start: sarama.OffsetOldest, diff: 1},
					end:   offset{typ: absOffset, start: 1<<63 - 1},
				},
			},
			expectedErr: nil,
		},
		{
			input: "0=+1:",
			expected: map[int32]interval{
				0: interval{
					start: offset{typ: relOffset, start: sarama.OffsetOldest, diff: 1},
					end:   offset{typ: absOffset, start: 1<<63 - 1},
				},
			},
			expectedErr: nil,
		},
		{
			input: "0=+1:-1",
			expected: map[int32]interval{
				0: interval{
					start: offset{typ: relOffset, start: sarama.OffsetOldest, diff: 1},
					end:   offset{typ: relOffset, start: sarama.OffsetNewest, diff: -1},
				},
			},
			expectedErr: nil,
		},
		{
			input: "0=+1:-1,all=1:10",
			expected: map[int32]interval{
				0: interval{
					start: offset{typ: relOffset, start: sarama.OffsetOldest, diff: 1},
					end:   offset{typ: relOffset, start: sarama.OffsetNewest, diff: -1},
				},
				-1: interval{
					start: offset{typ: absOffset, start: 1, diff: 0},
					end:   offset{typ: absOffset, start: 10, diff: 0},
				},
			},
			expectedErr: nil,
		},
		{
			input: "0=oldest:newest",
			expected: map[int32]interval{
				0: interval{
					start: offset{typ: relOffset, start: sarama.OffsetOldest, diff: 0},
					end:   offset{typ: relOffset, start: sarama.OffsetNewest, diff: 0},
				},
			},
			expectedErr: nil,
		},
		{
			input: "0=oldest+10:newest-10",
			expected: map[int32]interval{
				0: interval{
					start: offset{typ: relOffset, start: sarama.OffsetOldest, diff: 10},
					end:   offset{typ: relOffset, start: sarama.OffsetNewest, diff: -10},
				},
			},
			expectedErr: nil,
		},
		{
			input: "newest",
			expected: map[int32]interval{
				-1: interval{
					start: offset{typ: relOffset, start: sarama.OffsetNewest, diff: 0},
					end:   offset{typ: absOffset, start: 1<<63 - 1, diff: 0},
				},
			},
			expectedErr: nil,
		},
		{
			input: "10",
			expected: map[int32]interval{
				10: interval{
					start: offset{typ: relOffset, start: sarama.OffsetOldest, diff: 0},
					end:   offset{typ: absOffset, start: 1<<63 - 1, diff: 0},
				},
			},
			expectedErr: nil,
		},
		{
			input: "newest",
			expected: map[int32]interval{
				-1: interval{
					start: offset{typ: relOffset, start: sarama.OffsetNewest, diff: 0},
					end:   offset{typ: absOffset, start: 1<<63 - 1, diff: 0},
				},
			},
			expectedErr: nil,
		},
		{
			input: "all=newest:",
			expected: map[int32]interval{
				-1: interval{
					start: offset{typ: relOffset, start: sarama.OffsetNewest, diff: 0},
					end:   offset{typ: absOffset, start: 1<<63 - 1, diff: 0},
				},
			},
			expectedErr: nil,
		},
		{
			input: "newest-10:",
			expected: map[int32]interval{
				-1: interval{
					start: offset{typ: relOffset, start: sarama.OffsetNewest, diff: -10},
					end:   offset{typ: absOffset, start: 1<<63 - 1, diff: 0},
				},
			},
			expectedErr: nil,
		},
		{
			input: "oldest+10:",
			expected: map[int32]interval{
				-1: interval{
					start: offset{typ: relOffset, start: sarama.OffsetOldest, diff: 10},
					end:   offset{typ: absOffset, start: 1<<63 - 1, diff: 0},
				},
			},
			expectedErr: nil,
		},
		{
			input: "-10:",
			expected: map[int32]interval{
				-1: interval{
					start: offset{typ: relOffset, start: sarama.OffsetNewest, diff: -10},
					end:   offset{typ: absOffset, start: 1<<63 - 1, diff: 0},
				},
			},
			expectedErr: nil,
		},
		{
			input: "+10:",
			expected: map[int32]interval{
				-1: interval{
					start: offset{typ: relOffset, start: sarama.OffsetOldest, diff: 10},
					end:   offset{typ: absOffset, start: 1<<63 - 1, diff: 0},
				},
			},
			expectedErr: nil,
		},
	}

	for _, d := range data {
		actual, err := parseOffsets(d.input)
		if err != d.expectedErr || !reflect.DeepEqual(actual, d.expected) {
			t.Errorf(
				`
Expected: %+v, err=%v
Actual:   %+v, err=%v
Input:    %v
`,
				d.expected,
				d.expectedErr,
				actual,
				err,
				d.input,
			)
		}
	}

}

func TestFindPartitionsToConsume(t *testing.T) {
	data := []struct {
		config   consumeConfig
		consumer tConsumer
		expected []int32
	}{
		{
			config: consumeConfig{
				topic: "a",
				offsets: map[int32]interval{
					10: {offset{absOffset, 2, 0}, offset{absOffset, 4, 0}},
				},
			},
			consumer: tConsumer{
				topics:              []string{"a"},
				topicsErr:           nil,
				partitions:          map[string][]int32{"a": []int32{0, 10}},
				partitionsErr:       map[string]error{"a": nil},
				consumePartition:    map[tConsumePartition]tPartitionConsumer{},
				consumePartitionErr: map[tConsumePartition]error{},
				closeErr:            nil,
			},
			expected: []int32{10},
		},
		{
			config: consumeConfig{
				topic: "a",
				offsets: map[int32]interval{
					-1: {offset{absOffset, 3, 0}, offset{absOffset, 41, 0}},
				},
			},
			consumer: tConsumer{
				topics:              []string{"a"},
				topicsErr:           nil,
				partitions:          map[string][]int32{"a": []int32{0, 10}},
				partitionsErr:       map[string]error{"a": nil},
				consumePartition:    map[tConsumePartition]tPartitionConsumer{},
				consumePartitionErr: map[tConsumePartition]error{},
				closeErr:            nil,
			},
			expected: []int32{0, 10},
		},
	}

	for _, d := range data {
		actual := findPartitions(d.consumer, d.config)

		if !reflect.DeepEqual(actual, d.expected) {
			t.Errorf(
				`
Expected: %+v
Actual:   %+v
Input:    config=%+v
	`,
				d.expected,
				actual,
				d.config,
			)
			return
		}
	}
}

func TestConsume(t *testing.T) {
	closer := make(chan struct{})
	config := consumeConfig{
		topic:   "hans",
		brokers: []string{"localhost:9092"},
		offsets: map[int32]interval{
			-1: interval{start: offset{absOffset, 1, 0}, end: offset{absOffset, 5, 0}},
		},
	}
	messageChan := make(<-chan *sarama.ConsumerMessage)
	calls := make(chan tConsumePartition)
	consumer := tConsumer{
		consumePartition: map[tConsumePartition]tPartitionConsumer{
			tConsumePartition{"hans", 1, 1}: tPartitionConsumer{messages: messageChan},
			tConsumePartition{"hans", 2, 1}: tPartitionConsumer{messages: messageChan},
		},
		calls: calls,
	}
	partitions := []int32{1, 2}

	go consume(config, closer, consumer, partitions)
	defer close(closer)

	end := make(chan struct{})
	go func(c chan tConsumePartition, e chan struct{}) {
		actual := []tConsumePartition{}
		expected := []tConsumePartition{
			tConsumePartition{"hans", 1, 1},
			tConsumePartition{"hans", 2, 1},
		}
		for {
			select {
			case call := <-c:
				actual = append(actual, call)
				sort.Sort(ByPartitionOffset(actual))
				if reflect.DeepEqual(actual, expected) {
					e <- struct{}{}
					return
				}
				if len(actual) == len(expected) {
					t.Errorf(
						`Got expected number of calls, but they are different.
Expected: %#v
Actual:   %#v
`,
						expected,
						actual,
					)
				}
			case _, ok := <-e:
				if !ok {
					return
				}
			}
		}
	}(calls, end)

	select {
	case <-end:
	case <-time.After(1 * time.Second):
		t.Errorf("Did not receive calls to consume partitions before timeout.")
		close(end)
	}
}

type tConsumePartition struct {
	topic     string
	partition int32
	offset    int64
}

type ByPartitionOffset []tConsumePartition

func (a ByPartitionOffset) Len() int {
	return len(a)
}
func (a ByPartitionOffset) Swap(i, j int) {
	a[i], a[j] = a[j], a[i]
}
func (a ByPartitionOffset) Less(i, j int) bool {
	return a[i].partition < a[j].partition || a[i].offset < a[j].offset
}

type tConsumerError struct {
	Topic     string
	Partition int32
	Err       error
}

type tPartitionConsumer struct {
	closeErr            error
	highWaterMarkOffset int64
	messages            <-chan *sarama.ConsumerMessage
	errors              <-chan *sarama.ConsumerError
}

func (pc tPartitionConsumer) AsyncClose() {}
func (pc tPartitionConsumer) Close() error {
	return pc.closeErr
}
func (pc tPartitionConsumer) HighWaterMarkOffset() int64 {
	return pc.highWaterMarkOffset
}
func (pc tPartitionConsumer) Messages() <-chan *sarama.ConsumerMessage {
	return pc.messages
}
func (pc tPartitionConsumer) Errors() <-chan *sarama.ConsumerError {
	return pc.errors
}

type tConsumer struct {
	topics              []string
	topicsErr           error
	partitions          map[string][]int32
	partitionsErr       map[string]error
	consumePartition    map[tConsumePartition]tPartitionConsumer
	consumePartitionErr map[tConsumePartition]error
	closeErr            error
	calls               chan tConsumePartition
}

func (c tConsumer) Topics() ([]string, error) {
	return c.topics, c.topicsErr
}

func (c tConsumer) Partitions(topic string) ([]int32, error) {
	return c.partitions[topic], c.partitionsErr[topic]
}

func (c tConsumer) ConsumePartition(topic string, partition int32, offset int64) (sarama.PartitionConsumer, error) {
	cp := tConsumePartition{topic, partition, offset}
	c.calls <- cp
	return c.consumePartition[cp], c.consumePartitionErr[cp]
}

func (c tConsumer) Close() error {
	return c.closeErr
}

func TestConsumeParseArgs(t *testing.T) {
	configBefore := config
	defer func() {
		config = configBefore
	}()

	expectedTopic := "test-topic"
	givenBroker := "hans:9092"
	expectedBrokers := []string{givenBroker}

	config.consume.args.topic = ""
	config.consume.args.brokers = ""
	os.Setenv("KT_TOPIC", expectedTopic)
	os.Setenv("KT_BROKERS", givenBroker)

	consumeParseArgs()
	if config.consume.topic != expectedTopic ||
		!reflect.DeepEqual(config.consume.brokers, expectedBrokers) {
		t.Errorf(
			"Expected topic %v and brokers %v from env vars, got topic %v and brokers %v.",
			expectedTopic,
			expectedBrokers,
			config.consume.topic,
			config.consume.brokers,
		)
		return
	}

	// default brokers to localhost:9092
	os.Setenv("KT_TOPIC", "")
	os.Setenv("KT_BROKERS", "")
	config.consume.args.topic = expectedTopic
	config.consume.args.brokers = ""
	expectedBrokers = []string{"localhost:9092"}

	consumeParseArgs()
	if config.consume.topic != expectedTopic ||
		!reflect.DeepEqual(config.consume.brokers, expectedBrokers) {
		t.Errorf(
			"Expected topic %v and brokers %v from env vars, got topic %v and brokers %v.",
			expectedTopic,
			expectedBrokers,
			config.consume.topic,
			config.consume.brokers,
		)
		return
	}

	// command line arg wins
	os.Setenv("KT_TOPIC", "BLUBB")
	os.Setenv("KT_BROKERS", "BLABB")
	config.consume.args.topic = expectedTopic
	config.consume.args.brokers = givenBroker
	expectedBrokers = []string{givenBroker}

	consumeParseArgs()
	if config.consume.topic != expectedTopic ||
		!reflect.DeepEqual(config.consume.brokers, expectedBrokers) {
		t.Errorf(
			"Expected topic %v and brokers %v from env vars, got topic %v and brokers %v.",
			expectedTopic,
			expectedBrokers,
			config.consume.topic,
			config.consume.brokers,
		)
		return
	}
}
