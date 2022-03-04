package fsm

import (
	"Project/config"
	"Project/localElevator/elevator"
	"Project/localElevator/elevio"
	"Project/localElevator/requests"
	"Project/localElevator/timer"
	"fmt"
)

func Fsm_OnRequestButtonPress(btn_floor int, btn_type elevio.ButtonType, e *elevator.Elevator) {
	switch e.Behaviour {
	case elevator.EB_DoorOpen:
		if requests.Requests_shouldClearImmediately(*e, btn_floor, btn_type) {
			timer.TimerStart(e.DoorOpenDuration_s)
		} else {
			e.Requests[btn_floor][int(btn_type)] = true
		}

	case elevator.EB_Moving:
		e.Requests[btn_floor][int(btn_type)] = true

	case elevator.EB_Idle:
		e.Requests[btn_floor][int(btn_type)] = true
		action := requests.Requests_nextAction(*e)
		e.Dirn = action.Dirn
		e.Behaviour = action.Behaviour
		switch action.Behaviour {
		case elevator.EB_DoorOpen:
			elevio.SetDoorOpenLamp(true)
			timer.TimerStart(e.DoorOpenDuration_s)
			requests.Requests_clearAtCurrentFloor(e)

		case elevator.EB_Moving:
			elevio.SetMotorDirection(e.Dirn)

		case elevator.EB_Idle:
			break
		}
	}
	SetAllLights(*e)
}

func Fsm_OnFloorArrival(newFloor int, e *elevator.Elevator) {
	e.Floor = newFloor
	elevio.SetFloorIndicator(e.Floor)

	switch e.Behaviour {
	case elevator.EB_Moving:
		if requests.Requests_shouldStop(*e) { 
			elevio.SetMotorDirection(elevio.MD_Stop)
			elevio.SetDoorOpenLamp(true)
			requests.Requests_clearAtCurrentFloor(e)
			timer.TimerStart(e.DoorOpenDuration_s)
			SetAllLights(*e)
			e.Behaviour = elevator.EB_DoorOpen
		}

	default:
		break
	}
}

func Fsm_OnDoorTimeout(e *elevator.Elevator) {
	switch e.Behaviour {
	case elevator.EB_DoorOpen:
		action := requests.Requests_nextAction(*e)
		e.Dirn = action.Dirn
		e.Behaviour = action.Behaviour

		switch e.Behaviour {
		case elevator.EB_DoorOpen:
			timer.TimerStart(e.DoorOpenDuration_s)
			requests.Requests_clearAtCurrentFloor(e)
			SetAllLights(*e)
		case elevator.EB_Moving:
			//handling?
		case elevator.EB_Idle:
			elevio.SetDoorOpenLamp(false)
			elevio.SetMotorDirection(e.Dirn)
		}
	}
}

func Fsm_OnInitBetweenFloors(e *elevator.Elevator) {
	elevio.SetMotorDirection(elevio.MD_Down)
	e.Dirn = elevio.MD_Down
	e.Behaviour = elevator.EB_Moving
}

func SetAllLights(elev elevator.Elevator) {
	for floor := 0; floor < config.NumFloors; floor++ {
		for btn := 0; btn < config.NumButtons; btn++ {
			elevio.SetButtonLamp(elevio.ButtonType(btn), floor, elev.Requests[floor][btn]) 
		}
	}
}

func Fsm_OnInitArrivedAtFloor(e *elevator.Elevator, currentFloor int) {
	elevio.SetMotorDirection(elevio.MD_Stop)
	e.Dirn = elevio.MD_Stop
	e.Behaviour = elevator.EB_Idle
}

func RunElevator(
	ch_RequestButtonPress chan elevio.ButtonEvent,
	ch_FloorArrival chan int,
	ch_DoorTimeOut chan bool,
	ch_Obstruction chan bool) {

	//Initialize
	elev := elevator.InitElev()
	e := &elev
	SetAllLights(elev)

	uninitialized := true
	for uninitialized {
		select {
		case currentFloor := <-ch_FloorArrival:
			fmt.Println("Floor:", currentFloor)
			Fsm_OnInitArrivedAtFloor(e, currentFloor)
			uninitialized = false
		default:
			Fsm_OnInitBetweenFloors(e)
		}
	}

	elevator.PrintElevator(elev)

	//Elevator FSM
	for {
		select {
		case newOrder := <-ch_RequestButtonPress:
			fmt.Println("Order {Floor, Type}:", newOrder)
			Fsm_OnRequestButtonPress(newOrder.Floor, newOrder.Button, e)
			elevator.PrintElevator(elev)

		case newFloor := <-ch_FloorArrival:
			fmt.Println("Floor:", newFloor)
			Fsm_OnFloorArrival(newFloor, e)
			elevator.PrintElevator(elev)

		case <-ch_DoorTimeOut:
			Fsm_OnDoorTimeout(e)

		case obstruction := <-ch_Obstruction:
			if (elev.Behaviour == elevator.EB_DoorOpen) && obstruction { //restart timer if obstruction while door open
				timer.TimerStart(elev.DoorOpenDuration_s)
			} 
		}
	}
}
