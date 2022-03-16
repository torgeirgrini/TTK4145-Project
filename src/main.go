package main

import (
	"Project/config"
	"Project/localElevator/elevator"
	"Project/localElevator/elevio"
	"Project/localElevator/fsm"
	"Project/network"
	"flag"
	"fmt"
)

func main() {
	var id string
	var p string
	flag.StringVar(&p, "p", "15657", "port to elevator server")
	flag.StringVar(&id, "id", "", "id of this peer")
	flag.Parse()

	elevio.Init("localhost:"+p, config.NumFloors)

	//Hardware channels
	ch_drv_buttons := make(chan elevio.ButtonEvent)
	ch_drv_floors := make(chan int)
	ch_drv_obstr := make(chan bool)

	//Assigner channels
	//ch_elevatorSystemMap := make(chan map[string]elevator.Elevator)

	//Network channels
	ch_txEsm := make(chan network.ElevatorStateMessage)
	ch_rxEsm := make(chan network.ElevatorStateMessage)
	allElevators := make(chan map[string]elevator.Elevator)

	//Local elevator state channel
	ch_localElevatorStruct := make(chan elevator.Elevator)
	//Local elevator channels
	//ch_newLocalState = make(chan elevator.Elevator)
	//ch_localOrder = make(chan elevator.Elevator)

	ch_peerTxEnable := make(chan bool)

	go elevio.PollButtons(ch_drv_buttons)
	go elevio.PollFloorSensor(ch_drv_floors)
	go elevio.PollObstructionSwitch(ch_drv_obstr)

	go fsm.RunElevator(ch_drv_buttons, ch_drv_floors, ch_drv_obstr, ch_localElevatorStruct)

	go network.Network(id, ch_txEsm, ch_rxEsm, ch_localElevatorStruct, ch_peerTxEnable)

	//go distributor.Distributor(ch_drv_buttons)

	go network.States(id, ch_localElevatorStruct, allElevators)

	//distributor kanskje?
	for {
		select {
		case a := <-allElevators:
			fmt.Printf("All elevator states:\n")
			for id, e := range a {
				fmt.Printf("  %s  :  %+v\n", id, e)
			}
			fmt.Printf("\n")
			//Videresend til assigner sånn at den kan regne ut
		}
	}

	ch_wait := make(chan bool)
	<-ch_wait
}

/* MAC Adress, sug en feit en
package main

import (
    "fmt"
    "log"
    "net"
)

func getMacAddr() ([]string, error) {
    ifas, err := net.Interfaces()
    if err != nil {
        return nil, err
    }
    var as []string
    for _, ifa := range ifas {
        a := ifa.HardwareAddr.String()
        if a != "" {
            as = append(as, a)
        }
    }
    return as, nil
}

func main() {
    as, err := getMacAddr()
    if err != nil {
        log.Fatal(err)
    }
    for _, a := range as {
        fmt.Println(a)
    }
}*/
