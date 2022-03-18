package types

import (
	"Project/config"
	"Project/localElevator/elevio"
	"fmt"
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
		ClearRequestVariant: CV_InDirn,
		DoorOpenDuration_s:  config.DoorOpenDuration}
}

func PrintElevator(elev Elevator) {
	fmt.Println("Elevator: ")
	fmt.Println("	Current Floor: ", elev.Floor)
	fmt.Println("	Current Direction: ", elev.Dirn)
	fmt.Println("	Current Request Matrix: ")
	for floor := config.NumFloors - 1; floor > -1; floor-- {
		fmt.Println("		Orders at floor", floor, ": ", elev.Requests[floor])
	}
	fmt.Println("	Current Behaviour: ", elev.Behaviour)
}

func Dup(e Elevator) Elevator {
	e2 := InitElev()
	e2.Floor = e.Floor
	e2.Dirn = e.Dirn
	e2.Requests = e.Requests
	e2.Behaviour = e.Behaviour
	e2.ClearRequestVariant = e.ClearRequestVariant
	e2.DoorOpenDuration_s = e.DoorOpenDuration_s

	for floor := 0; floor < config.NumFloors; floor++ {
		for button, _ := range e.Requests[floor] {
			e2.Requests[floor][button] = e.Requests[floor][button]
		}
	}
	return e2
}

type MsgToDistributor struct {
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
	executerID string
	assignerID string
	orderState OrderState
	ackList    [config.NumElevators]string
}

type ElevatorStateMessage struct {
	ID        string
	HallCalls [][]HallCall
	ElevState Elevator
}

type Action struct {
	Dirn      elevio.MotorDirection
	Behaviour ElevatorBehaviour
}
