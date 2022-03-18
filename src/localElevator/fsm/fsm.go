package fsm

import (
	"Project/config"
	"Project/localElevator/elevator"
	"Project/localElevator/elevio"
	"Project/localElevator/requests"
	"Project/types"
	"fmt"
	"time"
)

func Fsm_OnInitBetweenFloors(e *types.Elevator) {
	elevio.SetMotorDirection(elevio.MD_Down)
	e.Dirn = elevio.MD_Down
	e.Behaviour = types.EB_Moving
}

func SetAllLights(elev types.Elevator) {
	for floor := 0; floor < config.NumFloors; floor++ {
		for btn := 0; btn < config.NumButtons; btn++ {
			elevio.SetButtonLamp(elevio.ButtonType(btn), floor, elev.Requests[floor][btn])
		}
	}
}

func Fsm_OnInitArrivedAtFloor(e *types.Elevator, currentFloor int) {
	elevio.SetMotorDirection(elevio.MD_Stop)
	e.Dirn = elevio.MD_Stop
	e.Behaviour = types.EB_Idle
	e.Floor = currentFloor
	elevio.SetFloorIndicator(currentFloor)
}

func RunElevator(
	ch_newAssignedOrder <-chan elevio.ButtonEvent,
	ch_FloorArrival <-chan int,
	ch_Obstruction <-chan bool,
	ch_localElevatorStruct chan<- types.Elevator) {

	//Initialize
	elev := elevator.InitElev()
	e := &elev
	SetAllLights(elev)
	elevio.SetDoorOpenLamp(false)

	Fsm_OnInitBetweenFloors(e)
	ch_localElevatorStruct <- *e

	currentFloor := <-ch_FloorArrival
	fmt.Println("Floor:", currentFloor)
	Fsm_OnInitArrivedAtFloor(e, currentFloor)
	ch_localElevatorStruct <- *e

	elevator.PrintElevator(elev)
	//Initialize Timers
	DoorTimer := time.NewTimer(time.Duration(config.DoorOpenDuration) * time.Second)
	DoorTimer.Stop()
	ch_doorTimer := DoorTimer.C
	RefreshStateTimer := time.NewTimer(time.Duration(config.RefreshStatePeriod) * time.Millisecond)
	ch_RefreshStateTimer := RefreshStateTimer.C
	//Elevator FSM
	var obstruction bool = false
	for {

		select {
		case newOrder := <-ch_newAssignedOrder:
			fmt.Println("Order {Floor, Type}:", newOrder)
			switch e.Behaviour {
			case types.EB_DoorOpen:
				if requests.Requests_shouldClearImmediately(*e, newOrder.Floor, newOrder.Button) {
					DoorTimer.Reset(time.Duration(config.DoorOpenDuration) * time.Second)
				} else {
					e.Requests[newOrder.Floor][int(newOrder.Button)] = true
				}

			case types.EB_Moving:
				e.Requests[newOrder.Floor][int(newOrder.Button)] = true

			case types.EB_Idle:
				e.Requests[newOrder.Floor][int(newOrder.Button)] = true
				action := requests.Requests_nextAction(*e)
				e.Dirn = action.Dirn
				e.Behaviour = action.Behaviour
				ch_localElevatorStruct <- *e
				switch action.Behaviour {
				case types.EB_DoorOpen:
					elevio.SetDoorOpenLamp(true)
					DoorTimer.Reset(time.Duration(config.DoorOpenDuration) * time.Second)
					requests.Requests_clearAtCurrentFloor(e)

				case types.EB_Moving:
					elevio.SetMotorDirection(e.Dirn)

				case types.EB_Idle:
					break
				}
			}
			SetAllLights(*e)
			elevator.PrintElevator(elev)

		case newFloor := <-ch_FloorArrival:
			fmt.Println("Floor:", newFloor)
			e.Floor = newFloor
			elevio.SetFloorIndicator(e.Floor)

			switch e.Behaviour {
			case types.EB_Moving:
				if requests.Requests_shouldStop(*e) {
					elevio.SetMotorDirection(elevio.MD_Stop)
					elevio.SetDoorOpenLamp(true)
					requests.Requests_clearAtCurrentFloor(e)
					DoorTimer.Reset(time.Duration(config.DoorOpenDuration) * time.Second)
					SetAllLights(*e)
					e.Behaviour = types.EB_DoorOpen
					ch_localElevatorStruct <- *e
				}

			default:
				break
			}

			elevator.PrintElevator(elev)

		case <-ch_doorTimer:
			if !obstruction {
				fmt.Println("Timer timed out")
				switch e.Behaviour {
				case types.EB_DoorOpen:
					action := requests.Requests_nextAction(*e)
					e.Dirn = action.Dirn
					e.Behaviour = action.Behaviour
					ch_localElevatorStruct <- *e

					switch e.Behaviour {
					case types.EB_DoorOpen:
						DoorTimer.Reset(time.Duration(config.DoorOpenDuration) * time.Second)
						requests.Requests_clearAtCurrentFloor(e)
						SetAllLights(*e)
					case types.EB_Moving:
						fallthrough
					case types.EB_Idle:
						elevio.SetDoorOpenLamp(false)
						elevio.SetMotorDirection(e.Dirn)
					}
				}

				elevator.PrintElevator(elev)
			}

		case obstruction = <-ch_Obstruction:
			if !obstruction {
				DoorTimer.Reset(time.Duration(config.DoorOpenDuration) * time.Second)
			}

		case <-ch_RefreshStateTimer:
			ch_localElevatorStruct <- *e
			RefreshStateTimer.Reset(time.Duration(config.RefreshStatePeriod) * time.Millisecond)
		}
	}
}
