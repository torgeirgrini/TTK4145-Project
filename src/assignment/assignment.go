package assigner

import (
	"Project/config"
	"Project/localElevator/elevator"
	"fmt"
)

// // Struct members must be public in order to be accessible by json.Marshal/.Unmarshal
// // This means they must start with a capital letter, so we need to use field renaming struct tags to make them camelCase

//Denne tar inn
//Buttonpress fra hardware IO
//Den får periodisk inn ESM fra distributor(main akkurat nå)
//Peer list for å vite hvilke heiser som er på nettet og som kan ta ordre (Denne burde kanskje sendes periodisk)

func Assignment(allElevators <-chan map[string]elevator.Elevator) {
	for {
		select {
		case a := <-allElevators:
			fmt.Println("All elevator states:", a)
			/*
				HRAElevStateArray := make(map[string]HRAElevState)
				for i,e := range(a) {
					cabreq := make([]bool, config.NumFloors)
					for j := 0; j < config.NumFloors; j++ {
						cabreq[j] = e.Requests[j][elevio.BT_Cab]
					}
					HRAes := HRAElevState{e.Behaviour, e.Floor,e.Dirn,cabreq}
					HRAElevStateArray[]
				}
				HallCalls := hallRequestsFromESM(a)
				CostFnInput := HRAInput{HallCalls,}*/
			//Videresend til assigner sånn at den kan regne ut
		}
	}
}

func hallRequestsFromESM(allElevators map[string]elevator.Elevator) [][]bool {
	//var HallCalls [][]HallCall
	Hallcalls := make([][]bool, config.NumFloors)
	for i := 0; i < config.NumFloors; i++ {
		Hallcalls[i] = make([]bool, config.NumButtons-1)
		for j := 0; j < config.NumButtons-1; j++ {
			Hallcalls[i][j] = false
			for _, e := range allElevators {
				Hallcalls[i][j] = Hallcalls[i][j] || e.Requests[i][j]
			}
		}
	}
	return Hallcalls
}
