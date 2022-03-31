package types

import (
	"Project/config"
	"Project/localElevator/elevio"
)

type ElevatorBehaviour int

const (
	EB_Idle ElevatorBehaviour = iota
	EB_DoorOpen
	EB_Moving
)

type ClearRequestVariant int

const (
	CV_All ClearRequestVariant = iota
	CV_InDirn
)

type AssignedOrder struct {
	OrderType elevio.ButtonEvent
	ID        string
}

type OrderState int

const (
	OS_Completed OrderState = iota
	OS_Unconfirmed
	OS_Confirmed
	OS_Unknown
)

type DoorState int

const (
	DS_Open DoorState = iota
	DS_Closed
	DS_Obstructed
)

type HallCall struct {
	ExecutorID 	string
	OrderState 	OrderState
	AckList    	[]string
}

type ElevStateNetMsg struct {
	SenderID    string
	ElevStateID string
	ElevState   Elevator
}

type HallCallsNetMsg struct{
	SenderID 	string
	HallCalls 	[][]HallCall
}

// type AssignerMessage struct {
// 	PeerStatus  peers.PeerUpdate
// 	ElevatorMap map[string]Elevator
// }


type Action struct {
	Dirn      elevio.MotorDirection
	Behaviour ElevatorBehaviour
}

type Elevator struct {
	Floor               int
	Dirn                elevio.MotorDirection
	Requests            [][]bool
	Behaviour           ElevatorBehaviour
	Available           bool
	ClearRequestVariant ClearRequestVariant
}

func InitElev() Elevator {
	requestMatrix := make([][]bool, config.NumFloors)
	for floor := 0; floor < config.NumFloors; floor++ {
		requestMatrix[floor] = make([]bool, config.NumButtons)
		for button := range requestMatrix[floor] {
			requestMatrix[floor][button] = false
		}
	}
	return Elevator{
		Floor:               0,
		Dirn:                elevio.MD_Stop,
		Requests:            requestMatrix,
		Behaviour:           EB_Idle,
		Available:           true,
		ClearRequestVariant: CV_InDirn}
}
