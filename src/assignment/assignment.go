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
	ch_elevatorMap <-chan map[string]types.Elevator,
	ch_hwButtonPress <-chan elevio.ButtonEvent,
	ch_assignedOrder chan<- types.MsgToDistributor,
) {

	elevators := make(map[string]types.Elevator)
	fmt.Printf("Assign elevators: %p, %#+v\n", elevators, elevators)
	//btn_event := elevio.ButtonEvent{}
	ch_peerUpdate := make(chan peers.PeerUpdate)
	go peers.Receiver(config.PortPeers, ch_peerUpdate)
	var peerAvailability peers.PeerUpdate

	for {
		select {
		case elevators = <-ch_elevatorMap:
			//fmt.Printf("assign | new elevators: %+v\n", elevators)
			//fmt.Printf("Assign elevators (after copy): %p, %#+v\n", elevators, elevators)

			//Videresend til assigner sånn at den kan regne ut
		case btn_event := <-ch_hwButtonPress:
			//Here we need to calcilate the cost functin
			if btn_event.Button == elevio.BT_Cab {
				ch_assignedOrder <- types.MsgToDistributor{
					OrderType: btn_event,
					ID:        localID,
				}

			} else {

				AssignedElevID := localID

				fmt.Println("Peerlist: ", peerAvailability.Peers)
				fmt.Println("Buttonevent")
				fmt.Printf("assign | elevators | %+#v\n", elevators)
				elev_copy := utilities.DeepCopyElevatorStruct(elevators[AssignedElevID])
				fmt.Printf("assign | elev copy | %+#v\n", elev_copy)
				elev_copy.Requests[btn_event.Floor][btn_event.Button] = true
				min_time := costfn.TimeToIdle(elev_copy)

				for _,id := range peerAvailability.Peers {
					fmt.Println("ID:", id)
					elev_copy = utilities.DeepCopyElevatorStruct(elevators[id])
					fmt.Println("elevcopyreq: ", elev_copy.Requests)
					fmt.Println("btneventfloor: ", btn_event.Floor)
					fmt.Println("btneventbtn: ", btn_event.Button)
					elev_copy.Requests[btn_event.Floor][btn_event.Button] = true
					if costfn.TimeToIdle(elev_copy) < min_time {
						AssignedElevID = id
						min_time = costfn.TimeToIdle(elev_copy)
					}
				}
				fmt.Println("Assigned elevator: ", AssignedElevID)

				ch_assignedOrder <- types.MsgToDistributor{OrderType: btn_event, ID: AssignedElevID}
			}

		case peerAvailability = <-ch_peerUpdate:
			// fmt.Printf("Peer update:\n")
			// fmt.Printf("  Peers:    %q\n", peerAvailability.Peers)
			// fmt.Printf("  New:      %q\n", peerAvailability.New)
			// fmt.Printf("  Lost:     %q\n", peerAvailability.Lost)
			// fmt.Println("ElevatorMap: ", elevators)

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
