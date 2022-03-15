package network

import (
	"Project/config"
	"Project/localElevator/elevator"
	"Project/network/peers"
	"flag"
	"fmt"
)

// We define some custom struct to send over the network.
// Note that all members we want to transmit must be public. Any private members
//  will be received as zero-values.
type HelloMsg struct {
	Message string
	Iter    int
}

//
func Network(
	id string,
	ch_txEsm chan<- map[string]elevator.Elevator,
	ch_rxEsm <-chan map[string]elevator.Elevator,
	ch_localElevatorStruct <-chan elevator.Elevator) {
	// Our id can be anything. Here we pass it on the command line, using
	//  `go run main.go -id=our_id`

	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	// We make a channel for receiving updates on the id's of the peers that are
	//  alive on the network
	peerUpdateCh := make(chan peers.PeerUpdate)
	// We can disable/enable the transmitter after it has been started.
	// This could be used to signal that we are somehow "unavailable".
	peerTxEnable := make(chan bool)
	go peers.Transmitter(config.PortPeers, id, peerTxEnable)
	go peers.Receiver(config.PortPeers, peerUpdateCh)

	// We make channels for sending and receiving our custom data types
	// helloTx := make(chan elevator.Elevator)
	// helloRx := make(chan elevator.Elevator)
	// ... and start the transmitter/receiver pair on some port
	// These functions can take any number of channels! It is also possible to
	//  start multiple transmitters/receivers on the same port.
	// go bcast.Transmitter(config.PortBroadcast, helloTx)
	// go bcast.Receiver(config.PortBroadcast, helloRx)

	// go bcast.Transmitter(config.PortBroadcast, ch_txEsm)
	// go bcast.Receiver(config.PortBroadcast, ch_rxEsm)

	// The example message. We just send one of these every second.
	go func() {

		for {
			// elev := <-ch_localElevatorStruct
			// helloMsg := elev
			// //elev2 := elevator.InitElev()
			// //helloMsg := make(map[string]elevator.Elevator)
			// //helloMsg[id] = elev
			// //helloMsg[id+"1"] = elev2
			// //helloMsg := HelloMsg{"Elevator floor: " + strconv.Itoa(elev.Floor) + " and id is " + id, 0}
			// //helloMsg.Iter++
			// helloTx <- helloMsg
			//ch_txEsm <- helloMsg
		}
	}()

	fmt.Println("Started")
	for {
		select {
		case p := <-peerUpdateCh:
			fmt.Printf("Peer update:\n")
			fmt.Printf("  Peers:    %q\n", p.Peers)
			fmt.Printf("  New:      %q\n", p.New)
			fmt.Printf("  Lost:     %q\n", p.Lost)

			// case a := <-helloRx:
			// 	fmt.Println("Received:")
			// 	elevator.PrintElevator(a)
			// 	fmt.Println("END")
			// case m := <-ch_rxEsm:
			// 	fmt.Println("MAP:", m)
			// }
		}
	}
}


func () {

	for {
		// elev := <-ch_localElevatorStruct
		// helloMsg := elev
		// //elev2 := elevator.InitElev()
		// //helloMsg := make(map[string]elevator.Elevator)
		// //helloMsg[id] = elev
		// //helloMsg[id+"1"] = elev2
		// //helloMsg := HelloMsg{"Elevator floor: " + strconv.Itoa(elev.Floor) + " and id is " + id, 0}
		// //helloMsg.Iter++
		// helloTx <- helloMsg
		//ch_txEsm <- helloMsg
	}
}()
