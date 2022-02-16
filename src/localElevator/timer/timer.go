package timer

import (
	"time"
)

func getWallTime() float64 {
	timeNow := time.Now()
	return float64(timeNow.Second()) + float64(timeNow.Nanosecond())/float64(1000000000)
}

var timerEndTime float64
var timerActive int

func TimerStart(duration float64) {
	timerEndTime = getWallTime() + duration
	timerActive = 1
}

func TimerStop() {
	timerActive = 0
}

func TimerTimedOut() int {
	timeRanOut := 0
	if getWallTime() > timerEndTime {
		timeRanOut = 1
	}
	return (timerActive & timeRanOut)
}
