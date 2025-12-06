package util

import (
	"github.com/bwmarrin/snowflake"
	"os"
)

type AutoInc struct {
	start, step int
	queue       chan int
	running     bool
}

func New(start, step int) (ai *AutoInc) {
	ai = &AutoInc{
		start:   start,
		step:    step,
		running: true,
		queue:   make(chan int, 4),
	}
	go ai.process()
	return
}

func (ai *AutoInc) process() {
	defer func() { recover() }()
	for i := ai.start; ai.running; i = i + ai.step {
		ai.queue <- i
	}
}

func (ai *AutoInc) Id() int {
	return <-ai.queue
}

func (ai *AutoInc) Close() {
	ai.running = false
	close(ai.queue)
}

var IdWorker *snowflake.Node

func Setup() {
	snowflake.NodeBits = 2
	snowflake.StepBits = 5
	n, err := snowflake.NewNode(1)
	if err != nil {
		println(err)
		os.Exit(1)
	}
	IdWorker = n
}

//func main() {
//	var ai *AutoInc
//	ai = New(100000,1)
//
//	for{
//		a :=ai.Id()
//		time.Sleep(5*time.Second)
//		b :=strconv.Itoa(a)
//		fmt.Println(b,reflect.TypeOf(b))
//	}
//
//}
