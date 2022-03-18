package costfn

import (
	"Project/config"
	"Project/localElevator/elevator"
	"Project/localElevator/elevio"
	"Project/localElevator/requests"
	"time"
)


func timeToIdle(e elevator.Elevator) int {
	duration := 0

	switch e.Behaviour {
	case elevator.EB_Idle:
		e.Dirn = requests.Requests_nextAction(e).Dirn
		if e.Dirn == elevio.MD_Stop {
			return duration
		}

	case elevator.EB_Moving:
		duration += int(time.Duration(config.TravelTime) * time.Millisecond) / 2
		e.Floor += int(e.Dirn)

	case elevator.EB_DoorOpen:
		duration -= int(time.Duration(config.DoorOpenDuration)*time.Second) / 2
	}

	for {
		if requests.Requests_shouldStop(e) {
			requests.Requests_clearAtCurrentFloor(&e)
			duration += int(time.Duration(config.DoorOpenDuration)*time.Second)
			e.Dirn = requests.Requests_nextAction(e).Dirn
			if e.Dirn == elevio.MD_Down {
				return duration
			}
		}
		e.Floor += int(e.Dirn)
		duration += int(time.Duration(config.TravelTime) * time.Millisecond)
	}
}
