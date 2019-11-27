package main

import "testing"

func TestGetMessageTimestamp(t *testing.T) {
	m, ts := getMessageTimestamp("1574833080.290678 devops-pc sudo[6817]: pam_unix(sudo:session): session closed for user root")
	expectedMessage := "devops-pc sudo[6817]: pam_unix(sudo:session): session closed for user root"
	var expectedTimestamp int64
	expectedTimestamp = 1574833080
	if m != expectedMessage {
		t.Fatalf("Message is Wrong!\nExpected: %s\nReceived: %s\n", expectedMessage, m)
	}
	if ts != expectedTimestamp {
		t.Fatalf("Timestamp is Wrong!\nExpected: %d\nReceived: %d\n", expectedTimestamp, ts)
	}
}
