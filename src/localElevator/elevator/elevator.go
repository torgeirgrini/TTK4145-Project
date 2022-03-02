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
	Requests [][]bool
	Behaviour ElevatorBehaviour
	// legg til avalible-bit?
	
	//vet ikke om dette bør være i egen struct i go?
	ClearRequestVariant ClearRequestVariant
	DoorOpenDuration_s float64
}

// opprett ordrematrise, sett alle ordre til 0, behaviour til idle,
// floor til det nederste og motor til stans:
func InitElev() Elevator{
	requestMatrix := make([][]bool, 0) //init tom 2d-slice
	for floor := 0; floor < config.NumFloors; floor++ {
		requestMatrix[floor] = make([]bool, config.NumButtons)
		for button := range requestMatrix[floor]{
			requestMatrix[floor][button] = false
		}
	}
	return Elevator{
		Floor: 		0,
		Dirn: 		elevio.MD_Stop,
		Requests: 	requestMatrix,
		Behaviour: 	EB_Idle,
		ClearRequestVariant: CV_InDirn,
		DoorOpenDuration_s: 0}
}






