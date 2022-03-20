package distribution

import (
	"Project/config"
	"Project/localElevator/elevio"
	"Project/network/bcast"
	"Project/types"
	"Project/utilities"
	"fmt"
	"reflect"
	"time"
)

//Denne har ansvar for ordrebestillinger, og håndtere ordretap osv
//Sende ourOrders til localElevator
//Send allOrders for å kunne sette lys overalt for eksempel

func Distribution(
	localID string,
	ch_localElevatorState <-chan types.Elevator,
	ch_elevatorMap chan<- map[string]types.Elevator,
	ch_assignedOrder <-chan types.MsgToDistributor,
	ch_newLocalOrder chan<- elevio.ButtonEvent,
) {

	// ch_peerUpdate := make(chan peers.PeerUpdate)
	// go peers.Receiver(config.PortPeers, ch_peerUpdate)

	ch_txElevatorMap := make(chan types.ElevatorStateMessage)
	ch_rxElevatorMap := make(chan types.ElevatorStateMessage)
	ch_newOrderAssignedToAll := make(chan types.MsgToDistributor)
	ch_readNewOrderFromNetwork := make(chan types.MsgToDistributor)
	go bcast.Transmitter(config.PortBroadcast, ch_txElevatorMap, ch_newOrderAssignedToAll)
	go bcast.Receiver(config.PortBroadcast, ch_rxElevatorMap, ch_readNewOrderFromNetwork)

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
		case newOrder := <-ch_assignedOrder:
			fmt.Printf("distribution | new order from assigner: %#+v\n", newOrder)
			elevators[newOrder.ID].Requests[newOrder.OrderType.Floor][newOrder.OrderType.Button] = true
			fmt.Println("elevators:", elevators[localID].Requests)

			if newOrder.ID == localID {
				ch_newLocalOrder <- newOrder.OrderType
			}
			fmt.Println("Order to network: ", newOrder)
			ch_newOrderAssignedToAll <- newOrder

		case e := <-ch_localElevatorState:
			//fmt.Printf("distribution | new local elevator: %#+v\n", e)
			if !reflect.DeepEqual(elevators[localID], e) {
				elevators[localID] = e
				elevators[localID] = e
				ch_elevatorMap <- utilities.DeepCopyElevatorMap(elevators)
			}

		case <-tick.C:
			//fmt.Println("Her3")
			ch_txElevatorMap <- types.ElevatorStateMessage{
				ID: localID, 
				HallCalls: Hallcalls, 
				ElevState: utilities.DeepCopyElevatorStruct(elevators[localID]),
			}
		case remote := <-ch_rxElevatorMap:
			//fmt.Printf("distribution | states from remote: %#+v\n", remote)
			if !reflect.DeepEqual(elevators[remote.ID], remote.ElevState) {
				elevators[remote.ID] = remote.ElevState
				ch_elevatorMap <- utilities.DeepCopyElevatorMap(elevators)
				setHallCalllights(elevators)
			}
		case newOrder := <-ch_readNewOrderFromNetwork:
			fmt.Printf("distribution | new order from net: %#+v\n", newOrder)
			fmt.Println("Local ID, RemoteID: ", localID, " ", newOrder.ID)
			if newOrder.ID == localID {
				ch_newLocalOrder <- newOrder.OrderType
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
