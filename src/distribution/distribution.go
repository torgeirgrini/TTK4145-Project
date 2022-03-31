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

func Distribution(
	localID string,
	ch_localElevatorState <-chan types.Elevator,
	ch_informationToAssigner chan<- types.AssignerMessage,
	ch_assignedOrder <-chan types.AssignedOrder,
	ch_newLocalOrder chan<- elevio.ButtonEvent,
	ch_localOrderCompleted <-chan elevio.ButtonEvent,
) {
	tick := time.NewTicker(config.TransmitInterval_ms * time.Millisecond)

	ch_txNetworkMsg := make(chan types.NetworkMessage, config.NumElevators) //spør om bufferstr/nødvendig
	ch_rxNetworkMsg := make(chan types.NetworkMessage, config.NumElevators)
	ch_newAssignedOrderToNetwork := make(chan types.AssignedOrder)
	ch_readNewOrderFromNetwork := make(chan types.AssignedOrder)
	ch_peerUpdate := make(chan peers.PeerUpdate, config.NumElevators)
	ch_peerTxEnable := make(chan bool)
	ch_tick := tick.C

	go bcast.Transmitter(config.PortBroadcast, ch_txNetworkMsg, ch_newAssignedOrderToNetwork)
	go bcast.Receiver(config.PortBroadcast, ch_rxNetworkMsg, ch_readNewOrderFromNetwork)
	go peers.Transmitter(config.PortPeers, localID, ch_peerTxEnable)
	go peers.Receiver(config.PortPeers, ch_peerUpdate)

	var peerAvailability peers.PeerUpdate
	peerAvailability = peers.PeerUpdate{
		Peers: []string{localID},
		New:   "",
		Lost:  make([]string, 0),
	}

	elevators := make(map[string]types.Elevator)
	elevators[localID] = <-ch_localElevatorState

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

	InitTimer := time.NewTimer(time.Duration(3) * time.Second)
	ch_initTimer := InitTimer.C
	init := true
	for init {
		select {
		case <-ch_initTimer:
			init = false
		case initState := <-ch_rxNetworkMsg:
			if initState.ElevStateID == localID {
				for floor := 0; floor < config.NumFloors; floor++ {
					if initState.ElevState.Requests[floor][elevio.BT_Cab] {
						ch_newLocalOrder <- elevio.ButtonEvent{
							Floor:  floor,
							Button: elevio.BT_Cab,
						}
					}
				}
				init = false

			}
		case e := <-ch_localElevatorState:
			elevators[localID] = e
		default:
		}
	}
	ch_informationToAssigner <- types.AssignerMessage{
		PeerStatus:  utilities.DeepCopyPeerStatus(peerAvailability),
		ElevatorMap: utilities.DeepCopyElevatorMap(elevators),
	}

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

			ch_txNetworkMsg <- types.NetworkMessage{
				SenderID:    localID,
				ElevStateID: localID,
				HallCalls:   utilities.DeepCopyHallCalls(Hallcalls),
				ElevState:   utilities.DeepCopyElevatorStruct(elevators[localID]),
			}
		case e := <-ch_localElevatorState:
			if !reflect.DeepEqual(elevators[localID], e) {
				elevators[localID] = utilities.DeepCopyElevatorStruct(e)
				ch_informationToAssigner <- types.AssignerMessage{
					PeerStatus:  utilities.DeepCopyPeerStatus(peerAvailability),
					ElevatorMap: utilities.DeepCopyElevatorMap(elevators),
				}
			}
		case localCompletedOrder := <-ch_localOrderCompleted:
			if Hallcalls[localCompletedOrder.Floor][localCompletedOrder.Button].OrderState == types.OS_Confirmed {
				Hallcalls[localCompletedOrder.Floor][localCompletedOrder.Button].OrderState = types.OS_Completed
				clearHallcall(localCompletedOrder.Floor, int(localCompletedOrder.Button), Hallcalls)
			}
		case <-ch_tick:
			/*
				cmd := exec.Command("clear")
				cmd.Stdout = os.Stdout
				cmd.Run()
			*/
			fmt.Println("local available:", elevators[localID].Available)
			for floor := 0; floor < config.NumFloors; floor++ {
				for btn, hc := range Hallcalls[floor] {
					if hc.OrderState == types.OS_Unconfirmed && utilities.EqualStringSlice(peerAvailability.Peers, hc.AckList) {
						Hallcalls[floor][btn].OrderState = types.OS_Confirmed
					}
				}
			}
			ch_txNetworkMsg <- types.NetworkMessage{
				SenderID:    localID,
				ElevStateID: localID,
				HallCalls:   utilities.DeepCopyHallCalls(Hallcalls),
				ElevState:   utilities.DeepCopyElevatorStruct(elevators[localID]),
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
			for _, id := range peerAvailability.Peers {
				if !elevators[id].Available {
					reassignHallcalls(id, Hallcalls, localID)
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
		case remote := <-ch_rxNetworkMsg:
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
				//update elevator map with new information from remote
				if !reflect.DeepEqual(elevators[remote.ElevStateID], remote.ElevState) && remote.ElevStateID == remote.SenderID {
					elevators[remote.ElevStateID] = utilities.DeepCopyElevatorStruct(remote.ElevState)
					//fmt.Println("ID, Avalaible: ", remote.ElevStateID, remote.ElevState.Available)
					ch_informationToAssigner <- types.AssignerMessage{
						PeerStatus:  utilities.DeepCopyPeerStatus(peerAvailability),
						ElevatorMap: utilities.DeepCopyElevatorMap(elevators),
					}
				}

			}
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
			if peerAvailability.New != localID && peerAvailability.New != "" {
				if _, ok := elevators[peerAvailability.New]; ok {
					ch_txNetworkMsg <- types.NetworkMessage{
						SenderID:    localID,
						ElevStateID: peerAvailability.New,
						HallCalls:   utilities.DeepCopyHallCalls(Hallcalls),
						ElevState:   utilities.DeepCopyElevatorStruct(elevators[peerAvailability.New]),
					}
				}
			}
			ch_informationToAssigner <- types.AssignerMessage{
				PeerStatus:  utilities.DeepCopyPeerStatus(peerAvailability),
				ElevatorMap: utilities.DeepCopyElevatorMap(elevators),
			}
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
