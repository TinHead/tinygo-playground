package main

import (
	"fmt"
	"machine"
	"time"

	"tinygo.org/x/drivers/l9110x"
)

type msg struct {
	msg_type []byte
	msg_data []byte
}

type motors struct {
	mot1 l9110x.Device
	mot2 l9110x.Device
	// mot3 l9110x.PWMDevice
	// mot4 l9110x.PWMDevice
}

var freq uint64 = 1e9 / 1000

func initMotors() motors {
	m1 := machine.GP16
	m2 := machine.GP17
	m3 := machine.GP18
	m4 := machine.GP19

	m1.Configure(machine.PinConfig{Mode: machine.PinOutput})
	m2.Configure(machine.PinConfig{Mode: machine.PinOutput})
	m3.Configure(machine.PinConfig{Mode: machine.PinOutput})
	m4.Configure(machine.PinConfig{Mode: machine.PinOutput})

	// err := machine.PWM6.Configure(machine.PWMConfig{Period: freq})
	// if err != nil {
	// 	println(err)
	// }
	// err = machine.PWM7.Configure(machine.PWMConfig{})
	// if err != nil {
	// 	println(err)
	// }
	// ch1, err := machine.PWM6.Channel(m1)
	// if err != nil {
	// 	println(err)
	// }
	// ch2, err := machine.PWM6.Channel(m2)
	// if err != nil {
	// 	println(err)
	// }
	// ch3, err := machine.PWM7.Channel(m3)
	// if err != nil {
	// 	println(err)
	// }
	// ch4, err := machine.PWM7.Channel(m4)
	// if err != nil {
	// 	println(err)
	// }

	return motors{mot1: l9110x.New(m1, m2), mot2: l9110x.New(m3, m4)}

}

func initUart() machine.UART {
	uart0 := machine.UART0
	uart0.Configure(machine.UARTConfig{TX: machine.GP0, RX: machine.GP1})
	uart0.SetBaudRate(115200)
	return *uart0
}

// handle serial commands
func handleComms(uart machine.UART) {
	fmt.Println("Starting comms ..")
	for {
		if uart.Buffered() > 0 {
			rcv, err := uart.ReadByte()
			if err != nil {
				fmt.Println("Error while receiving: ", err)
			}
			if rcv == 0 {
				fmt.Println("nothing received")
			} else {
				fmt.Println("received: ", rcv)
			}
		}

		time.Sleep(time.Millisecond * 1)
	}
}

func main() {
	fmt.Println("Ready")
	test := initMotors()
	test.mot1.Configure()
	test.mot1.Forward()
	test.mot2.Configure()
	test.mot2.Forward()

	uart := initUart()
	go handleComms(uart)
	for {
		fmt.Print(".")
		// uart.Write([]byte("Hello!\r\n"))
		time.Sleep(time.Second * 1)
	}
}
