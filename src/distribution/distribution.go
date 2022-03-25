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
	//retning, en i første tar ikke order i første når den tror den er på vei nedover eller noe 
	//obstruction
	//huske cabcalls

	tick := time.NewTicker(config.TransmitInterval_ms * time.Millisecond)

	//init Hallcalls
	Hallcalls := make([][]types.HallCall, config.NumFloors)
	for i := range Hallcalls {
		Hallcalls[i] = make([]types.HallCall, config.NumButtons-1)
		for j := range Hallcalls[i] {
			Hallcalls[i][j] = types.HallCall{ExecutorID: "", AssignerID: "", OrderState: types.OS_COMPLETED, AckList: make([]string, 0)}
		}
	}

	//init previous local orders
	for i := range prevLocalOrders {
		prevLocalOrders[i] = make([]bool, config.NumButtons-1)
		for j := range prevLocalOrders[i] {
			prevLocalOrders[i][j] = false
		}
	}

	peerAvailability = <-ch_peerUpdate
	elevators[localID] = <-ch_localElevatorState

	ch_informationToAssigner <- types.AssignerMessage{
		PeerStatus: utilities.DeepCopyPeerStatus(peerAvailability),
		ElevatorMap: utilities.DeepCopyElevatorMap(elevators),
	}
	for {

		select {
		case newAssignedOrder := <-ch_assignedOrder:

			//new order for local elevator, gir dette mening? Skal vi sende den før den er confirmed?
			if newAssignedOrder.ID == localID && newAssignedOrder.OrderType.Button == elevio.BT_Cab {
				ch_newLocalOrder <- newAssignedOrder.OrderType

			} //Se på logikken på hvor denne skal være*/
			if newAssignedOrder.OrderType.Button != elevio.BT_Cab &&
				(Hallcalls[newAssignedOrder.OrderType.Floor][newAssignedOrder.OrderType.Button].OrderState == types.OS_COMPLETED ||
				Hallcalls[newAssignedOrder.OrderType.Floor][newAssignedOrder.OrderType.Button].OrderState == types.OS_UNKNOWN) {

				Hallcalls[newAssignedOrder.OrderType.Floor][newAssignedOrder.OrderType.Button].AssignerID = localID
				Hallcalls[newAssignedOrder.OrderType.Floor][newAssignedOrder.OrderType.Button].ExecutorID = newAssignedOrder.ID
				Hallcalls[newAssignedOrder.OrderType.Floor][newAssignedOrder.OrderType.Button].OrderState = types.OS_UNCONFIRMED

				//write our ID on the ack list
				Hallcalls[newAssignedOrder.OrderType.Floor][newAssignedOrder.OrderType.Button].AckList =
					append(Hallcalls[newAssignedOrder.OrderType.Floor][newAssignedOrder.OrderType.Button].AckList, localID)
				Hallcalls[newAssignedOrder.OrderType.Floor][newAssignedOrder.OrderType.Button].AckList =
					removeDuplicates(Hallcalls[newAssignedOrder.OrderType.Floor][newAssignedOrder.OrderType.Button].AckList)
			}

			//send message to network
			ch_txNetworkMsg <- types.NetworkMessage{ID: localID,
				HallCalls: utilities.DeepCopyHallCalls(Hallcalls),
				ElevState: utilities.DeepCopyElevatorStruct(elevators[localID]),
			}

		case e := <-ch_localElevatorState: //change this to compl orders and move this channel to assigner

			//only update when we get something new
			if !reflect.DeepEqual(elevators[localID], e) {
				elevators[localID] = utilities.DeepCopyElevatorStruct(e)

				//send new information to assigner
				ch_informationToAssigner <- types.AssignerMessage{
					PeerStatus: utilities.DeepCopyPeerStatus(peerAvailability),
					ElevatorMap: utilities.DeepCopyElevatorMap(elevators),
				}
			}

		case localCompletedOrder := <-ch_localOrderCompleted:
			//removed this if because of bug when we gave an order in the same floor as the elevator. The order would not be confirmed before the elevator completed it and then the acklist would not be cleared
			if Hallcalls[localCompletedOrder.Floor][localCompletedOrder.Button].OrderState == types.OS_CONFIRMED {

			//set order as completed, clear order
			Hallcalls[localCompletedOrder.Floor][localCompletedOrder.Button].OrderState = types.OS_COMPLETED
			fmt.Println(Hallcalls[localCompletedOrder.Floor][localCompletedOrder.Button].AckList)
			Hallcalls[localCompletedOrder.Floor][localCompletedOrder.Button].AckList = make([]string, 0)
			Hallcalls[localCompletedOrder.Floor][localCompletedOrder.Button].ExecutorID = ""
			Hallcalls[localCompletedOrder.Floor][localCompletedOrder.Button].AssignerID = ""
			//fmt.Println("hc in compl order: ", Hallcalls)
			}
			fmt.Println("Completed orders: ", localCompletedOrder, Hallcalls[localCompletedOrder.Floor][localCompletedOrder.Button])

		case <-tick.C:
			fmt.Println(Hallcalls)
			//confirm orders that have a full ack list
			for floor := 0; floor < config.NumFloors; floor++ {
				for btn, hc := range Hallcalls[floor] {
					if hc.OrderState == types.OS_UNCONFIRMED && equalStringSlice(peerAvailability.Peers, hc.AckList) {
						Hallcalls[floor][btn].OrderState = types.OS_CONFIRMED
					}
				}
			}
			//send message to network
			ch_txNetworkMsg <- types.NetworkMessage{
				ID:        localID,
				HallCalls: utilities.DeepCopyHallCalls(Hallcalls),
				ElevState: utilities.DeepCopyElevatorStruct(elevators[localID]),
			}

			//extract the local elevators hallcalls from Hallcalls, make bool matrix
			ourOrders := utilities.GenerateOurHallcalls(Hallcalls, localID)
			//extract  all hallcalls from Hallcalls, make bool matrix
			allOrders := utilities.GenerateAllHallcalls(Hallcalls)
			//fmt.Println("prev: ", prevLocalOrders)
			//fmt.Println("our: ", ourOrders)
			for floor := 0; floor < config.NumFloors; floor++ {
				for btn, hc := range Hallcalls[floor] {
					
					//check if we have any new orders for us from the network
					if prevLocalOrders[floor][btn] != ourOrders[floor][btn] && hc.ExecutorID == localID {
						ch_newLocalOrder <- elevio.ButtonEvent{Floor: floor, Button: elevio.ButtonType(btn)}
					}
					prevLocalOrders[floor][btn] = ourOrders[floor][btn]

					//set lights
					//må se på button light contract

					elevio.SetButtonLamp(elevio.ButtonType(btn), floor, allOrders[floor][btn])
				}
			}

			// if alone on network, change completed to unknown
			if equalStringSlice(peerAvailability.Peers, []string{localID}) {
				for floor := 0; floor < config.NumFloors; floor++ {
					for btn, hc := range Hallcalls[floor] {
						if hc.OrderState == types.OS_COMPLETED {
							Hallcalls[floor][btn].OrderState = types.OS_UNKNOWN
						}
					}
				}
			}

			//reminder to make functions out ack, add order, clear order
		case remote := <-ch_rxNetworkMsg:
			if remote.ID != localID {
				for floor := 0; floor < config.NumFloors; floor++ {
					for btn, hc := range Hallcalls[floor] {
						switch hc.OrderState {
						case types.OS_COMPLETED:
							switch remote.HallCalls[floor][btn].OrderState {
							case types.OS_COMPLETED:

							case types.OS_UNCONFIRMED:
								//add order to Hallcalls, and ack
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
								//confirm order, and ack, do we need this?
								Hallcalls[floor][btn].OrderState = types.OS_CONFIRMED
								fallthrough
							case types.OS_UNCONFIRMED:
								//ack
								Hallcalls[floor][btn].AckList = append(Hallcalls[floor][btn].AckList, remote.HallCalls[floor][btn].AckList...)
								Hallcalls[floor][btn].AckList = append(Hallcalls[floor][btn].AckList, localID)
								Hallcalls[floor][btn].AckList = removeDuplicates(Hallcalls[floor][btn].AckList)

							case types.OS_UNKNOWN:
								//
							}
						case types.OS_CONFIRMED:
							switch remote.HallCalls[floor][btn].OrderState {
							case types.OS_COMPLETED:
								//change to completed and clear order
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
								//change to completed, clear order
								Hallcalls[floor][btn].OrderState = types.OS_COMPLETED
								Hallcalls[floor][btn].AckList = make([]string, 0)
								Hallcalls[floor][btn].ExecutorID = ""
								Hallcalls[floor][btn].AssignerID = ""
							case types.OS_UNCONFIRMED:
								//add order, and ack
								Hallcalls[floor][btn].OrderState = types.OS_UNCONFIRMED
								Hallcalls[floor][btn].AckList = append(Hallcalls[floor][btn].AckList, remote.HallCalls[floor][btn].AckList...)
								Hallcalls[floor][btn].AckList = append(Hallcalls[floor][btn].AckList, localID)
								Hallcalls[floor][btn].AckList = removeDuplicates(Hallcalls[floor][btn].AckList)
							case types.OS_CONFIRMED:
								//confirm
								Hallcalls[floor][btn].OrderState = types.OS_CONFIRMED
								//Hallcalls[floor][btn].AckList = make([]string, 0)
							case types.OS_UNKNOWN:
								//needed this because of scenario where the elevators are initialized simultaneously, and the peer list is not updated before the tick sets all orders to unknown
								//recieving unknown means there is another elevator sending, so orders should be set to completed
								Hallcalls[floor][btn].OrderState = types.OS_COMPLETED
							}
						}
					}
				}
			}

			//update elevator map with new information from remote
			if !reflect.DeepEqual(elevators[remote.ID], remote.ElevState) {
				elevators[remote.ID] = utilities.DeepCopyElevatorStruct(remote.ElevState)
				//send information to assigner
				ch_informationToAssigner <- types.AssignerMessage{
					PeerStatus: utilities.DeepCopyPeerStatus(peerAvailability),
					ElevatorMap: utilities.DeepCopyElevatorMap(elevators),
				}
			}
		case peerAvailability = <-ch_peerUpdate:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", peerAvailability.Peers)
			fmt.Printf("  New:      %q\n", peerAvailability.New)
			fmt.Printf("  Lost:     %q\n", peerAvailability.Lost)
			//fmt.Println("ElevatorMap: ", elevators)
			if len(peerAvailability.Lost) != 0 {
				//reassign orders to myself
				for _, id := range peerAvailability.Lost {
					for i := 0; i < config.NumFloors; i++ {
						for j, hc := range Hallcalls[i] {
							if hc.ExecutorID == id {
								switch hc.OrderState {
								case types.OS_COMPLETED:								
								case types.OS_UNCONFIRMED:
									Hallcalls[i][j].ExecutorID = localID
									Hallcalls[i][j].OrderState = hc.OrderState
									Hallcalls[i][j].AckList = make([]string, 0)
									Hallcalls[i][j].AckList = append(Hallcalls[i][j].AckList, localID)
								case types.OS_CONFIRMED:
									Hallcalls[i][j].ExecutorID = localID
									Hallcalls[i][j].OrderState = hc.OrderState
									Hallcalls[i][j].AckList = make([]string, 0)
								case types.OS_UNKNOWN:
								}
							}
						}
					}
				}
			}
			ch_informationToAssigner <- types.AssignerMessage{
				PeerStatus:    utilities.DeepCopyPeerStatus(peerAvailability),
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

func equalStringSlice(x, y []string) bool {
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
