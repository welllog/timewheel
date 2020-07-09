# timewheel
The go language implementation of the time wheel

## Usage

##### init timewheel
```go
tw := NewTimeWheel(100 * time.Millisecond, 50)
tw.Start()
defer tw.Stop()
```

##### adding one-time tasks 
```go
tw.AddOnce(time.Second, func() {
    fmt.Println("once")
})
```

##### adding tasks that run multiple times
```go
var i int
tw.AddWithTimes(time.Second, 5, func() {
    i++
    fmt.Println(i)
})
```

##### adding tasks that run all the time
```go
taskId := tw.AddCron(time.Second, func() {
    i++
})
```

##### remove task
```go
// warning!!! No guarantee that removing a task multiple times will return the correct value
// so it is better to remove it once if you need use the return value.
// remove task and return Whether the task has been run
tw.RemoveAndHasRun(taskId)
// only remove task
tw.Remove(taskId)
```

##### timer
```go
// the usage is similar to time.Timer
timer := tw.NewTimer(time.Second)
<-timer.C
timer.Reset(500 * time.Millisecond)
timer.Stop()
```

##### ticker
```go
// the usage is similar to time.Ticker
ticker := tw.NewTicker(500 * time.Millisecond)
var incr int
for {
    <-ticker.C
    
    incr++
    if incr == 20 {
        ticker.Stop()
        break
    }
}
```

##### other usage
```go
// Reference time.After
<-tw.After(time.Second)
// Reference time.AfterFunc
tw.AfterFunc(500 * time.Millisecond, func() {
    fmt.Println(1)
})
// Reference time.Sleep
tw.Sleep(time.Second)
```

#### use time wheel with pool
```go
twp := NewTimeWheelPool(5, 100 * time.Millisecond, 50)
twp.Start()
defer twp.Stop()

<-twp.Get().After(time.Second)
```

