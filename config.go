package main

var Dispatcher2Conf Dispatcher2Config

func init() {

}

// Dispatcher2Config ...
type Dispatcher2Config struct {
	Dispatcher2Db             string
	MaxRetries                int
	MaxConcurrent             int
	ServerPort                int
	LogDIR                    string
	DefaultQueueStatus        string
	StartOfSubmissionPeriod   int
	EndOfSubmissionPeriod     int
	UseSSL                    string
	UseGlobalSubmissionPeriod string
	RequestProcessInterval    int
}
