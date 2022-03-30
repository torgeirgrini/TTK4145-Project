package utilities

import (
	"Project/config"
	"Project/network/peers"
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
	e2.Available = e.Available
	e2.ClearRequestVariant = e.ClearRequestVariant
	e2.Requests = make([][]bool, len(e.Requests))
	for i := range e.Requests {
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
	for i := range copied {
		copied[i] = make([]types.HallCall, config.NumButtons-1)
		for j := range copied[i] {
			copied[i][j] = hallcalls[i][j]
		}
	}
	return copied
}

func DeepCopyStringSlice(slice []string) []string {
	copied := make([]string, len(slice))
	for i := range copied {
		copied[i] = slice[i]
	}
	return copied
}

func DeepCopyPeerStatus(peerstatus peers.PeerUpdate) peers.PeerUpdate {
	return peers.PeerUpdate{
		Peers: DeepCopyStringSlice(peerstatus.Peers),
		Lost: DeepCopyStringSlice(peerstatus.Lost),
		New: peerstatus.New,
	}
}

func DifferenceMatrix(m1 [][]bool, m2 [][]bool) [][]bool {
	DiffMatrix := make([][]bool, len(m1))
	for i := range m1 {
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

func EqualStringSlice(x, y []string) bool {
	if len(x) != len(y) {
		return false
	}
	diff := make(map[string]int, len(x))
	for _, _x := range x {
		diff[_x]++
	}
	for _, _y := range y {
		if _, ok := diff[_y]; !ok {
			return false
		}
		diff[_y] -= 1
		if diff[_y] == 0 {
			delete(diff, _y)
		}
	}
	return len(diff) == 0
}

func RemoveDuplicatesSlice(s []string) []string {
	inResult := make(map[string]bool)
	var result []string
	for _, str := range s {
		if _, ok := inResult[str]; !ok {
			inResult[str] = true
			result = append(result, str)
		}
	}
	return result
}