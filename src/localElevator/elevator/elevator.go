package elevator

import (
	"Project/config"
	"Project/localElevator/elevio"
	"Project/localElevator/requests"
	"Project/types"
	"Project/utilities"
)

func LocalElevator(
	ch_newLocalOrder 	  <-chan   elevio.ButtonEvent,
	ch_hwFloor 			  <-chan   int,
	ch_localElevatorState 	chan<- types.Elevator,
	ch_localOrderCompleted 	chan<- elevio.ButtonEvent,
	ch_openDoor 			chan<- bool,
	ch_doorClosed 		  <-chan   bool,
	ch_setMotorDirn			chan<- elevio.MotorDirection,
) {

	e := InitElev()
	SetCabLights(e)

	e.Dirn = elevio.MD_Down
	e.Behaviour = types.EB_Moving
	ch_setMotorDirn <- e.Dirn

	e.Floor = <-ch_hwFloor
	e.Dirn = elevio.MD_Stop
	e.Behaviour = types.EB_Idle
	ch_setMotorDirn <- e.Dirn
	elevio.SetFloorIndicator(e.Floor)

	for {
		ch_localElevatorState <- utilities.DeepCopyElevatorStruct(e)
		select {
		case newOrder := <-ch_newLocalOrder:
			switch e.Behaviour {
			case types.EB_DoorOpen:
				if requests.Requests_shouldClearImmediately(e, newOrder.Floor, newOrder.Button) {
					ch_openDoor <- true
					if newOrder.Button != elevio.BT_Cab {
						ch_localOrderCompleted <- elevio.ButtonEvent{Floor: newOrder.Floor, Button: newOrder.Button}
					}
				} else {
					e.Requests[newOrder.Floor][int(newOrder.Button)] = true
				}
			case types.EB_Moving:
				e.Requests[newOrder.Floor][int(newOrder.Button)] = true
			case types.EB_Idle:
				e.Requests[newOrder.Floor][int(newOrder.Button)] = true
				action := requests.Requests_nextAction(e, newOrder.Button)
				e.Dirn = action.Dirn
				e.Behaviour = action.Behaviour
				switch action.Behaviour {
				case types.EB_DoorOpen:
					ch_openDoor <- true
					requestsBeforeClear := utilities.DeepCopyElevatorStruct(e).Requests
					e = requests.Requests_clearAtCurrentFloor(e)
					sendLocalCompletedOrders(requestsBeforeClear, e.Requests, ch_localOrderCompleted)
				case types.EB_Moving:
					ch_setMotorDirn <- e.Dirn
				case types.EB_Idle:
					break
				}
			}
			SetCabLights(e)
		case newFloor := <-ch_hwFloor:
			e.Floor = newFloor
			elevio.SetFloorIndicator(e.Floor)
			switch e.Behaviour {
			case types.EB_Moving:
				if requests.Requests_shouldStop(e) {
					ch_setMotorDirn <- elevio.MD_Stop
					requestsBeforeClear := utilities.DeepCopyElevatorStruct(e).Requests
					e = requests.Requests_clearAtCurrentFloor(e)
					sendLocalCompletedOrders(requestsBeforeClear, e.Requests, ch_localOrderCompleted)
					ch_openDoor <- true
					SetCabLights(e)
					e.Behaviour = types.EB_DoorOpen
				}
			default:
				break
			}
		case <-ch_doorClosed:
			switch e.Behaviour {
			case types.EB_DoorOpen:
				action := requests.Requests_nextAction(e, elevio.BT_Cab)
				e.Dirn = action.Dirn
				e.Behaviour = action.Behaviour
				switch e.Behaviour {
				case types.EB_DoorOpen:
					ch_openDoor <- true
					requestsBeforeClear := utilities.DeepCopyElevatorStruct(e).Requests
					e = requests.Requests_clearAtCurrentFloor(e)
					sendLocalCompletedOrders(requestsBeforeClear, e.Requests, ch_localOrderCompleted)
					SetCabLights(e)
				case types.EB_Moving:
					fallthrough
				case types.EB_Idle:
					ch_setMotorDirn <- e.Dirn
				}
			case types.EB_Moving:
			case types.EB_Idle:
			}
		}
	}
}

func InitElev() types.Elevator {
	requestMatrix := make([][]bool, config.NumFloors)
	for floor := 0; floor < config.NumFloors; floor++ {
		requestMatrix[floor] = make([]bool, config.NumButtons)
		for button := range requestMatrix[floor] {
			requestMatrix[floor][button] = false
		}
	}
	return types.Elevator{
		Floor:               0,
		Dirn:                elevio.MD_Stop,
		Requests:            requestMatrix,
		Behaviour:           types.EB_Idle,
		ClearRequestVariant: types.CV_InDirn}
}

func sendLocalCompletedOrders(reqBeforeClear [][]bool, reqAfterClear [][]bool, ch_localOrderCompleted chan<- elevio.ButtonEvent) {
	diff := utilities.DifferenceMatrix(reqBeforeClear, reqAfterClear)
	for i := range diff {
		for j := 0; j < config.NumButtons-1; j++ {
			if diff[i][j] {
				ch_localOrderCompleted <- elevio.ButtonEvent{Floor: i, Button: elevio.ButtonType(j)}
			}
		}
	}
}

func SetCabLights(elev types.Elevator) {
	for floor := 0; floor < config.NumFloors; floor++ {
		elevio.SetButtonLamp(elevio.BT_Cab, floor, elev.Requests[floor][elevio.BT_Cab])
	}
}
