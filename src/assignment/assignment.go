package assigner

import (
	"Project/config"
	"Project/distribution"
	"Project/localElevator/elevio"
	"Project/types"
	"fmt"
)

// // Struct members must be public in order to be accessible by json.Marshal/.Unmarshal
// // This means they must start with a capital letter, so we need to use field renaming struct tags to make them camelCase

//Denne tar inn
//Buttonpress fra hardware IO
//Den får periodisk inn ESM fra distributor(main akkurat nå)
//Peer list for å vite hvilke heiser som er på nettet og som kan ta ordre (Denne burde kanskje sendes periodisk)

func Assignment(ch_allElevators <-chan map[string]types.Elevator,
	ch_newLocalOrder <-chan elevio.ButtonEvent,
	ch_newOrderAssigned chan<- types.MsgToDistributor) {

	elevators := make(map[string]types.Elevator)
	btn_event := elevio.ButtonEvent{}

	for {
		select {
		case elevators = <-ch_allElevators:
			fmt.Println("All elevator states:", elevators)

			//Videresend til assigner sånn at den kan regne ut
		case btn_event = <-ch_newLocalOrder:
			//Here we need to calcilate the cost functin
			//TEST START
			_ = btn_event
			assignedID := "id3"
			OrderFloor := 2
			OrderButton := elevio.BT_HallUp
			Ordertype_test := elevio.ButtonEvent{OrderFloor, OrderButton}
			elevators[assignedID].Requests[OrderFloor][OrderButton] = true
			send_msg := types.MsgToDistributor{Ordertype_test, distribution.DeepCopy(elevators)}
			ch_newOrderAssigned <- send_msg
			//TEST END
		}
	}
}

func HallRequestsFromESM(allElevators map[string]types.Elevator) [][]bool {
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
