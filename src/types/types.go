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

type Elevator struct {
	Floor     int
	Dirn      elevio.MotorDirection
	Requests  [][]bool
	Behaviour ElevatorBehaviour
	// legg til avalible-bit?

	//vet ikke om dette bør være i egen struct i go? - HØR MED STUD.ASSER
	ClearRequestVariant ClearRequestVariant
	DoorOpenDuration_s  float64
}

type MsgToDistributor struct {
	OrderType elevio.ButtonEvent
	Elevators map[string]Elevator
}

type OrderState int

const (
	OS_NONE OrderState = iota
	OS_UNCONFIRMED
	OS_CONFIRMED
	OS_COMPLETED
)

type HallCall struct {
	executerID string
	assignerID string
	orderState OrderState
	ackList    [config.NumElevators]string
}

type ElevatorStateMessage struct {
	ID            string
	LocalElevator Elevator
	HallCalls     [][]HallCall
	ElevStateMap  map[string]Elevator
}

type Action struct {
	Dirn      elevio.MotorDirection
	Behaviour ElevatorBehaviour
}
