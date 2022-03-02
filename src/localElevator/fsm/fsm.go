package fsm

import (
	"Project/config"
	"Project/localElevator/elevator"
	"Project/localElevator/elevio"
	"Project/localElevator/timer"
	"Project/localElevator/requests"
)


func Fsm_OnRequestButtonPress(btn_floor int, btn_type elevio.ButtonType, elev elevator.Elevator){
	switch elev.Behaviour{
	case elevator.EB_DoorOpen:
		if requests.Requests_shouldClearImmediately(elev, btn_floor, btn_type){
			timer.TimerStart(elev.DoorOpenDuration_s)
		} else{
			elev.Requests[btn_floor][int(btn_type)] = true //vetke om trenger int()
		}
		break

	case elevator.EB_Moving:
		elev.Requests[btn_floor][int(btn_type)] = true 
		break

	case elevator.EB_Idle:
		elev.Requests[btn_floor][int(btn_type)] = true
        action := requests.Requests_nextAction(elev)
        elev.Dirn = action.Dirn
        elev.Behaviour = action.Behaviour
		switch action.Behaviour{
		case elevator.EB_DoorOpen:
			elevio.SetDoorOpenLamp(true)
			timer.TimerStart(elev.DoorOpenDuration_s)
			elev = requests.Requests_clearAtCurrentFloor(elev)
			break
		
		case elevator.EB_Moving:
			elevio.SetMotorDirection(elev.Dirn)
			break
		
		case elevator.EB_Idle:
			break
		}
		break
	}
	SetAllLights(elev)
}


func Fsm_OnFloorArrival(newFloor int, elev elevator.Elevator){
	elev.Floor = newFloor
	elevio.SetFloorIndicator(elev.Floor)

	switch elev.Behaviour{
	case elevator.EB_Moving:
		if requests.Requests_shouldStop(elev) { //Have orders in floor
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetDoorOpenLamp(true)
			elev = requests.Requests_clearAtCurrentFloor(elev);
			timer.TimerStart(elev.DoorOpenDuration_s)
			SetAllLights(elev)
			elev.Behaviour = elevator.EB_DoorOpen
		}
		break

	default:
		break
	}
}

func Fsm_OnDoorTimeout(elev elevator.Elevator){
	switch elev.Behaviour {
	case elevator.EB_DoorOpen:
		action := requests.Requests_nextAction(elev)
		elev.Dirn = action.Dirn
		elev.Behaviour = action.Behaviour

		switch elev.Behaviour {
		case elevator.EB_DoorOpen:
			timer.TimerStart(elev.DoorOpenDuration_s)
			elev = requests.Requests_clearAtCurrentFloor(elev)
			SetAllLights(elev)
			break
		case elevator.EB_Moving:
			//skal det være noe her? føler vi kan få udefinert oppførsel eller noe
		case elevator.EB_Idle:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(elev.Dirn)
			break
		}
		break
	}
}

func Fsm_OnInitBetweenFloors(elev elevator.Elevator){
    elevio.SetMotorDirection(elevio.MD_Down)
    elev.Dirn = elevio.MD_Down
    elev.Behaviour = elevator.EB_Moving
}

func SetAllLights(elev elevator.Elevator) {
	for floor := 0; floor < config.NumFloors; floor++ {
        for btn := 0; btn < config.NumButtons; btn++ {
            elevio.SetButtonLamp(elevio.ButtonType(btn), floor, elev.Requests[floor][btn]); //vet ikke om denne kan ta inn int som første param
        }
    }
}


func RunElevator(
	ch_RequestButtonPress chan elevio.ButtonEvent,
	ch_FloorArrival chan int,
	ch_DoorTimeOut chan bool,
	ch_Obstruction chan bool) {

	elev := elevator.InitElev();

	
	for {
		select {
		case newOrder := <-ch_RequestButtonPress:
			Fsm_OnRequestButtonPress(newOrder.Floor, newOrder.Button, elev)
			
		case newFloor := <-ch_FloorArrival:
			Fsm_OnFloorArrival(newFloor, elev)
		
		case <- ch_DoorTimeOut:
			Fsm_OnDoorTimeout(elev)
		
		case obstruction := <-ch_Obstruction:
			if (elev.Behaviour == elevator.EB_DoorOpen) && obstruction{
				timer.TimerStart(elev.DoorOpenDuration_s) //er dette riktig oppførsel? er rtøtt
			}	// endre det over når fikset time-modul
		}
	}
}