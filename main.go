package main

import (
	"fmt"
	"sync"
	"time"
)

var (
	wg = sync.WaitGroup{}

	seats       []int
	endWorkChan = make(chan bool)
	endWork     = false
)

type wakeUp struct {
	wakeBell bool
	customer int
}

var wakeBellChan = make(chan wakeUp)

func main() {
	wg.Add(1)
	go incomingCustomer()
	wg.Add(1)
	go controlBarberWork()
	wg.Add(1)
	go startService()

	wg.Wait()
}

func incomingCustomer() {
	defer wg.Done()
	// Assuming the Barber Shop is opened for 1 minute per day and
	// every customer is arriving at an interval of 1 Second
	duration := 60 // 1 minute duration
	customerNumber := 0
	breakForBarber := 0
	noBreaksProvided := 0
	for i := 0; i < duration; i++ {
		time.Sleep(1 * time.Second)
		customerNumber++
		breakForBarber++
		if len(seats) < 5 {
			// Let's provide some break to the Barber by
			// not sending customers for a brief period of time
			if breakForBarber > 10 && noBreaksProvided < 6 {
				seats = append(seats, 2000)
				noBreaksProvided++
				customerNumber-- // Since, we are not adding actual customers for 5 iterations
			} else {
				// Now, send some new customers to the Barber
				seats = append(seats, customerNumber)
			}
		} else {
			fmt.Printf("customer%d has left as seats are full\n", customerNumber)
		}
	}
	endWork = true
}

func controlBarberWork() {
	defer wg.Done()
	// I assume that the Barber takes 2 seconds of time to provide
	// his/her service to the customer. And there is only ! Barber
	// to provide service to the customers.
	//
	// When not working barber is sleeping
	for {
		val := 0
		if len(seats) > 0 {
			val = seats[0]
			seats = seats[1:]
		}

		if endWork && len(seats) == 0 {
			// when the time to close the shop arrives it is time
			// to close the shop but there are some waiting customers.
			// So we didn't shut down the shop yet and patiently
			// served all the other waiting customers. And those
			// waiting customers have already been taken care of
			// in the previous process except the last waiting
			// customer. And we are sending this last waiting
			// customer(through the wakeBellChan) to the barber
			// along with the endwork signal in this section.
			newWakeBell := wakeUp{
				wakeBell: true,
				customer: val,
			}
			wakeBellChan <- newWakeBell

			endWorkChan <- true
			return
		} else {
			if val == 0 {
				continue // val == 0 means no customer is added to seats variable yet
			}
			newWakeBell := wakeUp{
				wakeBell: true,
				customer: val,
			}
			wakeBellChan <- newWakeBell

		}

	}

}

func startService() {
	defer wg.Done()
	for {
		time.Sleep(1 * time.Second)
		select {
		// wake up the Barber
		case customerNumber := <-wakeBellChan:
			if customerNumber.customer == 2000 {
				fmt.Println("Barber is sleeping now")
			} else {
				fmt.Printf("Barber has started providing its service to customer%d\n", customerNumber.customer)
				time.Sleep(2 * time.Second)
			}
		case <-endWorkChan:
			fmt.Println("Shop is closed now. Barber has left for Home. ")
			return
		default:
			fmt.Println("Barber is sleeping now")
		}
	}

}
