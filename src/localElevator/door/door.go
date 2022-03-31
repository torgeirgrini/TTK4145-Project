package door

import (
	"Project/config"
	"Project/localElevator/elevio"
	"Project/types"
	"time"
)

func Door(ch_hwObstruction <-chan bool,
	ch_openDoor <-chan bool,
	ch_stuck chan<- bool,
	ch_doorClosed chan<- bool,
) {
	var doorState types.DoorState = types.DS_Closed
	elevio.SetDoorOpenLamp(false)
	obstruction := false

	DoorTimer := time.NewTimer(time.Duration(config.DoorOpenDuration_s) * time.Second)
	DoorTimer.Stop()
	ch_doorTimer := DoorTimer.C

	ObstructionTimer := time.NewTimer(time.Duration(config.TimeBeforeUnavailable) * time.Second)
	ObstructionTimer.Stop()
	ch_obstructionTimer := ObstructionTimer.C

	for {
		select {
		case obstruction = <-ch_hwObstruction:
			switch doorState {
			case types.DS_Open:
				doorState = types.DS_Obstructed
				ObstructionTimer.Reset(time.Duration(config.TimeBeforeUnavailable) * time.Second)
				DoorTimer.Stop()
			case types.DS_Closed:
			case types.DS_Obstructed:
				ch_stuck <- false
				doorState = types.DS_Open
				DoorTimer.Reset(time.Duration(config.DoorOpenDuration_s) * time.Second)
				ObstructionTimer.Stop()
			}
		case <-ch_openDoor:
			switch doorState {
			case types.DS_Open:
				DoorTimer.Reset(time.Duration(config.DoorOpenDuration_s) * time.Second)
			case types.DS_Closed:
				if obstruction {
					doorState = types.DS_Obstructed
					ObstructionTimer.Reset(time.Duration(config.TimeBeforeUnavailable) * time.Second)
					DoorTimer.Stop()
				} else {
					doorState = types.DS_Open
					DoorTimer.Reset(time.Duration(config.DoorOpenDuration_s) * time.Second)
					ObstructionTimer.Stop()
				}
				elevio.SetDoorOpenLamp(true)
			case types.DS_Obstructed:
			}
		case <-ch_obstructionTimer:
			ObstructionTimer.Stop()
			ch_stuck <- true
		case <-ch_doorTimer:
			doorState = types.DS_Closed
			elevio.SetDoorOpenLamp(false)
			ch_doorClosed <- true
		}
	}
}
