package fsm

import (
	"Project/config"
	"Project/localElevator/elevator"
	"Project/localElevator/elevio"
	"Project/localElevator/timer"
)

func RunElevator(){//channels som input
	//init elev:
	elev := elevator.InitElev();
	//set lights and door open lamp off

	//for-select med state machines for ulike sensorinput
} 

func Fsm_OnRequestButtonPress(btn_floor int, btn_type elevio.ButtonType, elev elevator.Elevator){
	switch elev.Behaviour{
	case elevator.EB_DoorOpen:
		if requests.Requests_shouldClearImmediately(elev, btn_floor, btn_type){
			timer.TimerStart(elev.DoorOpenDuration_s)
		} else{
			elev.Requests[btn_floor][int(btn_type)] = true //vetke om trenger int()
		}

	case elevator.EB_Moving:
		elev.Requests[btn_floor][int(btn_type)] = true 

	case elevator.EB_Idle:
		elev.Requests[btn_floor][int(btn_type)] = true
        a = requests.Requests_nextAction(elev)
        elev.Dirn = a.Dirn
        elev.Behaviour = a.Behaviour
		switch a.Behaviour{
		case elevator.EB_DoorOpen:
			elevio.SetDoorOpenLamp(true)
			timer.TimerStart(elev.DoorOpenDuration_s)
			elev = requests.Requests_clearAtCurrentFloor(elev)
		case elevator.EB_Moving:
			elevio.SetMotorDirection(elev.Dirn)
		
		//case elevator.EB_Idle:

		}
	}
	setAllLights(elev)

}




