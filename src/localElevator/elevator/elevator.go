package elevator

import (
	"Project/config"
	"Project/localElevator/elevio"
	"Project/types"
	"fmt"
)

func InitElev() types.Elevator {
	elevio.SetMotorDirection(elevio.MD_Stop)
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
		ClearRequestVariant: types.CV_InDirn,
		DoorOpenDuration_s:  config.DoorOpenDuration}
}

func PrintElevator(elev types.Elevator) {
	fmt.Println("Elevator: ")
	fmt.Println("	Current Floor: ", elev.Floor)
	fmt.Println("	Current Direction: ", elev.Dirn)
	fmt.Println("	Current Request Matrix: ")
	for floor := config.NumFloors - 1; floor > -1; floor-- {
		fmt.Println("		Orders at floor", floor, ": ", elev.Requests[floor])
	}
	fmt.Println("	Current Behaviour: ", elev.Behaviour)
}
