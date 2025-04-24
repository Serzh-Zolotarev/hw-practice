package main

import (
	"bufio"
	"fmt"
	"os"
	"strconv"
	"sync"
	"time"
)

const (
	bufferSize          int           = 5
	drainBufferDuration time.Duration = 5 * time.Second

	errMsg      = "Неверный символ, пожалуйста, введите число"
	exitCommand = "exit"
)

type RingBuffer struct {
	intArr  []int
	pointer int
	size    int
	m       sync.Mutex
}

func (r *RingBuffer) Push(el int) {
	r.m.Lock()
	defer r.m.Unlock()
	if r.pointer == r.size-1 {
		for i := 1; i <= r.size-1; i++ {
			r.intArr[i-1] = r.intArr[i]
		}
		r.intArr[r.pointer] = el
	} else {
		r.pointer++
		r.intArr[r.pointer] = el
	}
}

func (r *RingBuffer) Get() []int {
	if r.pointer < 0 {
		return nil
	}
	r.m.Lock()
	defer r.m.Unlock()
	var output []int = r.intArr[:r.pointer+1]
	r.pointer = -1
	return output
}

func main() {
	printer(bufferStage(buffNonMult3Stage(onlyUnsignedStage(scanConsole()))))
}

// Фильтрация отрицательных чисел и нулей
func onlyUnsignedStage(done <-chan int, input <-chan int) (<-chan int, <-chan int) {
	fmt.Println("filter only unsigned stage started")

	unsignedStream := make(chan int)
	go func() {
		defer close(unsignedStream)
		for {
			select {
			case data := <-input:
				if data > 0 {
					select {
					case unsignedStream <- data:
					case <-done:
						return
					}
				}
			case <-done:
				return
			}
		}
	}()

	fmt.Println("filter only unsigned stage ended")

	return done, unsignedStream
}

// Фильтрация чисел не кратных 3
func buffNonMult3Stage(done <-chan int, input <-chan int) (<-chan int, <-chan int) {
	fmt.Println("filter non mult 3 stage started")

	mult3Stream := make(chan int)
	go func() {
		defer close(mult3Stream)
		for {
			select {
			case data := <-input:
				if data != 0 && data%3 == 0 {
					select {
					case mult3Stream <- data:
					case <-done:
						return
					}
				}
			case <-done:
				return
			}
		}
	}()

	fmt.Println("filter non mult 3 stage ended")

	return done, mult3Stream
}

// Буферизация
func bufferStage(done <-chan int, input <-chan int) (<-chan int, <-chan int) {
	fmt.Println("Buffer stage started")

	bufferChan := make(chan int)
	buffer := RingBuffer{make([]int, bufferSize), -1, bufferSize, sync.Mutex{}}

	go func() {
		for {
			select {
			case data := <-input:
				buffer.Push(data)
			case <-done:
				return
			}
		}
	}()

	go func() {
		for {
			select {
			case <-time.After(drainBufferDuration):
				data := buffer.Get()
				if data != nil {
					for _, i := range data {
						select {
						case bufferChan <- i:
						case <-done:
							return
						}
					}
				}
			case <-done:
				return
			}
		}
	}()

	fmt.Println("Buffer stage ended")

	return done, bufferChan
}

// Вывод в консоль
func printer(done <-chan int, c <-chan int) {
	for {
		select {
		case data := <-c:
			fmt.Printf("Подходящее число: %d\n", data)
		case <-done:
			return
		}
	}
}

func scanConsole() (<-chan int, <-chan int) {
	fmt.Println("scan console stage started")

	c := make(chan int)
	done := make(chan int)
	go func() {
		defer close(done)
		scanner := bufio.NewScanner(os.Stdin)
		var data string
		for {
			scanner.Scan()
			data = scanner.Text()
			if data == exitCommand {
				return
			}

			i, err := strconv.Atoi(data)
			if err != nil {
				fmt.Println(errMsg)
				continue
			}
			c <- i
		}
	}()

	fmt.Println("scan console stage ended")

	return done, c
}
