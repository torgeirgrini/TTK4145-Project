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

func SetCabLights(elev types.Elevator) {
	for floor := 0; floor < config.NumFloors; floor++ {
		elevio.SetButtonLamp(elevio.BT_Cab, floor, elev.Requests[floor][elevio.BT_Cab])
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
	ch_localElevatorState chan<- types.Elevator,
	ch_localOrderCompleted chan<- elevio.ButtonEvent,
	ch_peerTxEnable chan<- bool,
	ch_loneElevator <-chan bool,
) {

	//Initialize
	e := types.InitElev()
	SetCabLights(e)
	elevio.SetDoorOpenLamp(false)
	elevio.SetMotorDirection(elevio.MD_Stop)
	/*
		WatchDog := time.NewTimer(time.Duration(config.WatchDogBiteTime) * time.Second)
		WatchDog.Stop()
		ch_watchDog := WatchDog.C
	*/
	//WatchDog.Reset(time.Duration(config.WatchDogBiteTime) * time.Second)
	Fsm_OnInitBetweenFloors(&e)
	currentFloor := <-ch_hwFloor
	Fsm_OnInitArrivedAtFloor(&e, currentFloor)
	//WatchDog.Stop()
	//Initialize Timers
	DoorTimer := time.NewTimer(time.Duration(config.DoorOpenDuration_s) * time.Second)
	DoorTimer.Stop()
	ch_doorTimer := DoorTimer.C

	ObstructionTimer := time.NewTimer(time.Duration(config.TimeBeforeUnavailable) * time.Second)
	ObstructionTimer.Stop()
	ch_obstructionTimer := ObstructionTimer.C

	// RefreshStateTimer := time.NewTimer(time.Duration(config.RefreshStatePeriod_ms) * time.Millisecond)
	// ch_RefreshStateTimer := RefreshStateTimer.C
	//Elevator FSM
	var obstruction bool = false
	var loneElevator bool = true
	for {
		ch_localElevatorState <- utilities.DeepCopyElevatorStruct(e) //gir det mer mening å ha denne nederst??
		select {
		case newOrder := <-ch_newLocalOrder:
			fmt.Println("for2")
			switch e.Behaviour {
			case types.EB_DoorOpen:
				if requests.Requests_shouldClearImmediately(e, newOrder.Floor, newOrder.Button) {
					DoorTimer.Reset(time.Duration(config.DoorOpenDuration_s) * time.Second)
					if newOrder.Button != elevio.BT_Cab {
						ch_localOrderCompleted <- elevio.ButtonEvent{Floor: newOrder.Floor, Button: newOrder.Button}
						//Noe watchdog greier på clear immideatly?
					}
				} else {
					e.Requests[newOrder.Floor][int(newOrder.Button)] = true
				}

			case types.EB_Moving:
				e.Requests[newOrder.Floor][int(newOrder.Button)] = true

			case types.EB_Idle:

				e.Requests[newOrder.Floor][int(newOrder.Button)] = true
				action := requests.Requests_nextAction(e, newOrder) //!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!!
				e.Dirn = action.Dirn
				e.Behaviour = action.Behaviour

				switch action.Behaviour {
				case types.EB_DoorOpen:
					elevio.SetDoorOpenLamp(true)
					DoorTimer.Reset(time.Duration(config.DoorOpenDuration_s) * time.Second)
					requestCopy := utilities.DeepCopyElevatorStruct(e).Requests
					requests.Requests_clearAtCurrentFloor(&e)
					diff := utilities.DifferenceMatrix(requestCopy, e.Requests)
					for i := range diff {
						for j := 0; j < config.NumButtons-1; j++ {
							if diff[i][j] {
								ch_localOrderCompleted <- elevio.ButtonEvent{Floor: i, Button: elevio.ButtonType(j)}
							}
						}
					}
				case types.EB_Moving:
					elevio.SetMotorDirection(e.Dirn)
					//WatchDog.Reset(time.Duration(config.WatchDogBiteTime) * time.Second)

				case types.EB_Idle:
					break
				}
			}
			SetCabLights(e)

		case newFloor := <-ch_hwFloor:
			fmt.Println("Floor:", newFloor)
			e.Floor = newFloor
			elevio.SetFloorIndicator(e.Floor)

			switch e.Behaviour {
			case types.EB_Moving:
				if requests.Requests_shouldStop(e) {
					elevio.SetMotorDirection(elevio.MD_Stop)
					//WatchDog.Stop()
					elevio.SetDoorOpenLamp(true)
					requestCopy := utilities.DeepCopyElevatorStruct(e).Requests
					requests.Requests_clearAtCurrentFloor(&e)
					diff := utilities.DifferenceMatrix(requestCopy, e.Requests)
					for i := range diff {
						for j := 0; j < config.NumButtons-1; j++ {
							if diff[i][j] {
								ch_localOrderCompleted <- elevio.ButtonEvent{Floor: i, Button: elevio.ButtonType(j)}
							}
						}
					}
					DoorTimer.Reset(time.Duration(config.DoorOpenDuration_s) * time.Second)
					SetCabLights(e)
					e.Behaviour = types.EB_DoorOpen
				}

			default:
				break
			}

		case <-ch_doorTimer:
			fmt.Println("her1")
			//if !obstruction {
			fmt.Println("her2")
			switch e.Behaviour { //switch med bare en case?? Endre til if?
			case types.EB_DoorOpen:
				action := requests.Requests_nextAction(e, elevio.ButtonEvent{Floor: 0, Button: elevio.BT_Cab}) //litt for hard workaround?
				e.Dirn = action.Dirn
				e.Behaviour = action.Behaviour

				switch e.Behaviour {
				case types.EB_DoorOpen:
					DoorTimer.Reset(time.Duration(config.DoorOpenDuration_s) * time.Second)
					requestCopy := utilities.DeepCopyElevatorStruct(e).Requests
					requests.Requests_clearAtCurrentFloor(&e)
					diff := utilities.DifferenceMatrix(requestCopy, e.Requests)
					for i := range diff {
						for j := 0; j < config.NumButtons-1; j++ {
							if diff[i][j] {
								ch_localOrderCompleted <- elevio.ButtonEvent{Floor: i, Button: elevio.ButtonType(j)}
							}
						}
					}
					SetCabLights(e)
				case types.EB_Moving:
					fallthrough
				case types.EB_Idle:
					elevio.SetDoorOpenLamp(false)
					elevio.SetMotorDirection(e.Dirn)
					if e.Dirn != elevio.MD_Stop {
						//WatchDog.Reset(time.Duration(config.WatchDogBiteTime) * time.Second)
					}
				}
			}
			//}
			//vi vil sende unavab på nett når obst+dooropen lenger enn en viss tid
		case obstruction = <-ch_hwObstruction:
			fmt.Println("Obstruction val: ", obstruction) //venter her så lenge obstr er høy
			if !obstruction {                             //ikke lenger obstr
				fmt.Println("her3")
				DoorTimer.Reset(time.Duration(config.DoorOpenDuration_s) * time.Second)
				ch_peerTxEnable <- true
				ObstructionTimer.Stop()
			} else {
				fmt.Println("her4")
				DoorTimer.Stop()
				ObstructionTimer.Reset(time.Duration(config.TimeBeforeUnavailable) * time.Second)
			}
		case loneElevator = <-ch_loneElevator:
			fmt.Println(loneElevator)
		case <-ch_obstructionTimer:
			if !loneElevator {
				for floor := 0; floor < config.NumFloors; floor++ {
					for button := 0; button < config.NumButtons - 1; button++ {
						e.Requests[floor][button] = false
					}
				}
			}
			ch_peerTxEnable <- false

			//case <-ch_watchDog:
			//ch_peerTxEnable <- false
		}

	}
}
