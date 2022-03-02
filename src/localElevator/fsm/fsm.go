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

func setAllLights(Elevator es) {
	for floor := 0; floor < N_FLOORS; floor++ {
        for btn := 0; btn < N_BUTTONS; btn++ {
            elevio.SetButtonLamp(btn, floor, es.requests[floor][btn]);
        }
    }
}


func RunElevator (...) {

	for {

		select {
		case RequestButtonPress := <-ch.RequestButtonPress:
			//Eirik sin funksjon
			
		case FloorArrival := <-ch.FloorArrival:
			//print elevator
			elevio.SetFloorIndicator(elevator.floor)

			switch elev.Behaviour{
			case EB_Moving:
				if requests_shouldStop(elevator) { //Have orders in floor
					elevio.SetMotorDirection(MD_Stop)
					elevio.SetDoorOpenLamp(1)
					elev = requests_clearAtCurrentFloor(elev);
					timer.TimerStart(elev.doorOpenDuration_s)
					setAllLights(elev)
					elev.behaviour = DoorOpen
				}
			default:

			}

			//print elevator
		
		case DoorTimeOut := <- ch.DoorTimeOut:

			switch elev.behaviour {
			case EB_DoorOpen:
				Action a = requests.Requests_nextAction(elev)
				elev.dirn = a.dirn
				elev.behaviour = a.behaviour

				switch elev.behaviour {
				case EB_DoorOpen:
					timer.TimerStart(elev.config.doorOpenDuration_s)
					elev = requests.Requests_clearAtCurrentFloor(elev)
					setAllLights(elev)
				case EB_Moving:
				case EB_Idle:
					elevio.SetDoorOpenLamp(0)
					elevio.SetMotorDirection(elev.Dirn)
				}
			}
		}
	}
}

