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
	e2.Behaviour = e.Behaviour
	e2.ClearRequestVariant = e.ClearRequestVariant
	e2.Requests = make([][]bool, len(e.Requests))
	for i := range e.Requests{
    	e2.Requests[i] = make([]bool, len(e.Requests[i]))
    	copy(e2.Requests[i], e.Requests[i])
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

func DeepCopyStringSlice(slice []string, length int) []string {
	copied := make([]string, length)
	for i, _ := range copied {
		copied[i] = slice[i]
	}
	return copied
}

func DifferenceMatrix(m1 [][]bool, m2 [][]bool) [][]bool{
	DiffMatrix := make([][]bool, len(m1))
	for i := range m1{
		DiffMatrix[i] = make([]bool, len(m1[1]))
		for j := range m1[i] {
			if m1[i][j] != m2[i][j] {
				DiffMatrix[i][j] = true
			} else {
				DiffMatrix[i][j] = false
			}
		}
	}
	return DiffMatrix
}
