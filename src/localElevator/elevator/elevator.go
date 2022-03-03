package elevator

import (
	"Project/config"
	"Project/localElevator/elevio"
	"fmt"
)

type ElevatorBehaviour int

const (
	EB_Idle ElevatorBehaviour = iota //iota betyr auto 1,2,3,... osv oppover, standard
	EB_DoorOpen
	EB_Moving
)

//tror kanskje det er denne greia her som er ny for i år?:
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

	//vet ikke om dette bør være i egen struct i go?
	ClearRequestVariant ClearRequestVariant
	DoorOpenDuration_s  float64
}

// opprett ordrematrise, sett alle ordre til 0, behaviour til idle,
// floor til det nederste og motor til stans:
func InitElev() Elevator {
	elevio.SetMotorDirection(elevio.MD_Stop)
	requestMatrix := make([][]bool, config.NumFloors) //init tom 2d-slice
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
		DoorOpenDuration_s:  3}
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
