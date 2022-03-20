package utilities

import (
	"Project/config"
	"Project/types"
	"fmt"
)

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

func DeepCopyElevatorStruct(e types.Elevator) types.Elevator {
	e2 := types.InitElev()
	e2.Floor = e.Floor
	e2.Dirn = e.Dirn
	e2.Requests = e.Requests
	e2.Behaviour = e.Behaviour
	e2.ClearRequestVariant = e.ClearRequestVariant

	for floor := 0; floor < config.NumFloors; floor++ {
		for button, _ := range e.Requests[floor] {
			e2.Requests[floor][button] = e.Requests[floor][button]
		}
	}
	return e2
}

func DeepCopyElevatorMap(elevators map[string]types.Elevator) map[string]types.Elevator {
	copied := make(map[string]types.Elevator)
	for i, e := range elevators {
		copied[i] = DeepCopyElevatorStruct(e)
	}
	return copied
}

func DeepCopyHallCalls(hallcalls [][]types.HallCall) [][]types.HallCall {
	copied := make([][]types.HallCall, config.NumFloors)
	for i, _ := range copied {
		copied[i] = make([]types.HallCall, config.NumButtons-1)
		for j, _ := range copied[i] {
			copied[i][j] = hallcalls[i][j]
		}
	}
	return copied
}
