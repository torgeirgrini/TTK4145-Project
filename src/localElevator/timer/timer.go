package timer

import (
	"time"
)

func getWallTime() float64 {
	timeNow := time.Now()
	return float64(timeNow.Second()) + float64(timeNow.Nanosecond())/float64(1000000000)
}

var timerEndTime float64

func TimerStart(duration float64) {
	timerEndTime = getWallTime() + duration
}

func PollTimerTimedOut(ch_DoorTimeOut chan bool) {
	for {
		if getWallTime() > timerEndTime {
			ch_DoorTimeOut <- true
		}
	}
}

/*
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

func TimerStop() { //kalles denne funksjonen noen gang?
	timerActive = false
}

func PollTimerTimedOut(ch_DoorTimeOut chan bool) {
	for {
		timeRanOut := false
		if getWallTime() > timerEndTime {
			timeRanOut = true
		}
		if timerActive && timeRanOut {
			ch_DoorTimeOut <- true
		}
	}
}
*/
