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

	MotorWatchDog := time.NewTimer(time.Duration(config.MotorTimeOut_s) * time.Second)
	MotorWatchDog.Stop()
	ch_motorWatchDog := MotorWatchDog.C

	for {
		select {
		case <-ch_motorWatchDog:
			ch_stuck <- true
		case motorDir := <-ch_setMotorDirn:
			switch motorDir {
			case elevio.MD_Up:
				MotorWatchDog.Reset(time.Duration(config.MotorTimeOut_s) * time.Second)
			case elevio.MD_Down:
				MotorWatchDog.Reset(time.Duration(config.MotorTimeOut_s) * time.Second)
			case elevio.MD_Stop:
				MotorWatchDog.Stop()
				ch_stuck <- false
			}
			elevio.SetMotorDirection(motorDir)
		}
	}
}
