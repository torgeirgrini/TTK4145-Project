package assigner

import (
	"Project/assignment/costfn"
	"Project/config"
	"Project/localElevator/elevio"
	"Project/network/peers"
	"Project/types"
	"Project/utilities"
	"fmt"
)

// // Struct members must be public in order to be accessible by json.Marshal/.Unmarshal
// // This means they must start with a capital letter, so we need to use field renaming struct tags to make them camelCase

//Denne tar inn
//Buttonpress fra hardware IO
//Den får periodisk inn ESM fra distributor(main akkurat nå)
//Peer list for å vite hvilke heiser som er på nettet og som kan ta ordre (Denne burde kanskje sendes periodisk)

func Assignment(
	localID string,
	ch_informationToAssigner <-chan types.AssignerMessage,
	ch_hwButtonPress <-chan elevio.ButtonEvent,
	ch_assignedOrder chan<- types.AssignedOrder,
) {

	var elevatorMap map[string]types.Elevator
	var peerUpdate peers.PeerUpdate

	//wait for distribution to send information before assigning
	assignerMsg := types.AssignerMessage{}
	assignerMsg = <-ch_informationToAssigner

	elevatorMap = utilities.DeepCopyElevatorMap(assignerMsg.ElevatorMap) 
	peerUpdate = assignerMsg.PeerStatus 
	
	for {

	
		select {
			//update information from distribtion
		case assignerMsg = <-ch_informationToAssigner:
			elevatorMap = utilities.DeepCopyElevatorMap(assignerMsg.ElevatorMap) 
			peerUpdate = assignerMsg.PeerStatus

			//hardware button press
		case btn_event := <-ch_hwButtonPress:

			//if cab call
			if btn_event.Button == elevio.BT_Cab {
				//send directly to distributor
				ch_assignedOrder <- types.AssignedOrder{
					OrderType: btn_event,
					ID:        localID,
				}

			} else {
				//run cost function on all elevators, assign to lowest return value
				AssignedElevID := localID
				elev_copy := utilities.DeepCopyElevatorStruct(elevatorMap[AssignedElevID])
				elev_copy.Requests[btn_event.Floor][btn_event.Button] = true
				min_time := costfn.TimeToIdle(elev_copy)

				for _, id := range peerUpdate.Peers {
					elev_copy = utilities.DeepCopyElevatorStruct(elevatorMap[id])
					elev_copy.Requests[btn_event.Floor][btn_event.Button] = true
					if costfn.TimeToIdle(elev_copy) < min_time {
						AssignedElevID = id
						min_time = costfn.TimeToIdle(elev_copy)
					}
				}
				fmt.Println("Assigned elevator: ", AssignedElevID)

				ch_assignedOrder <- types.AssignedOrder{OrderType: btn_event, ID: AssignedElevID}
			}
		default: 
		
			if len(peerUpdate.Lost) != 0 {
			//reassign orders to myself
				for _, elev := range peerUpdate.Lost {
					for i := 0; i < config.NumFloors; i++ {
						for j:=0; j<config.NumButtons-1; j++ {
							if elevatorMap[elev].Requests[i][j] {
								ch_assignedOrder <- types.AssignedOrder{OrderType: elevio.ButtonEvent{Floor: i, Button: elevio.ButtonType(j)}, ID: localID}
							}
						}
					}
				}
			}
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
