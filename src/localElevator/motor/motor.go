package motor

import (
	"Project/config"
	"Project/localElevator/elevio"
	"time"
)

func Motor(
	ch_stuck chan<- bool,
	ch_setMotorDirn <-chan elevio.MotorDirection,
) {

	MotorWatchDog := time.NewTimer(time.Duration(config.DoorOpenDuration_s) * time.Second)
	MotorWatchDog.Stop()
	ch_motorWatchDog := MotorWatchDog.C

	for {
		select {
		case <-ch_motorWatchDog:
			ch_stuck <- true
		case dirn := <-ch_setMotorDirn:
			if dirn != elevio.MD_Stop {
				MotorWatchDog.Reset(time.Duration(config.MotorTimeOut_s) * time.Second)
			} else {
				MotorWatchDog.Stop()
			}
			elevio.SetMotorDirection(dirn)
		}
	}
}
