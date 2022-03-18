package costfn

import (
	"Project/config"
	"Project/localElevator/elevio"
	"Project/localElevator/requests"
	"Project/types"
	"fmt"
	"time"
)

func TimeToIdle(e types.Elevator) int {
	duration := 0

	switch e.Behaviour {
	case types.EB_Idle:
		fmt.Println("IDLE")
		e.Dirn = requests.Requests_nextAction(e).Dirn
		if e.Dirn == elevio.MD_Stop {
			return duration
		}

	case types.EB_Moving:
		fmt.Println("moving")
		duration += int(time.Duration(config.TravelTime)*time.Millisecond) / 2
		e.Floor += int(e.Dirn)

	case types.EB_DoorOpen:
		fmt.Println("dooropen")
		duration -= int(time.Duration(config.DoorOpenDuration)*time.Second) / 2
	}

	for {
		if requests.Requests_shouldStop(e) {
			fmt.Println("forloop")
			requests.Requests_clearAtCurrentFloor(&e)
			duration += int(time.Duration(config.DoorOpenDuration)*time.Second)
			e.Dirn = requests.Requests_nextAction(e).Dirn
			if e.Dirn == elevio.MD_Stop {
				fmt.Println(duration)
				return duration

			}
		}
		e.Floor += int(e.Dirn)
		duration += int(time.Duration(config.TravelTime) * time.Millisecond)
	}
}
