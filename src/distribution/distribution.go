package distribution

import (
	"Project/config"
	"Project/localElevator/elevio"
	"Project/network/bcast"
	"Project/network/peers"
	"Project/types"
	"fmt"
	"reflect"
	"time"
)

//Denne har ansvar for ordrebestillinger, og håndtere ordretap osv
//Sende ourOrders til localElevator
//Send allOrders for å kunne sette lys overalt for eksempel

/*
func Distribution(id string,
	ch_newLocalOrder <-chan elevio.ButtonType,
	ch_NewElevatorStateMap chan<- map[string]types.Elevator,
	ch_newLocalElevator <-chan types.Elevator) {

	elevatorSystemMap := make(map[string]types.Elevator)
	for {
		select {

		case newOrder := <-ch_newLocalOrder:
			elevatorSystemMap = updateMapWithOrder(id, elevatorSystemMap, newOrder)
			ch_unassignedOrder <- elevatorSystemMap     //Til assigner
			ch_NewElevatorStateMap <- elevatorSystemMap //Til network

		case updatedElevatorSystemMap := <-ch_updatedElevatorSystemMap:
			//sebd den ut på nettet!!
			//send vår elevstruct til localElevator fsm

		case newLocalElevator := <-ch_newLocalElevator:
			elevatorSystemMap = updateMapWithLocalElevator(id, elevatorSystemMap, newLocalElevator)
			// case <-ch_doorTimer:

			// case obstruction = <-ch_Obstruction:

		}
	}

}
*/

func DeepCopy(elevators map[string]types.Elevator) map[string]types.Elevator {
	copied := make(map[string]types.Elevator)
	for i, e := range elevators {
		copied[i] = types.Dup(e)
	}
	return copied
}

func Distribution(
	localID string,
	localElevator <-chan types.Elevator,
	allElevators chan<- map[string]types.Elevator,
	ch_newOrderAssigned <-chan types.MsgToDistributor,
	ch_orderAssignedToLocal chan<- elevio.ButtonEvent,
) {

	ch_peerUpdate := make(chan peers.PeerUpdate)
	go peers.Receiver(config.PortPeers, ch_peerUpdate)

	tx := make(chan types.ElevatorStateMessage)
	rx := make(chan types.ElevatorStateMessage)
	ch_newOrderAssignedToAll := make(chan types.MsgToDistributor)
	ch_readNewOrderFromNetwork := make(chan types.MsgToDistributor)
	go bcast.Transmitter(config.PortBroadcast, tx, ch_newOrderAssignedToAll)
	go bcast.Receiver(config.PortBroadcast, rx, ch_readNewOrderFromNetwork)

	elevators := make(map[string]types.Elevator)
	fmt.Printf("Distr elevator: %p, %#+v\n", elevators, elevators)

	tick := time.NewTicker(config.TransmitInterval * time.Millisecond)

	Hallcalls := make([][]types.HallCall, config.NumFloors)
	for i := range Hallcalls {
		Hallcalls[i] = make([]types.HallCall, config.NumButtons-1)
	}

	for {
		//fmt.Println("FORLOOP elevatosmap: ", elevators[localID].Requests)
		select {
		case newOrder := <-ch_newOrderAssigned:
			fmt.Printf("distribution | new order from assigner: %#+v\n", newOrder)
			elevators[newOrder.ID].Requests[newOrder.OrderType.Floor][newOrder.OrderType.Button] = true
			fmt.Println("Message rcvd")
			fmt.Println("elevators:", elevators[localID].Requests)

			if newOrder.ID == localID {
				ch_orderAssignedToLocal <- newOrder.OrderType
			}
			ch_newOrderAssignedToAll <- newOrder

		case e := <-localElevator:
			//fmt.Printf("distribution | new local elevator: %#+v\n", e)
			if !reflect.DeepEqual(elevators[localID], e) {
				elevators[localID] = e
				fmt.Println("DeepEqual entered from localelevator")
				elevators[localID] = e
				allElevators <- DeepCopy(elevators)
			}

		case <-tick.C:
			//fmt.Println("Her3")
			tx <- types.ElevatorStateMessage{ID: localID, HallCalls: Hallcalls, ElevState: types.Dup(elevators[localID])}
		case remote := <-rx:
			//fmt.Printf("distribution | states from remote: %#+v\n", remote)

			if !reflect.DeepEqual(elevators[remote.ID], remote.ElevState) {
				elevators[remote.ID] = remote.ElevState
				fmt.Println("DeepEqual entered from network")
				allElevators <- DeepCopy(elevators)
				setHallCalllights(elevators)
			}
		// case peerUpdate := <-ch_peerUpdate:
		// 	fmt.Printf("Peer update:\n")
		// 	fmt.Printf("  Peers:    %q\n", peerUpdate.Peers)
		// 	fmt.Printf("  New:      %q\n", peerUpdate.New)
		// 	fmt.Printf("  Lost:     %q\n", peerUpdate.Lost)
		// 	//Må si ifra om at noen har kommet på/fallt av nettet
		// 	//Kan for eksmpel gjøres ved å sette available bit i elevators(ESM'en)

		case newOrder := <-ch_readNewOrderFromNetwork:
			fmt.Printf("distribution | new order from net: %#+v\n", newOrder)
			if newOrder.ID == localID {
				ch_orderAssignedToLocal <- newOrder.OrderType
			}
		}
	}
}

func setHallCalllights(allElevators map[string]types.Elevator) {
	hallcalls := HallRequestsFromESM(allElevators)
	for i := 0; i < config.NumFloors; i++ {
		for j := 0; j < config.NumButtons-1; j++ {
			elevio.SetButtonLamp(elevio.ButtonType(j), i, hallcalls[i][j])
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
