package output_test

import (
	"testing"
	"time"

	"github.com/ejv2/disdup/output"
)

var expectedChannelOutputs = []string{
	"@user1 (guild1) #chan1: Message 1",
	"@user1 (guild1) #chan2: Message 2",
	"@user1 (guild1) #chan2: Message 3",
	"@user2 (guild1) #chan2: Message 4",
	"@user1 (guild2) #chan1: Message 5",
	"@user1 (guild1) #chan1: Message 6",
	"@user1 (guild1) #chan1: Message 7",
	"@user1 (guild2) #chan1: Message 8",
}

func testChannel(t *testing.T) {
	out := output.Channel{
		Output: make(chan string),
	}
	out.Open(fakeSession)

	for _, test := range testMessages {
		go out.Write(test)
	}

	// NOTE: Channels act like a stack, so recieve in FILO order
	for i := len(testMessages) - 1; i <= 0; i-- {
		if i >= len(expectedChannelOutputs) {
			panic("Not enough test cases")
		}

		if got := <-out.Output; got != expectedChannelOutputs[i] {
			t.Errorf("Wrong response from Channel\nExpect:\n%s\n\nGot:\n%s", expectedChannelOutputs[i], got)
		}
	}
}

func testChannelTimeout(t *testing.T) {
	out := output.Channel{
		Output:  make(chan string),
		Timeout: time.Millisecond * 500,
	}
	out.Open(fakeSession)

	go out.Write(testMessages[0])
	time.Sleep(time.Second * 1)

	select {
	case <-out.Output:
		t.Error("Got response over channel which should have timed out")
	default:
	}
}

func TestChannel(t *testing.T) {
	t.Run("Normal", testChannel)
	t.Run("Timeout", testChannelTimeout)
}
