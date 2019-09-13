# channelx

before
```golang
	var multipleChan = make(chan int, 4)
	var minusChan = make(chan int, 4)
	var harvestChan = make(chan int, 4)

	defer close(multipleChan)
    defer close(minusChan)
    defer close(harvestChan)

	go func() {
		for i:=1;i<=100;i++{
			multipleChan <- i
		}
	}()

	for i:=0; i<4; i++{
		go func() {
			for data := range multipleChan {
				minusChan <- data * 2
				time.Sleep(10* time.Millisecond)
			}
		}()

		go func() {
			for data := range minusChan {
				harvestChan <- data - 1
				time.Sleep(10* time.Millisecond)
			}
		}()
	}

	var sum = 0
	var index = 0
	for data := range harvestChan{
		sum += data
		index++
		if index == 100{
			break
		}
	}

	fmt.Println(sum)
```

after

```golang
	var sum = 0

	NewChannelStream(func(seedChan chan<- Result, quitChannel chan struct{}) {
		for i:=1; i<=100;i++{
			seedChan <- Result{Data:i}
		}
		close(seedChan) //don't forget to close it
	}).Pipe(func(result Result) Result {
		return Result{Data: result.Data.(int) * 2}
	}).Pipe(func(result Result) Result {
		return Result{Data: result.Data.(int) - 1}
	}).Harvest(func(result Result) {
		sum += result.Data.(int)
	})

	fmt.Println(sum)
}
```