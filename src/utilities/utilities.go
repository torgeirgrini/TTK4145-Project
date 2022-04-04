package utilities

import (
	"Project/config"
	"Project/network/peers"
	"Project/types"
)

func DeepCopyElevatorStruct(e types.Elevator) types.Elevator {
	requestMatrix := make([][]bool, config.NumFloors)
	for floor := 0; floor < config.NumFloors; floor++ {
		requestMatrix[floor] = make([]bool, config.NumButtons)
		for button := range requestMatrix[floor] {
			requestMatrix[floor][button] = e.Requests[floor][button]
		}
	}
	return types.Elevator{
		Floor:               e.Floor,
		Dirn:                e.Dirn,
		Requests:            requestMatrix,
		Behaviour:           e.Behaviour,
		ClearRequestVariant: e.ClearRequestVariant}

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
		Lost:  DeepCopyStringSlice(peerstatus.Lost),
		New:   peerstatus.New,
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

func ContainsStringSlice(slice, subslice []string) bool {
	for _, element := range subslice {
		containselement := ContainsString(slice, element)
		if !containselement {
			return false
		}
	}
	return true
}

func ContainsString(sl []string, str string) bool {
	for _, value := range sl {
		if value == str {
			return true
		}
	}
	return false
}
