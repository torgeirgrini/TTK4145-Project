package main

import (
	"Project/config"
	"Project/localElevator/elevio"
	"Project/localElevator/fsm"
	"Project/network"
)

func main() {

	elevio.Init("localhost:15657", config.NumFloors)

	drv_buttons := make(chan elevio.ButtonEvent)
	drv_floors := make(chan int)
	drv_obstr := make(chan bool)
	//drv_stop := make(chan bool)

	go elevio.PollButtons(drv_buttons)
	go elevio.PollFloorSensor(drv_floors)
	go elevio.PollObstructionSwitch(drv_obstr)

	//go elevio.PollStopButton(drv_stop)

	//run FSM:
	go fsm.RunElevator(drv_buttons, drv_floors, drv_obstr)

	go network.Network()

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
