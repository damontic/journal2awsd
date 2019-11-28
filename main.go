package main

import (
	"bufio"
	"flag"
	"fmt"
	"log"
	"os"
	"os/exec"
	"sort"
	"strconv"
	"strings"

	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/cloudwatchlogs"
)

func main() {
	version := "0.0.0"

	versionFlag := flag.Bool("version", false, "Set if you want to see the version and exit.")
	dryRun := flag.Bool("dry-run", false, "Set if you want to output messages to console. Useful for testing.")
	logGroup := flag.String("group", "", "Specify the log group where you want to send the logs")
	logStream := flag.String("stream", "", "Specify the log stream where you want to send the logs")
	eventSize := flag.Int("size", 10, "Specify the number of events to send to AWS Cloudwatch.")
	flag.Parse()

	if *versionFlag {
		fmt.Println(version)
		os.Exit(0)
	}

	if !*dryRun && (*logGroup == "" || *logStream == "") {
		log.Fatalf("You must specify both the log group and the log stream.\nCurrent logGroup: %s\nCurrent logStream: %s\nSee %s -h for help.", *logGroup, *logStream, os.Args[0])
	}

	journal2awsd(dryRun, eventSize, logGroup, logStream)
}

func journal2awsd(dryRun *bool, eventSize *int, logGroup, logStream *string) {
	mySession := session.Must(session.NewSession())
	cloudwatchlogsClient := cloudwatchlogs.New(mySession)

	cmd := exec.Command("journalctl", "-f", "--no-pager", "-o", "short-unix")

	pipe, err := cmd.StdoutPipe()
	if err != nil {
		log.Fatalf("Error in StdoutPipe\n%v", err)
	}

	err = cmd.Start()
	if err != nil {
		log.Fatalf("Error in Start\n%v", err)
	}

	var counter = 0
	var events = make([]*cloudwatchlogs.InputLogEvent, *eventSize)
	var nextToken *string

	scanner := bufio.NewScanner(pipe)
	scanner.Split(bufio.ScanLines)
	for scanner.Scan() {
		m := scanner.Text()

		if m[0] == '-' || m[0] == ' ' {
			continue
		}

		message, timestamp := getMessageTimestamp(m)

		events[counter] = &cloudwatchlogs.InputLogEvent{
			Message:   &message,
			Timestamp: &timestamp,
		}
		if counter == *eventSize-1 {
			counter = 0
			if !*dryRun {
				nextToken, err = sendEventsCloudwatch(events, logGroup, logStream, nextToken, cloudwatchlogsClient)
				if err != nil {
					firstErrorLine := strings.Split(err.Error(), "\n")[0]
					splittedError := strings.Split(firstErrorLine, " ")
					nextToken, err = sendEventsCloudwatch(events, logGroup, logStream, &splittedError[len(splittedError)-1], cloudwatchlogsClient)
					if err != nil {
						log.Fatalf("%v", err)
					}
				}
			} else {
				sendEventsConsole(events)
			}
		}
		counter++
	}
	cmd.Wait()
}

func sendEventsCloudwatch(events []*cloudwatchlogs.InputLogEvent, logGroupName *string, logStreamName *string, nextToken *string, cloudwatchlogsClient *cloudwatchlogs.CloudWatchLogs) (*string, error) {

	sort.Sort(ByTimestamp(events))

	putLogEventInput := &cloudwatchlogs.PutLogEventsInput{
		LogEvents:     events,
		LogGroupName:  logGroupName,
		LogStreamName: logStreamName,
		SequenceToken: nextToken,
	}
	putLogEventsOutput, err := cloudwatchlogsClient.PutLogEvents(putLogEventInput)
	return putLogEventsOutput.NextSequenceToken, err
}

func sendEventsConsole(events []*cloudwatchlogs.InputLogEvent) {
	fmt.Printf("%v\n", events)
}

func getMessageTimestamp(m string) (string, int64) {
	secondsPart := m[:10]
	millisecondsPart := m[11:14]
	millisecondsString := secondsPart + millisecondsPart

	milliseconds, err := strconv.ParseInt(millisecondsString, 10, 64)
	if err != nil {
		log.Fatalf("Error in getMessageTimestamp\n%v\nMessage: %s\ntimestamp: %s\n", m, millisecondsString, err)
	}

	return m[18:], milliseconds
}

type ByTimestamp []*cloudwatchlogs.InputLogEvent

func (a ByTimestamp) Len() int           { return len(a) }
func (a ByTimestamp) Swap(i, j int)      { a[i], a[j] = a[j], a[i] }
func (a ByTimestamp) Less(i, j int) bool { return *a[i].Timestamp < *a[j].Timestamp }
