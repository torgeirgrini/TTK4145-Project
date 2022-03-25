package costfn

import (
	"Project/config"
	"Project/localElevator/elevio"
	"Project/localElevator/requests"
	"Project/types"
	"time"
)


func TimeToIdle(e types.Elevator) int {
	duration := 0

	switch e.Behaviour {
	case types.EB_Idle:
		e.Dirn = requests.Requests_nextAction(e, elevio.ButtonEvent{Floor:-1, Button: elevio.BT_Cab}).Dirn
		if e.Dirn == elevio.MD_Stop {
			return duration
		}

	case types.EB_Moving:
		duration += int(time.Duration(config.TravelTime_ms)*time.Millisecond) / 2
		e.Floor += int(e.Dirn)

	case types.EB_DoorOpen:
		duration -= int(time.Duration(config.DoorOpenDuration_s)*time.Second) / 2
	}

	for {
		if requests.Requests_shouldStop(e) {
			requests.Requests_clearAtCurrentFloor(&e)
			duration += int(time.Duration(config.DoorOpenDuration_s) * time.Second)
			e.Dirn = requests.Requests_nextAction(e, elevio.ButtonEvent{Floor:-1, Button: elevio.BT_Cab}).Dirn
			if e.Dirn == elevio.MD_Stop {
				return duration

			}
		}
		e.Floor += int(e.Dirn)
		duration += int(time.Duration(config.TravelTime_ms) * time.Millisecond)
	}
}
