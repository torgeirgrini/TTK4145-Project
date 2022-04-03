package distribution

import (
	"Project/config"
	"Project/localElevator/elevio"
	"Project/network/bcast"
	"Project/network/peers"
	"Project/types"
	"Project/utilities"
	"fmt"
	"time"
)

func Distribution(
	localID string,
	ch_peerStatus chan<- peers.PeerUpdate,
	ch_assignedOrder <-chan types.AssignedOrder,
	ch_newLocalOrder chan<- elevio.ButtonEvent,
	ch_localOrderCompleted <-chan elevio.ButtonEvent,
	ch_stuck <-chan bool,
) {
	tick := time.NewTicker(config.TransmitInterval_ms * time.Millisecond)

	ch_txHallCalls := make(chan types.HallCallsNetMsg, config.NumElevators) 
	ch_rxHallCalls := make(chan types.HallCallsNetMsg, config.NumElevators)
	ch_peerUpdate := make(chan peers.PeerUpdate, config.NumElevators)
	ch_peerTxEnable := make(chan bool)
	ch_tick := tick.C

	go bcast.Transmitter(config.PortBroadcast, ch_txHallCalls)
	go bcast.Receiver(config.PortBroadcast, ch_rxHallCalls)
	go peers.Transmitter(config.PortPeersDist, localID, ch_peerTxEnable)
	go peers.Receiver(config.PortPeersDist, ch_peerUpdate)

	peerAvailability := peers.PeerUpdate{
		Peers: []string{localID},
		New:   "",
		Lost:  make([]string, 0),
	}

	prevLocalOrders := make([][]bool, config.NumFloors)
	for i := range prevLocalOrders {
		prevLocalOrders[i] = make([]bool, config.NumButtons-1)
		for j := range prevLocalOrders[i] {
			prevLocalOrders[i][j] = false
		}
	}

	hallCalls := make([][]types.HallCall, config.NumFloors)
	for i := range hallCalls {
		hallCalls[i] = make([]types.HallCall, config.NumButtons-1)
		for j := range hallCalls[i] {
			hallCalls[i][j] = types.HallCall{ExecutorID: "", OrderState: types.OS_Unknown, AckList: make([]string, 0)}
		}
	}

	unavailableSet := make(map[string]types.Void) 
	
	ch_peerStatus <- utilities.DeepCopyPeerStatus(peerAvailability)
	
	for {
		select {
		case order := <-ch_assignedOrder:
			if order.ID == localID && order.OrderType.Button == elevio.BT_Cab {
				ch_newLocalOrder <- order.OrderType
			}
			if order.OrderType.Button != elevio.BT_Cab &&
				(hallCalls[order.OrderType.Floor][order.OrderType.Button].OrderState == types.OS_Completed ||
					hallCalls[order.OrderType.Floor][order.OrderType.Button].OrderState == types.OS_Unknown) {

				hallCalls[order.OrderType.Floor][order.OrderType.Button].ExecutorID = order.ID
				hallCalls[order.OrderType.Floor][order.OrderType.Button].OrderState = types.OS_Unconfirmed
				hallCalls[order.OrderType.Floor][order.OrderType.Button].AckList =
					append(hallCalls[order.OrderType.Floor][order.OrderType.Button].AckList, localID)
				hallCalls[order.OrderType.Floor][order.OrderType.Button].AckList =
					utilities.RemoveDuplicatesSlice(hallCalls[order.OrderType.Floor][order.OrderType.Button].AckList)
			}

		case compOrder := <-ch_localOrderCompleted:
			if hallCalls[compOrder.Floor][compOrder.Button].OrderState == types.OS_Confirmed {
				hallCalls[compOrder.Floor][compOrder.Button].OrderState = types.OS_Completed
				clearHallCall(compOrder.Floor, int(compOrder.Button), hallCalls)
			}
		case <-ch_tick:
			for floor := 0; floor < config.NumFloors; floor++ {
				for btn, hc := range hallCalls[floor] {
					if hc.OrderState == types.OS_Unconfirmed && utilities.ContainsStringSlice(hc.AckList, peerAvailability.Peers) {
						hallCalls[floor][btn].OrderState = types.OS_Confirmed
					}
				}
			}

			ourOrders := generateOurHallCalls(hallCalls, localID)
			allOrders := generateAllHallCalls(hallCalls)
			for floor := 0; floor < config.NumFloors; floor++ {
				for btn, hc := range hallCalls[floor] {
					if prevLocalOrders[floor][btn] != ourOrders[floor][btn] && hc.ExecutorID == localID {
						ch_newLocalOrder <- elevio.ButtonEvent{Floor: floor, Button: elevio.ButtonType(btn)}
					}
					prevLocalOrders[floor][btn] = ourOrders[floor][btn]
					elevio.SetButtonLamp(elevio.ButtonType(btn), floor, allOrders[floor][btn])
				}
			}

			for _, id := range peerAvailability.Peers {
				if _, ok := unavailableSet[id]; ok {
					hallCalls = reassignHallcalls(hallCalls, id,localID)
				}
			}
			if utilities.EqualStringSlice(peerAvailability.Peers, []string{localID}) {
				for floor := 0; floor < config.NumFloors; floor++ {
					for btn, hc := range hallCalls[floor] {
						if hc.OrderState == types.OS_Completed {
							hallCalls[floor][btn].OrderState = types.OS_Unknown
						}
					}
				}
			}			
			ch_txHallCalls <- types.HallCallsNetMsg{
				SenderID:  localID,
				HallCalls: utilities.DeepCopyHallCalls(hallCalls),
			}
		case remote := <-ch_rxHallCalls:
			if remote.SenderID != localID {
				for floor := 0; floor < config.NumFloors; floor++ {
					for btn, hc := range hallCalls[floor] {
						switch hc.OrderState {
						case types.OS_Completed:
							switch remote.HallCalls[floor][btn].OrderState {
							case types.OS_Completed:
							case types.OS_Unconfirmed:
								hallCalls[floor][btn].ExecutorID = remote.HallCalls[floor][btn].ExecutorID
								hallCalls[floor][btn].OrderState = types.OS_Unconfirmed
								appendToAckList(hallCalls, remote.HallCalls, floor, btn, localID)
							case types.OS_Confirmed:
							case types.OS_Unknown:
							}
						case types.OS_Unconfirmed:
							switch remote.HallCalls[floor][btn].OrderState {
							case types.OS_Completed:
							case types.OS_Confirmed:
								hallCalls[floor][btn].OrderState = types.OS_Confirmed
								fallthrough
							case types.OS_Unconfirmed:
								appendToAckList(hallCalls, remote.HallCalls, floor, btn, localID)
							case types.OS_Unknown:
							}
						case types.OS_Confirmed:
							switch remote.HallCalls[floor][btn].OrderState {
							case types.OS_Completed:
								hallCalls[floor][btn].OrderState = types.OS_Completed
								clearHallCall(floor, btn, hallCalls)
							case types.OS_Unconfirmed:
							case types.OS_Confirmed:
							case types.OS_Unknown:
							}
						case types.OS_Unknown:
							switch remote.HallCalls[floor][btn].OrderState {
							case types.OS_Completed:
								hallCalls[floor][btn].OrderState = types.OS_Completed
								clearHallCall(floor, btn, hallCalls)
							case types.OS_Unconfirmed:
								hallCalls[floor][btn].ExecutorID = remote.HallCalls[floor][btn].ExecutorID
								hallCalls[floor][btn].OrderState = types.OS_Unconfirmed
								appendToAckList(hallCalls, remote.HallCalls, floor, btn, localID)
							case types.OS_Confirmed:
								hallCalls[floor][btn].ExecutorID = remote.HallCalls[floor][btn].ExecutorID
								hallCalls[floor][btn].OrderState = types.OS_Confirmed
								appendToAckList(hallCalls, remote.HallCalls, floor, btn, localID)
							case types.OS_Unknown:
							}
						}
					}
				}
			}
		case peerAvailability = <-ch_peerUpdate:
			fmt.Println(peerAvailability)
			/*if len(peerAvailability.Lost) != 0 {
				for _, id := range peerAvailability.Lost {
					hallCalls = reassignHallcalls(hallCalls, id,localID)
					unavailableSet[id] = types.Void{}
				}
			} else if peerAvailability.New != "" {
				delete(unavailableSet, peerAvailability.New)
			}
			ch_peerStatus <- utilities.DeepCopyPeerStatus(peerAvailability)
		case localElevStuck := <-ch_stuck:
			ch_peerTxEnable <- !localElevStuck*/
			if len(peerAvailability.Lost) != 0 {
				for _, id := range peerAvailability.Lost {
					hallCalls = reassignHallcalls( hallCalls, id, localID)
					unavailableSet[id] = types.Void{}
				}
			} else if peerAvailability.New != "" {
					delete(unavailableSet, peerAvailability.New)
			}
			ch_peerStatus <- utilities.DeepCopyPeerStatus(peerAvailability)
		case localElevStuck := <-ch_stuck:
			if localElevStuck {
				fmt.Println("hillo")
				unavailableSet[localID] = types.Void{}
			}
			ch_peerTxEnable <- !localElevStuck
		}
	}
}

func reassignHallcalls(HC [][]types.HallCall, reassignID string, assignToID string) [][]types.HallCall{
	hallCalls := utilities.DeepCopyHallCalls(HC)
	for i := 0; i < config.NumFloors; i++ {
		for j, hc := range hallCalls[i] {
			if hc.ExecutorID == reassignID {
				switch hc.OrderState {
				case types.OS_Completed:
				case types.OS_Unconfirmed:
					hallCalls[i][j].ExecutorID = assignToID
					hallCalls[i][j].OrderState = hc.OrderState
					hallCalls[i][j].AckList = make([]string, 0)
					hallCalls[i][j].AckList = append(hallCalls[i][j].AckList, assignToID)
				case types.OS_Confirmed:
					hallCalls[i][j].ExecutorID = assignToID
					hallCalls[i][j].OrderState = hc.OrderState
					hallCalls[i][j].AckList = make([]string, 0)
				case types.OS_Unknown:
				}
			}
		}
	}
	return hallCalls
}

func generateOurHallCalls(hc [][]types.HallCall, elevatorID string) [][]bool {
	orderMatrix := make([][]bool, config.NumFloors)
	for i := range orderMatrix {
		orderMatrix[i] = make([]bool, config.NumButtons-1)
		for j := range orderMatrix[i] {
			orderMatrix[i][j] = ((hc[i][j].ExecutorID == elevatorID) && hc[i][j].OrderState == types.OS_Confirmed)
		}
	}
	return orderMatrix
}

func generateAllHallCalls(hc [][]types.HallCall) [][]bool {
	orderMatrix := make([][]bool, config.NumFloors)
	for i := range orderMatrix {
		orderMatrix[i] = make([]bool, config.NumButtons-1)
		for j := range orderMatrix[i] {
			orderMatrix[i][j] = hc[i][j].OrderState == types.OS_Confirmed
		}
	}
	return orderMatrix
}

func appendToAckList(localHC [][]types.HallCall, remoteHC [][]types.HallCall, floor int, btn int, ID string) {
	localHC[floor][btn].AckList = append(localHC[floor][btn].AckList, remoteHC[floor][btn].AckList...)
	localHC[floor][btn].AckList = append(localHC[floor][btn].AckList, ID)
	localHC[floor][btn].AckList = utilities.RemoveDuplicatesSlice(localHC[floor][btn].AckList)
}

func clearHallCall(floor int, btn int, hallCalls [][]types.HallCall) {
	hallCalls[floor][btn].AckList = make([]string, 0)
	hallCalls[floor][btn].ExecutorID = ""
}
