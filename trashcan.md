

func PeriodicTransmit(
	ch_TxNewElevatorMessage chan<- MessageStruct,
	ch_newLocalState <-chan elevator.Elevator) {

	for {
		select {
		case LocalState := <-ch_newLocalState:
			_ = LocalState
		case <-time.After(config.TransmitInterval * time.Millisecond):
		}
		//msgstruct := createMsgStruct()
		//ch_TxNewElevatorMessage <- msgstruct
	}
	//informasjon om oss selv

	//lage en messagestruct

	//send på tc
	//sleep(some duration)
}





type ElevatorMessage struct {
	id        string
	Floor     int
	Dirn      elevio.MotorDirection
	Behaviour elevator.ElevatorBehaviour
	Available bool
	CabCalls  []bool
}





type MessageStruct struct {
	ThisElevator ElevatorMessage
	HallCalls    [config.NumFloors][config.NumButtons - 1]HallCall
}







func createMsgStruct(id string, hallCalls [config.NumFloors][config.NumButtons - 1]HallCall, localElev elevator.Elevator) MessageStruct {

	var localCabCalls []bool
	for i := 0; i < config.NumFloors; i++ {
		localCabCalls[i] = localElev.Requests[i][elevio.BT_Cab]
	}

	LocalElevMsg := ElevatorMessage{
		Floor:     localElev.Floor,
		Dirn:      localElev.Dirn,
		Behaviour: localElev.Behaviour,
		Available: true,
		CabCalls:  localCabCalls}

	return MessageStruct{
		ThisElevator: LocalElevMsg,
		HallCalls:    hallCalls}
}