package assigner

import (
	"Project/assignment/costfn"
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
	var assignerMsg types.AssignerMessage
	assignerMsg = <-ch_informationToAssigner

	elevatorMap = utilities.DeepCopyElevatorMap(assignerMsg.ElevatorMap)
	peerUpdate = utilities.DeepCopyPeerStatus(assignerMsg.PeerStatus) //deepcopy

	for {

		select {
		//update information from distribtion
		case assignerMsg = <-ch_informationToAssigner:

			elevatorMap = utilities.DeepCopyElevatorMap(assignerMsg.ElevatorMap)
			peerUpdate = utilities.DeepCopyPeerStatus(assignerMsg.PeerStatus) //deepcopy

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
					if _, ok := elevatorMap[id]; ok {
						elev_copy = utilities.DeepCopyElevatorStruct(elevatorMap[id])
						elev_copy.Requests[btn_event.Floor][btn_event.Button] = true
						if costfn.TimeToIdle(elev_copy) < min_time {
							AssignedElevID = id
							min_time = costfn.TimeToIdle(elev_copy)
						}
					}
				}
				fmt.Println("Assigned elevator: ", AssignedElevID)

				ch_assignedOrder <- types.AssignedOrder{OrderType: btn_event, ID: AssignedElevID}
			}

		}
	}
}
