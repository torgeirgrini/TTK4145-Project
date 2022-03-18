package assigner

import (
	"Project/assignment/costfn"
	"Project/config"
	"Project/distribution"
	"Project/localElevator/elevio"
	"Project/network/peers"
	"Project/types"
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
	ch_allElevators <-chan map[string]types.Elevator,
	ch_newLocalOrder <-chan elevio.ButtonEvent,
	ch_newOrderAssigned chan<- types.MsgToDistributor,
) {

	elevators := make(map[string]types.Elevator)
	fmt.Printf("Assign elevators: %p, %#+v\n", elevators, elevators)
	//btn_event := elevio.ButtonEvent{}
	ch_peerUpdate := make(chan peers.PeerUpdate)
	go peers.Receiver(config.PortPeers, ch_peerUpdate)
	var peerAvailability peers.PeerUpdate

	for {
		select {
		case elevat := <-ch_allElevators:
			//fmt.Printf("assign | new elevators: %+v\n", elevat)
			elevators = distribution.DeepCopy(elevat)
			//fmt.Printf("Assign elevators (after copy): %p, %#+v\n", elevators, elevators)

			//Videresend til assigner sånn at den kan regne ut
		case btn_event := <-ch_newLocalOrder:
			//Here we need to calcilate the cost functin
			if btn_event.Button == elevio.BT_Cab {
				ch_newOrderAssigned <- types.MsgToDistributor{
					OrderType: btn_event,
					ID:        localID,
				}

			} else {

				AssignedElevID := localID

				fmt.Println("Peerlist: ", peerAvailability.Peers)
				fmt.Println("Buttonevent")
				fmt.Printf("assign | elevators | %+#v\n", elevators)
				elev_copy := types.Dup(elevators[AssignedElevID])
				fmt.Printf("assign | elev copy | %+#v\n", elev_copy)
				elev_copy.Requests[btn_event.Floor][btn_event.Button] = true
				min_time := costfn.TimeToIdle(elev_copy)

				for id, _ := range elevators {
					fmt.Println("ID:", id)
					elev_copy = types.Dup(elevators[id])
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

				send_msg := types.MsgToDistributor{OrderType: btn_event, ID: AssignedElevID}
				ch_newOrderAssigned <- send_msg
			}

		case peerAvailability = <-ch_peerUpdate:

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
