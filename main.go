package main

import (
	"fmt"
	"sync"
	"time"
)

var (
	wg  = sync.WaitGroup{}
	mut = sync.RWMutex{}

	seats       []int
	endWorkChan = make(chan bool)
	endWork     = false
)

type wakeUp struct {
	wakeBell bool
	customer int
}

const barberNum = 2

var wakeBellChan = make(chan wakeUp)

func main() {
	wg.Add(1)
	go incomingCustomer()
	wg.Add(1)
	go controlBarberWork()
	// if there is only one Barber then run only one goroutine
	// for the Barber which is startService() and in this
	// scenario, you have to set barberNum = 1
	//
	// wg.Add(1)
	// go startService()
	//
	// If there are multiple Barber then you have to run one
	// goroutine(startService()) for each Barber.
	// Suppose there are 5 Barbers then you have to run 5
	// instances of goroutine startService() which is
	// 5 startService() goroutines.in this scenario,
	// you have to set barberNum = 5
	for i := 1; i <= barberNum; i++ {
		wg.Add(1)
		go startService(i)
	}

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
		mut.RLock() // mutex read lock
		seatLength := len(seats)
		mut.RUnlock() // mutex read unlock
		if seatLength < 5 {
			// Let's provide some break to the Barber by
			// not sending customers for a brief period of time
			if breakForBarber > 10 && noBreaksProvided < 6 {
				mut.Lock() // mutex write lock
				seats = append(seats, 2000)
				mut.Unlock() // mutex write unlock
				noBreaksProvided++
				customerNumber-- // Since, we are not adding actual customers for 5 iterations
			} else {
				// Now, send some new customers to the Barber
				mut.Lock() // mutex write lock
				seats = append(seats, customerNumber)
				mut.Unlock() // mutex write unlock
			}
		} else {
			fmt.Printf("customer%d has left as seats are full\n", customerNumber)
		}
	}
	mut.Lock() // mutex write lock
	endWork = true
	mut.Unlock() // mutex write unlock
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
		mut.RLock() // mutex read lock
		seatLength := len(seats)
		mut.RUnlock() // mutex read lock
		if seatLength > 0 {
			mut.Lock() // mutex write lock
			val = seats[0]
			seats = seats[1:]
			mut.Unlock() // mutex write unlock
		}

		mut.RLock() // mutex read lock
		checkSeatLength := len(seats)
		endWorkVal := endWork
		mut.RUnlock() // mutex read lock
		if endWorkVal && checkSeatLength == 0 {
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

			// if there are more than one Barber then we
			// have to send signal to the endWorkChan that
			// many times so that every Barbers gets the
			// signal to end work and go home
			for i := 0; i < barberNum; i++ {
				endWorkChan <- true
			}

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

// func startService() {
// 	defer wg.Done()
// 	for {
// 		time.Sleep(1 * time.Second)
// 		select {
// 		// wake up the Barber
// 		case customerNumber := <-wakeBellChan:
// 			if customerNumber.customer == 2000 {
// 				fmt.Println("Barber is sleeping now")
// 			} else {
// 				fmt.Printf("Barber has started providing its service to customer%d\n", customerNumber.customer)
// 				time.Sleep(2 * time.Second)
// 			}
// 		case <-endWorkChan:
// 			fmt.Println("Shop is closed now. Barber has left for Home. ")
// 			return
// 		default:
// 			fmt.Println("Barber is sleeping now")
// 		}
// 	}

// }

func startService(barbers int) {
	defer wg.Done()
	for {
		time.Sleep(1 * time.Second)
		select {
		// wake up the Barber
		case customerNumber := <-wakeBellChan:
			if customerNumber.customer == 2000 {
				fmt.Printf("Barber%d is sleeping now\n", barbers)
			} else {
				fmt.Printf("Barber%d has started providing its service to customer%d\n", barbers, customerNumber.customer)
				time.Sleep(2 * time.Second)
			}
		case <-endWorkChan:
			fmt.Printf("Shop is closed now and no customer is waiting. So, Barber%d has left for Home. \n", barbers)
			return
		default:
			fmt.Printf("Barber%d is sleeping now\n", barbers)
		}
	}

}

// run "go run -race main.go" command to check the existance of Race Condition
// in this application.
// To mitigate Race condition Problem, use Mutex.
// RWMutex is faster than the Mutex that's why RWMutex is used in this application.
// Since, Mutex blocks for both the read and write whereas RWMutex only blocks
// during write that's why RWMutex is faster(more performant) than Mutex.
