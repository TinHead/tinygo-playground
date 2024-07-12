package main

import (
	"errors"
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

func handleMotors(packet []byte) error {
	drive := initMotors()
	if len(packet) > 1 {
		switch packet[1] {
		case 102:
			// forward
			fmt.Println("Got fw command!")
			drive.mot1.Forward()
			drive.mot2.Forward()
		case 98:
			// backward
			drive.mot1.Backward()
			drive.mot2.Backward()
		case 115:
			//stop
			drive.mot1.Stop()
			drive.mot2.Stop()
		case 108:
			//left
		case 114:
			//right
		default:
			return errors.New("Wrong motor packet format:" + string(packet))

		}
	} else {
		return errors.New("Packet is missing command")
	}
	return nil
}

func handlePacket(packet []byte, uart machine.UART) error {
	// packet format [type,data] where type is packet type and data nil or int value
	switch packet[0] {
	case 109:
		// motor control - one or two values:
		// forward - f 1-100
		// backward - b 1-100
		// stop s
		// left - l 1-100
		// right - r 1-100
		// turn left a 1-100
		// turn right d 1-100
		err := handleMotors(packet)
		if err != nil {
			return err
		}
	case 101:
		// echo request no value
	case 98:
		// battery level no value
	case 99:
		// custom request?
	default:
		return errors.New("Don't know how to handle packet: " + string(packet[:]))
	}
	return nil
}

// handle serial commands
func handleComms(uart machine.UART) {
	fmt.Println("Starting comms ..")
	packet := []byte{}
	for {
		if uart.Buffered() > 0 {
			rcv, err := uart.ReadByte()
			if err != nil {
				fmt.Println("Error while receiving: ", err)
			}
			switch rcv {
			case 13:
				fmt.Println("End of packet, contains:", packet)
				/// handle the complete packet
				if packet != nil {
					err := handlePacket(packet, uart)
					if err != nil {
						fmt.Println(err)
					}
				}
				packet = nil
			default:
				fmt.Println("Got ", rcv)
				packet = append(packet, rcv)
				fmt.Println("Packet so far: ", packet)
			}
			// if rcv == 0 {
			// 	fmt.Println("nothing received")
			// } else {
			// 	fmt.Println("received: ", rcv)
			// }
		}

		time.Sleep(time.Millisecond * 1)
	}
}

func main() {
	fmt.Println("Ready")
	// test := initMotors()
	// test.mot1.Configure()
	// test.mot1.Forward()
	// test.mot2.Configure()
	// test.mot2.Forward()

	uart := initUart()
	go handleComms(uart)
	for {
		fmt.Print(".")
		// uart.Write([]byte("Hello!\r\n"))
		time.Sleep(time.Second * 1)
	}
}
