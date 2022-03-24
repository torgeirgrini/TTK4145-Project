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
	ch_localOrderCompleted <-chan elevio.ButtonEvent,
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
	prevLocalOrders := make([][]bool, config.NumFloors)

	//fmt.Printf("Distr elevator: %p, %#+v\n", elevators, elevators)

	tick := time.NewTicker(config.TransmitInterval_ms * time.Millisecond)

	Hallcalls := make([][]types.HallCall, config.NumFloors)
	for i := range Hallcalls {
		Hallcalls[i] = make([]types.HallCall, config.NumButtons-1)
		for j := range Hallcalls[i] {
			Hallcalls[i][j] = types.HallCall{ExecutorID: "", AssignerID: "", OrderState: types.OS_COMPLETED, AckList: make([]string, 0)}
		}
	}

	for i := range prevLocalOrders {
		prevLocalOrders[i] = make([]bool, config.NumButtons-1)
		for j := range prevLocalOrders[i] {
			prevLocalOrders[i][j] = false
		}
	}

	//Wait til elevator initialized
	peerAvailability = <-ch_peerUpdate
	elevators[localID] = <-ch_localElevatorState
	ch_informationToAssigner <- types.AssignerMessage{
		PeerList:    utilities.DeepCopyStringSlice(peerAvailability.Peers),
		ElevatorMap: utilities.DeepCopyElevatorMap(elevators),
	}
	for {
		select {
		case newAssignedOrder := <-ch_assignedOrder:

			if newAssignedOrder.ID == localID {
				ch_newLocalOrder <- newAssignedOrder.OrderType

			}
			if newAssignedOrder.OrderType.Button != elevio.BT_Cab &&
				(Hallcalls[newAssignedOrder.OrderType.Floor][newAssignedOrder.OrderType.Button].OrderState == types.OS_COMPLETED ||
					Hallcalls[newAssignedOrder.OrderType.Floor][newAssignedOrder.OrderType.Button].OrderState == types.OS_UNKNOWN) {
				Hallcalls[newAssignedOrder.OrderType.Floor][newAssignedOrder.OrderType.Button].AssignerID = localID
				Hallcalls[newAssignedOrder.OrderType.Floor][newAssignedOrder.OrderType.Button].ExecutorID = newAssignedOrder.ID
				Hallcalls[newAssignedOrder.OrderType.Floor][newAssignedOrder.OrderType.Button].OrderState = types.OS_UNCONFIRMED
				Hallcalls[newAssignedOrder.OrderType.Floor][newAssignedOrder.OrderType.Button].AckList =
					append(Hallcalls[newAssignedOrder.OrderType.Floor][newAssignedOrder.OrderType.Button].AckList, localID)
				Hallcalls[newAssignedOrder.OrderType.Floor][newAssignedOrder.OrderType.Button].AckList =
					removeDuplicates(Hallcalls[newAssignedOrder.OrderType.Floor][newAssignedOrder.OrderType.Button].AckList)
			}
			ch_txNetworkMsg <- types.NetworkMessage{ID: localID,
				HallCalls: utilities.DeepCopyHallCalls(Hallcalls),
				ElevState: utilities.DeepCopyElevatorStruct(elevators[localID]),
			}

		case e := <-ch_localElevatorState: //change this to compl orders and move this channel to assigner
			if !reflect.DeepEqual(elevators[localID], e) {

				elevators[localID] = utilities.DeepCopyElevatorStruct(e)
				ch_informationToAssigner <- types.AssignerMessage{
					PeerList:    utilities.DeepCopyStringSlice(peerAvailability.Peers),
					ElevatorMap: utilities.DeepCopyElevatorMap(elevators),
				}
			}
		case localCompletedOrder := <-ch_localOrderCompleted:
			//if Hallcalls[localCompletedOrder.Floor][localCompletedOrder.Button].OrderState == types.OS_CONFIRMED {
			Hallcalls[localCompletedOrder.Floor][localCompletedOrder.Button].OrderState = types.OS_COMPLETED
			//fmt.Println(Hallcalls[localCompletedOrder.Floor][localCompletedOrder.Button].AckList)
			Hallcalls[localCompletedOrder.Floor][localCompletedOrder.Button].AckList = make([]string, 0)
			Hallcalls[localCompletedOrder.Floor][localCompletedOrder.Button].ExecutorID = ""
			Hallcalls[localCompletedOrder.Floor][localCompletedOrder.Button].AssignerID = ""
			//fmt.Println("I removed3")
			//}
			//fmt.Println("Completed orders: ", localCompletedOrder, Hallcalls[localCompletedOrder.Floor][localCompletedOrder.Button])

		case <-tick.C:
			// confirm orders that have a full ack list
			for floor := 0; floor < config.NumFloors; floor++ {
				for btn, hc := range Hallcalls[floor] {
					if hc.OrderState == types.OS_UNCONFIRMED && sameStringSlice(peerAvailability.Peers, hc.AckList) {
						Hallcalls[floor][btn].OrderState = types.OS_CONFIRMED
					}
				}
			}
			ch_txNetworkMsg <- types.NetworkMessage{
				ID:        localID,
				HallCalls: utilities.DeepCopyHallCalls(Hallcalls),
				ElevState: utilities.DeepCopyElevatorStruct(elevators[localID]),
			}
			fmt.Println("hc1: ",Hallcalls)
			ourOrders := generateOurHallcalls(Hallcalls, localID)
			allOrders := generateAllHallcalls(Hallcalls)
			for floor := 0; floor < config.NumFloors; floor++ {
				for btn, hc := range Hallcalls[floor] {
					if prevLocalOrders[floor][btn] != ourOrders[floor][btn] && hc.ExecutorID == localID {
						ch_newLocalOrder <- elevio.ButtonEvent{Floor: floor, Button: elevio.ButtonType(btn)}
					}
					prevLocalOrders[floor][btn] = ourOrders[floor][btn]
					elevio.SetButtonLamp(elevio.ButtonType(btn), floor, allOrders[floor][btn])
				}
			}

			// if alone on network, change completed to unknown
			if sameStringSlice(peerAvailability.Peers, []string{localID}) {
				for floor := 0; floor < config.NumFloors; floor++ {
					for btn, hc := range Hallcalls[floor] {
						if hc.OrderState == types.OS_COMPLETED {
							Hallcalls[floor][btn].OrderState = types.OS_UNKNOWN
						}
					}
				}
			}

		case remote := <-ch_rxNetworkMsg:
			if remote.ID != localID {
				for floor := 0; floor < config.NumFloors; floor++ {
					for btn, hc := range Hallcalls[floor] {
						switch hc.OrderState {
						case types.OS_COMPLETED:
							switch remote.HallCalls[floor][btn].OrderState {
							case types.OS_COMPLETED:

							case types.OS_UNCONFIRMED:
								Hallcalls[floor][btn].ExecutorID = remote.HallCalls[floor][btn].ExecutorID
								Hallcalls[floor][btn].AssignerID = remote.HallCalls[floor][btn].AssignerID
								Hallcalls[floor][btn].OrderState = types.OS_UNCONFIRMED
								Hallcalls[floor][btn].AckList = append(Hallcalls[floor][btn].AckList, remote.HallCalls[floor][btn].AckList...)
								Hallcalls[floor][btn].AckList = append(Hallcalls[floor][btn].AckList, localID)
								Hallcalls[floor][btn].AckList = removeDuplicates(Hallcalls[floor][btn].AckList)
							case types.OS_CONFIRMED:
								//
							case types.OS_UNKNOWN:
								//
							}
						case types.OS_UNCONFIRMED:
							switch remote.HallCalls[floor][btn].OrderState {
							case types.OS_COMPLETED:
								//
							case types.OS_CONFIRMED:
								Hallcalls[floor][btn].OrderState = types.OS_CONFIRMED
								fallthrough
							case types.OS_UNCONFIRMED:
								Hallcalls[floor][btn].AckList = append(Hallcalls[floor][btn].AckList, remote.HallCalls[floor][btn].AckList...)
								Hallcalls[floor][btn].AckList = append(Hallcalls[floor][btn].AckList, localID)
								Hallcalls[floor][btn].AckList = removeDuplicates(Hallcalls[floor][btn].AckList)
								//Hallcalls[floor][btn].AckList = make([]string, 0)

							case types.OS_UNKNOWN:
								//
							}
						case types.OS_CONFIRMED:
							switch remote.HallCalls[floor][btn].OrderState {
							case types.OS_COMPLETED:
								Hallcalls[floor][btn].OrderState = types.OS_COMPLETED
								Hallcalls[floor][btn].AckList = make([]string, 0)
								Hallcalls[floor][btn].ExecutorID = ""
								Hallcalls[floor][btn].AssignerID = ""
							case types.OS_UNCONFIRMED:
								//
							case types.OS_CONFIRMED:
								//
							case types.OS_UNKNOWN:
								//
							}
						case types.OS_UNKNOWN:
							switch remote.HallCalls[floor][btn].OrderState {
							case types.OS_COMPLETED:
								Hallcalls[floor][btn].OrderState = types.OS_COMPLETED
								Hallcalls[floor][btn].AckList = make([]string, 0)
								Hallcalls[floor][btn].ExecutorID = ""
								Hallcalls[floor][btn].AssignerID = ""
							case types.OS_UNCONFIRMED:
								Hallcalls[floor][btn].OrderState = types.OS_UNCONFIRMED
								Hallcalls[floor][btn].AckList = append(Hallcalls[floor][btn].AckList, remote.HallCalls[floor][btn].AckList...)
								Hallcalls[floor][btn].AckList = append(Hallcalls[floor][btn].AckList, localID)
								Hallcalls[floor][btn].AckList = removeDuplicates(Hallcalls[floor][btn].AckList)
							case types.OS_CONFIRMED:
								Hallcalls[floor][btn].OrderState = types.OS_CONFIRMED
								//Hallcalls[floor][btn].AckList = make([]string, 0)
							case types.OS_UNKNOWN:
								Hallcalls[floor][btn].OrderState = types.OS_COMPLETED
							}
						}
					}
				}
			}

			if !reflect.DeepEqual(elevators[remote.ID], remote.ElevState) {
				elevators[remote.ID] = utilities.DeepCopyElevatorStruct(remote.ElevState)

			}
		case peerAvailability = <-ch_peerUpdate:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", peerAvailability.Peers)
			fmt.Printf("  New:      %q\n", peerAvailability.New)
			fmt.Printf("  Lost:     %q\n", peerAvailability.Lost)
			//fmt.Println("ElevatorMap: ", elevators)
			ch_informationToAssigner <- types.AssignerMessage{
				PeerList:    utilities.DeepCopyStringSlice(peerAvailability.Peers),
				ElevatorMap: utilities.DeepCopyElevatorMap(elevators),
			}

		}
	}
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
	return len(diff) == 0
}

func generateOurHallcalls(hc [][]types.HallCall, localID string) [][]bool {
	orderMatrix := make([][]bool, config.NumFloors)
	for i := range orderMatrix {
		orderMatrix[i] = make([]bool, config.NumButtons-1)
		for j := range orderMatrix[i] {
			orderMatrix[i][j] = ((hc[i][j].ExecutorID == localID) && hc[i][j].OrderState == types.OS_CONFIRMED)
		}
	}
	return orderMatrix
}

func generateAllHallcalls(hc [][]types.HallCall) [][]bool {
	orderMatrix := make([][]bool, config.NumFloors)
	for i := range orderMatrix {
		orderMatrix[i] = make([]bool, config.NumButtons-1)
		for j := range orderMatrix[i] {
			orderMatrix[i][j] = hc[i][j].OrderState == types.OS_CONFIRMED
		}
	}
	return orderMatrix
}

/*


array of these:
struct {
	state : none, unconfirmed, confirmed, completed
	map[string]struct{} : acks
	assignedTo string/id
}

prevLocalOrders [][]bool

recv from remote: (locally spawned bcast.receiver)
	(ignore msg from self)
	foreach floor, 	foreach button
		v ours | remote >	completed 	unconfirmed 	confirmed 	unknown
		completed			--- 		unconf, +ack	--- 		---
		unconfirmed			--- 		+ack			conf		---
		confirmed			compl		---				---			---
		unknown				completed 	unconf, +ack	confirmed	---



tick: (timer.NewTicker())
	find any that we can confirm:
		foreach unconfirmed
			if all (via peer list) have acked: => confirmed
	send table on net
	generate our orders from big table ([][]orderstate (ours && confirmed) => [][]bool)
		if different from prev => send to whoever needs it (fsm?)
	generate ALL orders (confirmed)
		send to lights


peer list:	(locally spawned peers.receiver)
	if alone on net:
		make all completed into unknown

assigned order	(from assigner)
	if none => unconfirmed

completed order (from fsm)
	if confirmed
		state none, clear ack list








*/
