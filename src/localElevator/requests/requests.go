package requests

import (
	"Project/config"
	"Project/localElevator/elevio"
	"Project/types"
	"Project/utilities"
)

func requests_above(e types.Elevator) bool {
	for f := e.Floor + 1; f < config.NumFloors; f++ {
		for btn := 0; btn < config.NumButtons; btn++ {
			if e.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func requests_below(e types.Elevator) bool {
	for f := 0; f < e.Floor; f++ {
		for btn := 0; btn < config.NumButtons; btn++ {
			if e.Requests[f][btn] {
				return true
			}
		}
	}
	return false
}

func requests_here(e types.Elevator) bool {
	for btn := 0; btn < config.NumButtons; btn++ {
		if e.Requests[e.Floor][btn] {
			return true
		}
	}
	return false
}

func Requests_nextAction(e types.Elevator, btn elevio.ButtonType) types.Action {
	switch e.Dirn {
	case elevio.MD_Up:
		if requests_above(e) {
			return types.Action{Dirn: elevio.MD_Up, Behaviour: types.EB_Moving}
		} else if requests_here(e) {
			return types.Action{Dirn: elevio.MD_Down, Behaviour: types.EB_DoorOpen}
		} else if requests_below(e) {
			return types.Action{Dirn: elevio.MD_Down, Behaviour: types.EB_Moving}
		} else {
			return types.Action{Dirn: elevio.MD_Stop, Behaviour: types.EB_Idle}
		}
	case elevio.MD_Down:
		if requests_below(e) {
			return types.Action{Dirn: elevio.MD_Down, Behaviour: types.EB_Moving}
		} else if requests_here(e) {
			return types.Action{Dirn: elevio.MD_Up, Behaviour: types.EB_DoorOpen}
		} else if requests_above(e) {
			return types.Action{Dirn: elevio.MD_Up, Behaviour: types.EB_Moving}
		} else {
			return types.Action{Dirn: elevio.MD_Stop, Behaviour: types.EB_Idle}
		}
	case elevio.MD_Stop:
		if requests_here(e) {
			if btn == elevio.BT_HallUp {
				return types.Action{Dirn: elevio.MD_Up, Behaviour: types.EB_DoorOpen}
			} else if btn == elevio.BT_HallDown {
				return types.Action{Dirn: elevio.MD_Down, Behaviour: types.EB_DoorOpen}
			}
			return types.Action{Dirn: elevio.MD_Stop, Behaviour: types.EB_DoorOpen}
		} else if requests_above(e) {
			return types.Action{Dirn: elevio.MD_Up, Behaviour: types.EB_Moving}
		} else if requests_below(e) {
			return types.Action{Dirn: elevio.MD_Down, Behaviour: types.EB_Moving}
		} else {
			return types.Action{Dirn: elevio.MD_Stop, Behaviour: types.EB_Idle}
		}
	default:
		return types.Action{Dirn: elevio.MD_Stop, Behaviour: types.EB_Idle}
	}
}

func Requests_shouldStop(e types.Elevator) bool {
	switch e.Dirn {
	case elevio.MD_Down:
		return (e.Requests[e.Floor][elevio.BT_HallDown]) || (e.Requests[e.Floor][elevio.BT_Cab]) || (!requests_below(e))
	case elevio.MD_Up:
		return (e.Requests[e.Floor][elevio.BT_HallUp]) || (e.Requests[e.Floor][elevio.BT_Cab]) || (!requests_above(e))
	default:
		return true
	}
}

func Requests_shouldClearImmediately(e types.Elevator, btn_floor int, btn_type elevio.ButtonType) bool {
	switch e.ClearRequestVariant {
	case types.CV_All:
		return e.Floor == btn_floor
	case types.CV_InDirn:
		return e.Floor == btn_floor &&
			((e.Dirn == elevio.MD_Up && btn_type == elevio.BT_HallUp) ||
				(e.Dirn == elevio.MD_Down && btn_type == elevio.BT_HallDown) ||
				e.Dirn == elevio.MD_Stop || btn_type == elevio.BT_Cab ||
				(e.Dirn == elevio.MD_Up && e.Floor == config.NumFloors-1) ||
				(e.Dirn == elevio.MD_Down && e.Floor == 0))

	default:
		return false
	}
}

func Requests_clearAtCurrentFloor(elev types.Elevator) types.Elevator {
	e := utilities.DeepCopyElevatorStruct(elev)
	switch e.ClearRequestVariant {
	case types.CV_All:
		for btn := 0; btn < config.NumButtons; btn++ {
			e.Requests[e.Floor][btn] = false
		}
	case types.CV_InDirn:
		e.Requests[e.Floor][elevio.BT_Cab] = false
		switch e.Dirn {
		case elevio.MD_Up:
			if (!requests_above(e)) && (!e.Requests[e.Floor][elevio.BT_HallUp]) {
				e.Requests[e.Floor][elevio.BT_HallDown] = false
			}
			e.Requests[e.Floor][elevio.BT_HallUp] = false
		case elevio.MD_Down:
			if (!requests_below(e)) && (!e.Requests[e.Floor][elevio.BT_HallDown]) {
				e.Requests[e.Floor][elevio.BT_HallUp] = false
			}
			e.Requests[e.Floor][elevio.BT_HallDown] = false
		case elevio.MD_Stop:
			fallthrough
		default:
			e.Requests[e.Floor][elevio.BT_HallUp] = false
			e.Requests[e.Floor][elevio.BT_HallDown] = false
		}
	default:
		break
	}
	return e
}
