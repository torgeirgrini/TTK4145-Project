package timer

import (
	"time"
)

// tror denne modulen må endres ganske mye, vi må sende en bool på
// DoorTimeOut-kanalen for at det skal kunne utløse noe handling i fsm-en,
// Da må vi ta inn den kanalen som inputparameter til funksjonene tror jeg..(?)

func getWallTime() float64 {
	timeNow := time.Now()
	return float64(timeNow.Second()) + float64(timeNow.Nanosecond())/float64(1000000000)
}

var timerEndTime float64
var timerActive bool

func TimerStart(duration float64) {
	timerEndTime = getWallTime() + duration
	timerActive = true
}

func TimerStop() {
	timerActive = false
}

func PollTimerTimedOut(ch_DoorTimeOut chan bool) {
	timeRanOut := false
	if getWallTime() > timerEndTime {
		timeRanOut = true
	}
	//må snde på kanalen hvis andre moduler skal få vite om det:
	if timerActive && timeRanOut {
		ch_DoorTimeOut <- true
	}
}
