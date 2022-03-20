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
	ch_assignedOrder <-chan types.AssignedOrder,
	ch_newLocalOrder chan<- elevio.ButtonEvent,
) {

	ch_txNetworkMsg := make(chan types.NetworkMessage)
	ch_rxNetworkMsg := make(chan types.NetworkMessage)
	ch_newAssignedOrderToNetwork := make(chan types.AssignedOrder)
	ch_readNewOrderFromNetwork := make(chan types.AssignedOrder)
	go bcast.Transmitter(config.PortBroadcast, ch_txNetworkMsg, ch_newAssignedOrderToNetwork)
	go bcast.Receiver(config.PortBroadcast, ch_rxNetworkMsg, ch_readNewOrderFromNetwork)

	elevators := make(map[string]types.Elevator)
	//fmt.Printf("Distr elevator: %p, %#+v\n", elevators, elevators)

	tick := time.NewTicker(config.TransmitInterval * time.Millisecond)

	Hallcalls := make([][]types.HallCall, config.NumFloors)
	for i := range Hallcalls {
		Hallcalls[i] = make([]types.HallCall, config.NumButtons-1)
	}

	for {
		select {
		case newAssignedOrder := <-ch_assignedOrder:
			fmt.Printf("distribution | new order from assigner: %#+v\n", newAssignedOrder)
			elevators[newAssignedOrder.ID].Requests[newAssignedOrder.OrderType.Floor][newAssignedOrder.OrderType.Button] = true
			fmt.Println("elevators:", elevators[localID].Requests)

			if newAssignedOrder.ID == localID {
				ch_newLocalOrder <- newAssignedOrder.OrderType
			}
			fmt.Println("Order to network: ", newAssignedOrder)
			ch_newAssignedOrderToNetwork <- newAssignedOrder

			//Lag en ny HallCall
			if newAssignedOrder.OrderType.Button != elevio.BT_Cab {
				newHallCall := types.HallCall{
				ExecutorID: newAssignedOrder.ID,
				AssignerID: localID,
				OrderState: types.OS_UNCONFIRMED,
				AckList: make([]string, 0),
				}
			Hallcalls[newAssignedOrder.OrderType.Floor][newAssignedOrder.OrderType.Button] = newHallCall

			//Må endres
			/*
			ch_txNetworkMsg <- types.NetworkMessage{
				ID: localID, 
				HallCalls: utilities.DeepCopyHallCalls(Hallcalls), 
				ElevState: utilities.DeepCopyElevatorStruct(elevators[localID]),
			}*/
			}


		case e := <-ch_localElevatorState:
			if !reflect.DeepEqual(elevators[localID], e) {
				elevators[localID] = e
				elevators[localID] = e
				ch_elevatorMap <- utilities.DeepCopyElevatorMap(elevators)
				/*if ordre utført {
					Hallcalls[f][b] = types.OS_COMPLETED
				}*/
			}

		case <-tick.C:
			//fmt.Println("Hallcalls tx: ", Hallcalls)
			
			ch_txNetworkMsg <- types.NetworkMessage{
				ID: localID, 
				HallCalls: utilities.DeepCopyHallCalls(Hallcalls), 
				ElevState: utilities.DeepCopyElevatorStruct(elevators[localID]),
			}
		case remote := <-ch_rxNetworkMsg:
			//fmt.Printf("distribution | states from remote: %#+v\n", remote)
			//switch case med order state. bare telle oppover. cyclic counter. kan ikke sjekke om de er ulike, vi må sjekke om den på remote har kommet lengre i sykelen, isåfall kan vi oppdatere
			if remote.ID != localID{
				if !reflect.DeepEqual(elevators[remote.ID], remote.ElevState) {
				elevators[remote.ID] = remote.ElevState
				ch_elevatorMap <- utilities.DeepCopyElevatorMap(elevators)
				setHallCalllights(elevators)
			}

			//btn := elevio.BT_HallDown
			//floor := 2
			fmt.Println("HallCalls local save: ", Hallcalls)
			if !reflect.DeepEqual(remote.HallCalls, Hallcalls) {
				Hallcalls = utilities.DeepCopyHallCalls(remote.HallCalls)
				//fmt.Println("HallCalls from network, execID: ", remote.HallCalls[floor][btn].ExecutorID)
				//fmt.Println("HallCalls from network, assID: ", remote.HallCalls[floor][btn].AssignerID)
				
				fmt.Println("HallCalls from network: ", remote.HallCalls)
			}
			}
			



		case newOrder := <-ch_readNewOrderFromNetwork:
			fmt.Printf("distribution | new order from net: %#+v\n", newOrder)
			fmt.Println("Local ID, RemoteID: ", localID, " ", newOrder.ID)
			if newOrder.ID == localID && newOrder.OrderType.Button != elevio.BT_Cab {
				ch_newLocalOrder <- newOrder.OrderType
				Hallcalls[newOrder.OrderType.Floor][newOrder.OrderType.Button].OrderState = types.OS_CONFIRMED
			}
			//Hallcalls[newOrder.OrderType.Floor][newOrder.OrderType.Button].AckList = (Hallcalls[newOrder.OrderType.Floor][newOrder.OrderType.Button].AckList, localID)
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
