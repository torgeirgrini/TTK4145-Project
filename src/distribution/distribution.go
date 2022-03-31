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
	ch_peerStatusUpdate chan<- peers.PeerUpdate,
	ch_assignedOrder <-chan types.AssignedOrder,
	ch_newLocalOrder chan<- elevio.ButtonEvent,
	ch_localOrderCompleted <-chan elevio.ButtonEvent,
	ch_reassignHallCalls <-chan string,
) {
	tick := time.NewTicker(config.TransmitInterval_ms * time.Millisecond)

	ch_txHallCalls := make(chan types.HallCallsNetMsg, config.NumElevators) //spør om bufferstr/nødvendig
	ch_rxHallCalls := make(chan types.HallCallsNetMsg, config.NumElevators)
	ch_peerUpdate := make(chan peers.PeerUpdate, config.NumElevators)
	ch_peerTxEnable := make(chan bool)
	ch_tick := tick.C

	go bcast.Transmitter(config.PortBroadcast, ch_txHallCalls)
	go bcast.Receiver(config.PortBroadcast, ch_rxHallCalls)
	go peers.Transmitter(config.PortPeers, localID, ch_peerTxEnable)
	go peers.Receiver(config.PortPeers, ch_peerUpdate)

	var peerAvailability peers.PeerUpdate
	peerAvailability = peers.PeerUpdate{
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

	Hallcalls := make([][]types.HallCall, config.NumFloors)
	for i := range Hallcalls {
		Hallcalls[i] = make([]types.HallCall, config.NumButtons-1)
		for j := range Hallcalls[i] {
			Hallcalls[i][j] = types.HallCall{ExecutorID: "", OrderState: types.OS_Unknown, AckList: make([]string, 0)}
		}
	}

	ch_peerStatusUpdate <- utilities.DeepCopyPeerStatus(peerAvailability)

	for {
		select {
		case newAssignedOrder := <-ch_assignedOrder:
			if newAssignedOrder.ID == localID && newAssignedOrder.OrderType.Button == elevio.BT_Cab {
				ch_newLocalOrder <- newAssignedOrder.OrderType
			}
			if newAssignedOrder.OrderType.Button != elevio.BT_Cab &&
				(Hallcalls[newAssignedOrder.OrderType.Floor][newAssignedOrder.OrderType.Button].OrderState == types.OS_Completed ||
					Hallcalls[newAssignedOrder.OrderType.Floor][newAssignedOrder.OrderType.Button].OrderState == types.OS_Unknown) {

				Hallcalls[newAssignedOrder.OrderType.Floor][newAssignedOrder.OrderType.Button].ExecutorID = newAssignedOrder.ID
				Hallcalls[newAssignedOrder.OrderType.Floor][newAssignedOrder.OrderType.Button].OrderState = types.OS_Unconfirmed
				Hallcalls[newAssignedOrder.OrderType.Floor][newAssignedOrder.OrderType.Button].AckList =
					append(Hallcalls[newAssignedOrder.OrderType.Floor][newAssignedOrder.OrderType.Button].AckList, localID)
				Hallcalls[newAssignedOrder.OrderType.Floor][newAssignedOrder.OrderType.Button].AckList =
					utilities.RemoveDuplicatesSlice(Hallcalls[newAssignedOrder.OrderType.Floor][newAssignedOrder.OrderType.Button].AckList)
			}

		case localCompletedOrder := <-ch_localOrderCompleted:
			if Hallcalls[localCompletedOrder.Floor][localCompletedOrder.Button].OrderState == types.OS_Confirmed {
				Hallcalls[localCompletedOrder.Floor][localCompletedOrder.Button].OrderState = types.OS_Completed
				clearHallcall(localCompletedOrder.Floor, int(localCompletedOrder.Button), Hallcalls)
			}
		case <-ch_tick:
			for floor := 0; floor < config.NumFloors; floor++ {
				for btn, hc := range Hallcalls[floor] {
					if hc.OrderState == types.OS_Unconfirmed && utilities.EqualStringSlice(peerAvailability.Peers, hc.AckList) {
						Hallcalls[floor][btn].OrderState = types.OS_Confirmed
					}
				}
			}
			ch_txHallCalls <- types.HallCallsNetMsg{
				SenderID:  localID,
				HallCalls: utilities.DeepCopyHallCalls(Hallcalls),
			}
			//check if we have any new orders for us from the network
			ourOrders := generateOurHallcalls(Hallcalls, localID)
			allOrders := generateAllHallcalls(Hallcalls)
			for floor := 0; floor < config.NumFloors; floor++ {
				for btn, hc := range Hallcalls[floor] {
					if prevLocalOrders[floor][btn] != ourOrders[floor][btn] && hc.ExecutorID == localID {
						ch_newLocalOrder <- elevio.ButtonEvent{Floor: floor, Button: elevio.ButtonType(btn)}
					}
					prevLocalOrders[floor][btn] = ourOrders[floor][btn]
					//må se på button light contract
					elevio.SetButtonLamp(elevio.ButtonType(btn), floor, allOrders[floor][btn])
				}
			}
			
			// if alone on network, change completed to unknown
			if utilities.EqualStringSlice(peerAvailability.Peers, []string{localID}) {
				for floor := 0; floor < config.NumFloors; floor++ {
					for btn, hc := range Hallcalls[floor] {
						if hc.OrderState == types.OS_Completed {
							Hallcalls[floor][btn].OrderState = types.OS_Unknown
						}
					}
				}
			}
		case remote := <-ch_rxHallCalls:
			if remote.SenderID != localID {
				for floor := 0; floor < config.NumFloors; floor++ {
					for btn, hc := range Hallcalls[floor] {
						switch hc.OrderState {
						case types.OS_Completed:
							switch remote.HallCalls[floor][btn].OrderState {
							case types.OS_Completed:
							case types.OS_Unconfirmed:
								Hallcalls[floor][btn].ExecutorID = remote.HallCalls[floor][btn].ExecutorID
								Hallcalls[floor][btn].OrderState = types.OS_Unconfirmed
								appendToAckList(Hallcalls, remote.HallCalls, floor, btn, localID)
							case types.OS_Confirmed:
							case types.OS_Unknown:
							}
						case types.OS_Unconfirmed:
							switch remote.HallCalls[floor][btn].OrderState {
							case types.OS_Completed:
							case types.OS_Confirmed:
								Hallcalls[floor][btn].OrderState = types.OS_Confirmed
								fallthrough
							case types.OS_Unconfirmed:
								appendToAckList(Hallcalls, remote.HallCalls, floor, btn, localID)
							case types.OS_Unknown:
							}
						case types.OS_Confirmed:
							switch remote.HallCalls[floor][btn].OrderState {
							case types.OS_Completed:
								Hallcalls[floor][btn].OrderState = types.OS_Completed
								clearHallcall(floor, btn, Hallcalls)
							case types.OS_Unconfirmed:
							case types.OS_Confirmed:
							case types.OS_Unknown:
							}
						case types.OS_Unknown:
							switch remote.HallCalls[floor][btn].OrderState {
							case types.OS_Completed:
								Hallcalls[floor][btn].OrderState = types.OS_Completed
								clearHallcall(floor, btn, Hallcalls)
							case types.OS_Unconfirmed:
								Hallcalls[floor][btn].ExecutorID = remote.HallCalls[floor][btn].ExecutorID
								Hallcalls[floor][btn].OrderState = types.OS_Unconfirmed
								appendToAckList(Hallcalls, remote.HallCalls, floor, btn, localID)
							case types.OS_Confirmed:
								Hallcalls[floor][btn].ExecutorID = remote.HallCalls[floor][btn].ExecutorID
								Hallcalls[floor][btn].OrderState = types.OS_Confirmed
								appendToAckList(Hallcalls, remote.HallCalls, floor, btn, localID)
							case types.OS_Unknown:
							}
						}
					}
				}
			}
		case reassignID := <- ch_reassignHallCalls:
			reassignHallcalls(reassignID, Hallcalls, localID)
		case peerAvailability = <-ch_peerUpdate:
			peerAvailability.Peers = utilities.RemoveDuplicatesSlice(append(utilities.DeepCopyStringSlice(peerAvailability.Peers), localID))
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", peerAvailability.Peers)
			fmt.Printf("  New:      %q\n", peerAvailability.New)
			fmt.Printf("  Lost:     %q\n", peerAvailability.Lost)
			if len(peerAvailability.Lost) != 0 {
				for _, id := range peerAvailability.Lost {
					reassignHallcalls(id, Hallcalls, localID)
				}
			}
			ch_peerStatusUpdate <- utilities.DeepCopyPeerStatus(peerAvailability)
		}
	}
}

/* FUNCTIONS */

func reassignHallcalls(reassignID string, hallCalls [][]types.HallCall, assignToID string) {
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
}

func generateOurHallcalls(hc [][]types.HallCall, elevatorID string) [][]bool {
	orderMatrix := make([][]bool, config.NumFloors)
	for i := range orderMatrix {
		orderMatrix[i] = make([]bool, config.NumButtons-1)
		for j := range orderMatrix[i] {
			orderMatrix[i][j] = ((hc[i][j].ExecutorID == elevatorID) && hc[i][j].OrderState == types.OS_Confirmed)
		}
	}
	return orderMatrix
}

func generateAllHallcalls(hc [][]types.HallCall) [][]bool {
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

func clearHallcall(floor int, btn int, Hallcalls [][]types.HallCall) {
	Hallcalls[floor][btn].AckList = make([]string, 0)
	Hallcalls[floor][btn].ExecutorID = ""
}
