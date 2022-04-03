package assigner

import (
	"Project/assignment/costfn"
	"Project/localElevator/elevio"
	"Project/network/peers"
	"Project/types"
	"Project/utilities"
)

func Assignment(
	localID 				   string,
	ch_peerStatus     <-chan   peers.PeerUpdate,
	ch_elevMap 		  <-chan   map[string]types.Elevator,
	ch_hwButtonPress  <-chan   elevio.ButtonEvent,
	ch_assignedOrder 	chan<- types.AssignedOrder,
) {

	var elevatorMap map[string]types.Elevator
	var peerUpdate peers.PeerUpdate

	elevatorMap = <-ch_elevMap
	peerUpdate = <-ch_peerStatus

	var AssignedElevID string
	for {
		select {
		case elevatorMap = <-ch_elevMap:
		case peerUpdate = <-ch_peerStatus:
		case buttonEvent := <-ch_hwButtonPress:
			AssignedElevID = localID
			if buttonEvent.Button != elevio.BT_Cab {
				for _, id := range peerUpdate.Peers {
					if _, ok := elevatorMap[id]; ok {
						AssignedElevID = id
					}
				}
				elevCopy := utilities.DeepCopyElevatorStruct(elevatorMap[AssignedElevID])
				elevCopy.Requests[buttonEvent.Floor][buttonEvent.Button] = true
				minCost := costfn.TimeToIdle(elevCopy)
				for _, id := range peerUpdate.Peers {
					if _, ok := elevatorMap[id]; ok {
						elevCopy = utilities.DeepCopyElevatorStruct(elevatorMap[id])
						elevCopy.Requests[buttonEvent.Floor][buttonEvent.Button] = true
						cost := costfn.TimeToIdle(elevCopy)
						if cost < minCost {
							AssignedElevID = id
							minCost = cost
						}
					}
				}
			}
			ch_assignedOrder <- types.AssignedOrder{OrderType: buttonEvent, ID: AssignedElevID}
		}
	}
}
