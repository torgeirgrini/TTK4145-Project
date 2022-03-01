package elevator

import (
	"Project/localElevator/elevio"
	"Project/config"
)

type ElevatorBehaviour int

const(
	EB_Idle ElevatorBehaviour = iota //iota betyr auto 1,2,3,... osv oppover, standard
	EB_DoorOpen 			  
	EB_Moving				  
)

//tror kanskje det er denne greia her som er ny for i år?:
type ClearRequestVariant int

const(
	CV_All ClearRequestVariant = iota
	CV_InDirn
)

type Elevator struct{
	Floor int
	Dirn elevio.MotorDirection
	Requests [][]int //bool isteden(?), 
	Behaviour ElevatorBehaviour
	// legg til avalible-bit?
	// vet ikke om vi trenger struct i struct (under)?
	// vet heller ikke helt hvordan å implementere i go, kommenterer ut foreløpig
	/*
	Config struct{
		clearRequestVariant ClearRequestVariant
		DoorOpenDuration_s float64
	}
	*/
}

// opprett ordrematrise, sett alle ordre til 0, behaviour til idle,
// floor til det nederste og motor til stans:
func initElev() Elevator{
	requestMatrix := make([][]int, 0) //init tom 2d-slice
	for floor := 0; floor < config.NumFloors; floor++ {
		requestMatrix[floor] = make([]int, config.NumButtons)
		for button := range requestMatrix[floor]{
			requestMatrix[floor][button] = 0
		}
	}
	return Elevator{
		Floor: 		0,
		Dirn: 		elevio.MD_Stop,
		Requests: 	requestMatrix,
		Behaviour: 	EB_Idle}
		//timer shit når funnet ut av det
}






