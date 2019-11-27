package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"strconv"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

func main() {

	dryRun := flag.Bool("dry-run", false, "Set if you want to output messages to console. Useful for testing.")
	logGroup := flag.String("log-group", "", "Specify the log group where you want to send the logs")
	logStream := flag.String("log-stream", "", "Specify the log stream where you want to send the logs")
	flag.Parse()

	if !*dryRun && (*logGroup == "" || *logStream == "") {
		log.Fatalf("You must specify both the log group and the log stream.\nCurrent logGroup: %s\nCurrent logStream: %s\nSee %s -h for help.", *logGroup, *logStream, os.Args[0])
	}

	cmd := exec.Command("journalctl", "-f", "--no-pager", "-o", "short-unix")

	pipe, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("Error in StdoutPipe\n%v", err)
	}

	err = cmd.Start()
	if err != nil {
		log.Fatalf("Error in Start\n%v", err)
	}

	mySession := session.Must(session.NewSession())
	cloudwatchlogsClient := cloudwatchlogs.New(mySession)

	var counter = 0
	var events = make([]*cloudwatchlogs.InputLogEvent, 10)

	scanner := bufio.NewScanner(pipe)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		m := scanner.Text()

		if m[0] == '-' {
			continue
		}

		message, timestamp := getMessageTimestamp(m)

		events[counter] = &cloudwatchlogs.InputLogEvent{
			Message:   &message,
			Timestamp: &timestamp,
		}
		if counter == 9 {
			counter = 0
			if !*dryRun {
				sendEventsCloudwatch(events, logGroup, logStream, cloudwatchlogsClient)
			}
		}
		sendEventsConsole(events)
		counter++
	}
	cmd.Wait()
}

func sendEventsCloudwatch(events []*cloudwatchlogs.InputLogEvent, logGroupName *string, logStreamName *string, cloudwatchlogsClient *cloudwatchlogs.CloudWatchLogs) {
	putLogEventInput := &cloudwatchlogs.PutLogEventsInput{
		LogEvents:     events,
		LogGroupName:  logGroupName,
		LogStreamName: logStreamName,
	}
	_, err := cloudwatchlogsClient.PutLogEvents(putLogEventInput)
	if err != nil {
		log.Fatalf("Error in sendEventsCloudwatch\n%v", err)
	}
}

func sendEventsConsole(events []*cloudwatchlogs.InputLogEvent) {
	fmt.Printf("%v\n", events)
}

func getMessageTimestamp(m string) (string, int64) {
	splittedMessage := strings.Split(m, " ")
	timestamp := strings.Split(splittedMessage[0], ".")

	timestampNum, err := strconv.ParseInt(timestamp[0], 10, 64)
	if err != nil {
		log.Fatalf("Error in getMessageTimestamp\n%v\nMessage: %s\ntimestamp: %s\n", m, timestamp[0], err)
	}

	t := time.Unix(timestampNum, 0)

	message := strings.Join(splittedMessage[1:], " ")
	return message, t.Unix()
}
