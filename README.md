# channelx
[![Go Report Card](https://goreportcard.com/badge/github.com/Ksloveyuan/channelx)](https://goreportcard.com/report/github.com/Ksloveyuan/channelx)
[![build](https://api.travis-ci.org/Ksloveyuan/channelx.svg?branch=master)](https://api.travis-ci.org/Ksloveyuan/channelx.svg?branch=master)
[![codecov](https://codecov.io/gh/Ksloveyuan/channelx/branch/master/graph/badge.svg)](https://codecov.io/gh/Ksloveyuan/channelx)
[![GoDoc](https://godoc.org/github.com/Ksloveyuan/channelx?status.svg)](https://godoc.org/github.com/Ksloveyuan/channelx)

Proivde useful tools by channel to increaes development efficiency.

## stream
### before
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

### after

```golang
var sum = 0

NewChannelStream(func(seedChan chan<- Item, quitChannel chan struct{}) {
    for i:=1; i<=100;i++{
        seedChan <- Item{Data:i}
    }
    close(seedChan) //don't forget to close it
}).Pipe(func(Item Item) Item {
    return Item{Data: Item.Data.(int) * 2}
}).Pipe(func(Item Item) Item {
    return Item{Data: Item.Data.(int) - 1}
}).Harvest(func(Item Item) {
    sum += Item.Data.(int)
})

fmt.Println(sum)
```
