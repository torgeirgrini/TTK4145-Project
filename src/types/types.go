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
	OS_NONE OrderState = iota
	OS_UNCONFIRMED
	OS_CONFIRMED
	OS_COMPLETED
)

type HallCall struct {
	ExecutorID string
	AssignerID string
	OrderState OrderState
	AckList    []string
}

type NetworkMessage struct {
	ID        string
	HallCalls [][]HallCall
	ElevState Elevator
}

type Action struct {
	Dirn      elevio.MotorDirection
	Behaviour ElevatorBehaviour
}

type Elevator struct {
	Floor     int
	Dirn      elevio.MotorDirection
	Requests  [][]bool
	Behaviour ElevatorBehaviour
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
		ClearRequestVariant: CV_InDirn}
}