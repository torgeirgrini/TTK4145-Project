package fsm

import (
	"Project/config"
	"Project/localElevator/elevator"
	"Project/localElevator/elevio"
	"Project/localElevator/requests"
	"fmt"
	"time"
)

func Fsm_OnRequestButtonPress(btn_floor int, btn_type elevio.ButtonType, e *elevator.Elevator) {
	switch e.Behaviour {
	case elevator.EB_DoorOpen:
		if requests.Requests_shouldClearImmediately(*e, btn_floor, btn_type) {
			e.DoorTimer.Reset(time.Duration(config.DoorOpenDuration) * time.Second)
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
			e.DoorTimer.Reset(time.Duration(config.DoorOpenDuration) * time.Second)
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
			e.DoorTimer.Reset(time.Duration(config.DoorOpenDuration) * time.Second)
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
			e.DoorTimer.Reset(time.Duration(config.DoorOpenDuration) * time.Second)
			requests.Requests_clearAtCurrentFloor(e)
			SetAllLights(*e)
		case elevator.EB_Moving:
			fallthrough
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
	e.Floor = currentFloor
	elevio.SetFloorIndicator(currentFloor)
}

func RunElevator(
	ch_RequestButtonPress  chan elevio.ButtonEvent,
	ch_FloorArrival        chan int,
	ch_Obstruction         chan bool) { //husk å definer read/write med piler på channelsene, se forelesning

	//Initialize
	elev := elevator.InitElev()
	e := &elev
	SetAllLights(elev)
	elevio.SetDoorOpenLamp(false)

	Fsm_OnInitBetweenFloors(e)

	currentFloor := <-ch_FloorArrival
	fmt.Println("Floor:", currentFloor)
	Fsm_OnInitArrivedAtFloor(e, currentFloor)

	elevator.PrintElevator(elev)
	//Initialize Timer
	ch_doorTimer := e.DoorTimer.C
	//Elevator FSM
	var obstruction bool = false
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

		case <-ch_doorTimer:
			if !obstruction {
				fmt.Println("Timer timed out")
				Fsm_OnDoorTimeout(e)
				elevator.PrintElevator(elev)
			}

		case obstruction = <-ch_Obstruction:
			if !obstruction {
				e.DoorTimer.Reset(time.Duration(config.DoorOpenDuration) * time.Second)
			}
		}
	}
}
