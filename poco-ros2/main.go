package main

import (
	"fmt"
	"machine"
	"tinygo.org/x/drivers/l9110x"
)

type motors struct {
	mot1 l9110x.PWMDevice
	mot2 l9110x.PWMDevice
	// mot3 l9110x.PWMDevice
	// mot4 l9110x.PWMDevice
}

func initMotors() motors {
	m1 := machine.GP10
	m2 := machine.GP11
	m3 := machine.GP12
	m4 := machine.GP13

	m1.Configure(machine.PinConfig{Mode: machine.PinOutput})
	m2.Configure(machine.PinConfig{Mode: machine.PinOutput})
	m3.Configure(machine.PinConfig{Mode: machine.PinOutput})
	m4.Configure(machine.PinConfig{Mode: machine.PinOutput})

	err := machine.PWM0.Configure(machine.PWMConfig{})
	if err != nil {
		println(err)
	}
	err = machine.PWM1.Configure(machine.PWMConfig{})
	if err != nil {
		println(err)
	}
	ch1, err := machine.PWM0.Channel(m1)
	if err != nil {
		println(err)
	}
	ch2, err := machine.PWM0.Channel(m2)
	if err != nil {
		println(err)
	}
	ch3, err := machine.PWM1.Channel(m3)
	if err != nil {
		println(err)
	}
	ch4, err := machine.PWM1.Channel(m4)
	if err != nil {
		println(err)
	}

	return motors{mot1: l9110x.NewWithSpeed(ch1, ch2, machine.PWM0), mot2: l9110x.NewWithSpeed(ch3, ch4, machine.PWM1)}

}

func main() {
	test := initMotors()
	test.mot1.Configure()
	test.mot2.Configure()
	fmt.Println("hello poco")
}
