package main

import (
	"context"
	"flag"
	"fmt"
	"math/big"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"
)

type RangeFlags [][2]int

func (rf *RangeFlags) String() string {
	return ""
}

func (rf *RangeFlags) Set(value string) error {
	parts := strings.Split(value, ":")
	if len(parts) != 2 {
		return fmt.Errorf("invalid range format")
	}

	start, err := strconv.Atoi(parts[0])
	if err != nil {
		return err
	}

	end, err := strconv.Atoi(parts[1])
	if err != nil {
		return err
	}

	*rf = append(*rf, [2]int{start, end})

	return nil
}

func main() {
	fileName := flag.String("file", "output.txt", "Файл для записи")
	timeout := flag.Int("timeout", 10, "время выполнения в секундах")

	var ranges RangeFlags
	flag.Var(&ranges, "range", "диапазон номеров в формате начало:конец")

	flag.Parse()

	ctx, cancel := context.WithTimeout(context.Background(), time.Duration(*timeout)*time.Second)
	defer cancel()

	primeChan := make(chan int)

	doneChan := make(chan struct{})

	var wg sync.WaitGroup

	go writePrimesToFile(ctx, *fileName, primeChan, doneChan)

	for _, r := range ranges {
		wg.Add(1)
		go func(r [2]int) {
			defer wg.Done()
			findPrimesInRange(ctx, r[0], r[1], primeChan)
		}(r)
	}

	wg.Wait()

	close(primeChan)

	<-doneChan

}

func findPrimesInRange(ctx context.Context, start int, end int, primeCh chan<- int) {
	for num := start; num <= end; num++ {
		select {
		case <-ctx.Done():
			return
		default:
			if isPrime(num) {
				primeCh <- num
			}
		}
	}
}

func isPrime(n int) bool {
	return big.NewInt(int64(n)).ProbablyPrime(0)
}

func writePrimesToFile(ctx context.Context, fileName string, primeCh <-chan int, doneCh chan<- struct{}) {
	file, err := os.Create(fileName)
	if err != nil {
		fmt.Println("error creating file:", err)
		close(doneCh)
		return
	}
	defer file.Close()

	for {
		select {
		case <-ctx.Done():

			close(doneCh)
			return

		case prime, ok := <-primeCh:
			if !ok {
				close(doneCh)
				return
			}
			_, err := file.WriteString(fmt.Sprintf("%d\n", prime))
			if err != nil {
				fmt.Println("error while write number", err)
				close(doneCh)
				return
			}
		}
	}
}
