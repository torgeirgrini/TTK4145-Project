package distribution

import (
	"Project/config"
	"Project/localElevator/elevio"
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
		copied[i] = e
	}
	return copied
}

func Distribution(
	localID string,
	localElevator <-chan types.Elevator,
	allElevators chan<- map[string]types.Elevator,
	tx chan<- types.ElevatorStateMessage,
	rx <-chan types.ElevatorStateMessage,
	ch_peerUpdate <-chan peers.PeerUpdate,
	ch_newOrderAssigned <-chan types.MsgToDistributor,
	ch_orderAssignedToLocal chan<- elevio.ButtonEvent,
	ch_newOrderAssignedToAll chan<- types.MsgToDistributor,
	ch_readNewOrderFromNetwork <-chan types.MsgToDistributor) {

	elevators := make(map[string]types.Elevator)

	tick := time.NewTicker(config.TransmitInterval * time.Millisecond)

	Hallcalls := make([][]types.HallCall, config.NumFloors)
	for i := range Hallcalls {
		Hallcalls[i] = make([]types.HallCall, config.NumButtons-1)
	}
	for {
		select {
		case e := <-localElevator:
			if !reflect.DeepEqual(elevators[localID], e) {
				elevators[localID] = e
				allElevators <- DeepCopy(elevators)
			}
		case <-tick.C:
			tx <- types.ElevatorStateMessage{localID, elevators[localID], Hallcalls, elevators}
		case remote := <-rx:
			if !reflect.DeepEqual(elevators[remote.ID], remote.LocalElevator) {
				elevators[remote.ID] = remote.LocalElevator
				allElevators <- DeepCopy(elevators)
				setHallCalllights(elevators)
			}
		case peerUpdate := <-ch_peerUpdate:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", peerUpdate.Peers)
			fmt.Printf("  New:      %q\n", peerUpdate.New)
			fmt.Printf("  Lost:     %q\n", peerUpdate.Lost)
			//Må si ifra om at noen har kommet på/fallt av nettet
			//Kan for eksmpel gjøres ved å sette available bit i elevators(ESM'en)
		case updatedElevators := <-ch_newOrderAssigned:
			if !reflect.DeepEqual(elevators[localID], updatedElevators.Elevators[localID]) {
				elevators[localID] = updatedElevators.Elevators[localID]
				ch_orderAssignedToLocal <- updatedElevators.OrderType
			}
			ch_newOrderAssignedToAll <- updatedElevators
		case updatedOrders := <-ch_readNewOrderFromNetwork:
			if !reflect.DeepEqual(elevators[localID], updatedOrders.Elevators[localID]) {
				elevators[localID] = updatedOrders.Elevators[localID]
				ch_orderAssignedToLocal <- updatedOrders.OrderType
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
