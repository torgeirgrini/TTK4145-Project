package assigner

import (
	"Project/assignment/costfn"
	"Project/localElevator/elevio"
	"Project/network/peers"
	"Project/types"
	"Project/utilities"
	"fmt"
)

func Assignment(
	localID string,
	ch_informationToAssigner <-chan types.AssignerMessage,
	ch_hwButtonPress <-chan elevio.ButtonEvent,
	ch_assignedOrder chan<- types.AssignedOrder,
) {

	var elevatorMap map[string]types.Elevator
	var peerUpdate peers.PeerUpdate
	var assignerMsg types.AssignerMessage

	assignerMsg = <-ch_informationToAssigner
	elevatorMap = utilities.DeepCopyElevatorMap(assignerMsg.ElevatorMap)
	peerUpdate = utilities.DeepCopyPeerStatus(assignerMsg.PeerStatus)

	var AssignedElevID string
	for {
		select {
		case assignerMsg = <-ch_informationToAssigner:
			elevatorMap = utilities.DeepCopyElevatorMap(assignerMsg.ElevatorMap)
			peerUpdate = utilities.DeepCopyPeerStatus(assignerMsg.PeerStatus)
		case btn_event := <-ch_hwButtonPress:
			AssignedElevID = localID

			if btn_event.Button != elevio.BT_Cab {
				for _, id := range peerUpdate.Peers {
					if _, ok := elevatorMap[id]; ok && elevatorMap[id].Available {
						AssignedElevID = id
					}
				}
				e_copy := utilities.DeepCopyElevatorStruct(elevatorMap[AssignedElevID])
				e_copy.Requests[btn_event.Floor][btn_event.Button] = true
				min_time := costfn.TimeToIdle(e_copy)
				for _, id := range peerUpdate.Peers {
					if _, ok := elevatorMap[id]; ok && elevatorMap[id].Available {
						e_copy = utilities.DeepCopyElevatorStruct(elevatorMap[id])
						e_copy.Requests[btn_event.Floor][btn_event.Button] = true
						calc_cost := costfn.TimeToIdle(e_copy)
						if calc_cost < min_time {
							AssignedElevID = id
							min_time = calc_cost
						}
					}
				}
			}
			fmt.Println(AssignedElevID)
			ch_assignedOrder <- types.AssignedOrder{OrderType: btn_event, ID: AssignedElevID}
		}
	}
}
