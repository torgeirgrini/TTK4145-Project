package fsm

import (
	"Project/config"
	"Project/localElevator/elevio"
	"Project/localElevator/requests"
	"Project/types"
	"Project/utilities"
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

func RunLocalElevator(
	ch_newLocalOrder <-chan elevio.ButtonEvent,
	ch_hwFloor <-chan int,
	ch_hwObstruction <-chan bool,
	ch_localElevatorState chan<- types.Elevator) {

	//Initialize
	e := types.InitElev()
	SetAllLights(e)
	elevio.SetDoorOpenLamp(false)
	elevio.SetMotorDirection(elevio.MD_Stop)

	Fsm_OnInitBetweenFloors(&e)
	ch_localElevatorState <- e

	currentFloor := <-ch_hwFloor
	Fsm_OnInitArrivedAtFloor(&e, currentFloor)
	ch_localElevatorState <- e

	//Initialize Timers
	DoorTimer := time.NewTimer(time.Duration(config.DoorOpenDuration) * time.Second)
	DoorTimer.Stop()
	ch_doorTimer := DoorTimer.C
	// RefreshStateTimer := time.NewTimer(time.Duration(config.RefreshStatePeriod) * time.Millisecond)
	// ch_RefreshStateTimer := RefreshStateTimer.C
	//Elevator FSM
	var obstruction bool = false
	for {
		//types.PrintElevator(e)
		ch_localElevatorState <- utilities.DeepCopyElevatorStruct(e)

		select {
		case newOrder := <-ch_newLocalOrder:
			fmt.Println("Order received: Order {Floor, Type}:", newOrder)
			switch e.Behaviour {
			case types.EB_DoorOpen:
				if requests.Requests_shouldClearImmediately(e, newOrder.Floor, newOrder.Button) {
					DoorTimer.Reset(time.Duration(config.DoorOpenDuration) * time.Second)
				} else {
					e.Requests[newOrder.Floor][int(newOrder.Button)] = true
				}

			case types.EB_Moving:
				e.Requests[newOrder.Floor][int(newOrder.Button)] = true

			case types.EB_Idle:
				e.Requests[newOrder.Floor][int(newOrder.Button)] = true
				action := requests.Requests_nextAction(e)
				e.Dirn = action.Dirn
				e.Behaviour = action.Behaviour
				switch action.Behaviour {
				case types.EB_DoorOpen:
					elevio.SetDoorOpenLamp(true)
					DoorTimer.Reset(time.Duration(config.DoorOpenDuration) * time.Second)
					requests.Requests_clearAtCurrentFloor(&e)

				case types.EB_Moving:
					elevio.SetMotorDirection(e.Dirn)

				case types.EB_Idle:
					break
				}
			}
			SetAllLights(e)

		case newFloor := <-ch_hwFloor:
			fmt.Println("Floor:", newFloor)
			e.Floor = newFloor
			elevio.SetFloorIndicator(e.Floor)

			switch e.Behaviour {
			case types.EB_Moving:
				if requests.Requests_shouldStop(e) {
					elevio.SetMotorDirection(elevio.MD_Stop)
					elevio.SetDoorOpenLamp(true)
					requests.Requests_clearAtCurrentFloor(&e)
					DoorTimer.Reset(time.Duration(config.DoorOpenDuration) * time.Second)
					SetAllLights(e)
					e.Behaviour = types.EB_DoorOpen
				}

			default:
				break
			}

		case <-ch_doorTimer:
			if !obstruction {
				switch e.Behaviour {
				case types.EB_DoorOpen:
					action := requests.Requests_nextAction(e)
					e.Dirn = action.Dirn
					e.Behaviour = action.Behaviour

					switch e.Behaviour {
					case types.EB_DoorOpen:
						DoorTimer.Reset(time.Duration(config.DoorOpenDuration) * time.Second)
						requests.Requests_clearAtCurrentFloor(&e)
						SetAllLights(e)
					case types.EB_Moving:
						fallthrough
					case types.EB_Idle:
						elevio.SetDoorOpenLamp(false)
						elevio.SetMotorDirection(e.Dirn)
					}
				}

			}

		case obstruction = <-ch_hwObstruction:
			if !obstruction {
				DoorTimer.Reset(time.Duration(config.DoorOpenDuration) * time.Second)
			}

		}
	}
}
