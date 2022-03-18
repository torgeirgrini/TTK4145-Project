package network
/*

func Network(
	id string,
	ch_TxNewElevatorStateMap chan ElevatorStateMessage,
	ch_RxNewElevatorStateMap chan ElevatorStateMessage,
	ch_newLocalState <-chan elevator.Elevator,
	ch_peerTxEnable <-chan bool) {

	// Our id can be anything. Here we pass it on the command line, using
	//  `go run main.go -id=our_id`

	ch_peerUpdate := make(chan peers.PeerUpdate)

	go peers.Transmitter(config.PortPeers, id, ch_peerTxEnable)
	go peers.Receiver(config.PortPeers, ch_peerUpdate)

	go bcast.Transmitter(config.PortBroadcast, ch_TxNewElevatorStateMap)
	go bcast.Receiver(config.PortBroadcast, ch_RxNewElevatorStateMap)

	//go PeriodicTransmit(ch_TxNewElevatorStateMap)
	go Receive(ch_RxNewElevatorStateMap)
	fmt.Println("Started")
	//FSM for å motta fra nettet
	for {
		select {
		case peerUpdate := <-ch_peerUpdate:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", peerUpdate.Peers)
			fmt.Printf("  New:      %q\n", peerUpdate.New)
			fmt.Printf("  Lost:     %q\n", peerUpdate.Lost)
			//Må si ifra om at noen har kommet på/fallt av nettet
		case NewElevatorStateMap := <-ch_RxNewElevatorStateMap:
			_ = NewElevatorStateMap
			//inn på en ny channel
		}
	}
}
*/
