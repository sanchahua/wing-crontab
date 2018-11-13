package main

import (
	"sync/atomic"
	"unsafe"
	"fmt"
	"os"
	"sync"
)

type ShareMemory struct {
	read unsafe.Pointer
	write unsafe.Pointer
}

func (sh *ShareMemory) Write(data interface{}) {
	atomic.StorePointer(&sh.write, unsafe.Pointer(&data))
	atomic.SwapPointer(&sh.read, sh.write)
}

func (sh *ShareMemory) Read() interface{}  {
	vv := (*interface{})(atomic.LoadPointer(&sh.read))
	if vv == nil {
		return nil
	}
	return *vv
}

type Test struct {
	Value int
}

func main() {
	wg := new(sync.WaitGroup)
	mem := &ShareMemory{}
	wg.Add(1000)
	for i := 0; i< 1000; i++ {
		go func() {
			defer wg.Done()
			v := &Test{1}

			mem.Write(v)

			d := mem.Read()
			dd, ok := d.(*Test)
			if ok {
				fmt.Fprintf(os.Stderr, "%+v\r\n", *dd)
			} else {
				fmt.Println("error")
			}

			mem.Write(nil)
			d = mem.Read()
			dd, ok = d.(*Test)
			fmt.Fprintf(os.Stderr, "dd=%+v, %+v\r\n", dd, nil == dd)

			if ok {
				fmt.Fprintf(os.Stderr, "%+v\r\n", *dd)
			} else {
				fmt.Println("error")
			}

			mem.Write(1)
			d2 := mem.Read()
			dd2, ok := d2.(int)
			if ok {
				fmt.Fprintf(os.Stderr, "%+v\r\n", dd2)
			} else {
				fmt.Println("error\r\n")
			}
		}()
	}
	wg.Wait()
}