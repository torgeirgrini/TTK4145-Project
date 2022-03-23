package distribution

import (
	"Project/config"
	"Project/localElevator/elevio"
	"Project/network/bcast"
	"Project/network/peers"
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
	ch_informationToAssigner chan<- types.AssignerMessage,
	ch_assignedOrder <-chan types.AssignedOrder,
	ch_newLocalOrder chan<- elevio.ButtonEvent,
) {

	ch_txNetworkMsg := make(chan types.NetworkMessage)
	ch_rxNetworkMsg := make(chan types.NetworkMessage)
	ch_newAssignedOrderToNetwork := make(chan types.AssignedOrder)
	ch_readNewOrderFromNetwork := make(chan types.AssignedOrder)
	go bcast.Transmitter(config.PortBroadcast, ch_txNetworkMsg, ch_newAssignedOrderToNetwork)
	go bcast.Receiver(config.PortBroadcast, ch_rxNetworkMsg, ch_readNewOrderFromNetwork)

	ch_peerUpdate := make(chan peers.PeerUpdate)
	go peers.Receiver(config.PortPeers, ch_peerUpdate)
	var peerAvailability peers.PeerUpdate

	elevators := make(map[string]types.Elevator)
	//fmt.Printf("Distr elevator: %p, %#+v\n", elevators, elevators)

	tick := time.NewTicker(config.TransmitInterval_ms * time.Millisecond)

	Hallcalls := make([][]types.HallCall, config.NumFloors)
	for i := range Hallcalls {
		Hallcalls[i] = make([]types.HallCall, config.NumButtons-1)
		for j := range Hallcalls[i] {
			Hallcalls[i][j] = types.HallCall{ExecutorID: "", AssignerID: "", OrderState: types.OS_NONE, AckList: make([]string, 0)}
		}
	}

	//Wait til elevator initialized
	elevators[localID] = <-ch_localElevatorState
	ch_informationToAssigner <- types.AssignerMessage{
		PeerList:    utilities.DeepCopyStringSlice(peerAvailability.Peers, len(peerAvailability.Peers)),
		ElevatorMap: utilities.DeepCopyElevatorMap(elevators),
	}
	for {
		select {
		case newAssignedOrder := <-ch_assignedOrder:

			if newAssignedOrder.ID == localID {
				ch_newLocalOrder <- newAssignedOrder.OrderType
			} else {
				Hallcalls[newAssignedOrder.OrderType.Floor][newAssignedOrder.OrderType.Button].AssignerID = localID
				Hallcalls[newAssignedOrder.OrderType.Floor][newAssignedOrder.OrderType.Button].ExecutorID = newAssignedOrder.ID
				Hallcalls[newAssignedOrder.OrderType.Floor][newAssignedOrder.OrderType.Button].OrderState = types.OS_UNCONFIRMED
			}
			ch_txNetworkMsg <- types.NetworkMessage{ID: localID,
				HallCalls: utilities.DeepCopyHallCalls(Hallcalls),
				ElevState: utilities.DeepCopyElevatorStruct(elevators[localID]),
			}

		case e := <-ch_localElevatorState:
			if !reflect.DeepEqual(elevators[localID], e) {
				elevators[localID] = utilities.DeepCopyElevatorStruct(e)
				ch_informationToAssigner <- types.AssignerMessage{
					PeerList:    utilities.DeepCopyStringSlice(peerAvailability.Peers, len(peerAvailability.Peers)),
					ElevatorMap: utilities.DeepCopyElevatorMap(elevators),
				}
				setHallCalllights(elevators)
			}

		case <-tick.C:
			ch_txNetworkMsg <- types.NetworkMessage{
				ID:        localID,
				HallCalls: utilities.DeepCopyHallCalls(Hallcalls),
				ElevState: utilities.DeepCopyElevatorStruct(elevators[localID]),
			}
		case remote := <-ch_rxNetworkMsg:
			if remote.ID != localID {
				for floor := 0; floor < config.NumFloors; floor++ {
					for btn, remote_hc := range remote.HallCalls[floor] {
						if remote_hc.ExecutorID == localID && remote_hc.OrderState == types.OS_UNCONFIRMED {
							Hallcalls[floor][btn].OrderState = types.OS_CONFIRMED
							ch_newLocalOrder <- elevio.ButtonEvent{
								Floor:  floor,
								Button: elevio.ButtonType(btn),
							}
						}
					}
				}
			}
			if !reflect.DeepEqual(elevators[remote.ID], remote.ElevState) {
				elevators[remote.ID] = remote.ElevState
				ch_informationToAssigner <- types.AssignerMessage{
					PeerList:    utilities.DeepCopyStringSlice(peerAvailability.Peers, len(peerAvailability.Peers)),
					ElevatorMap: utilities.DeepCopyElevatorMap(elevators),
				}
				setHallCalllights(elevators)
			}
		case peerAvailability = <-ch_peerUpdate:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", peerAvailability.Peers)
			fmt.Printf("  New:      %q\n", peerAvailability.New)
			fmt.Printf("  Lost:     %q\n", peerAvailability.Lost)
			//fmt.Println("ElevatorMap: ", elevators)

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

func removeDuplicates(s []string) []string {
	inResult := make(map[string]bool)
	var result []string
	for _, str := range s {
		if _, ok := inResult[str]; !ok {
			inResult[str] = true
			result = append(result, str)
		}
	}
	return result
}

func sameStringSlice(x, y []string) bool {
	if len(x) != len(y) {
		return false
	}
	diff := make(map[string]int, len(x))
	for _, _x := range x {
		diff[_x]++
	}
	for _, _y := range y {
		if _, ok := diff[_y]; !ok {
			return false
		}
		diff[_y] -= 1
		if diff[_y] == 0 {
			delete(diff, _y)
		}
	}
	if len(diff) == 0 {
		return true
	}
	return false
}