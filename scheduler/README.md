# scheduler
This scheduler collects objects to queue according it's deadline time.
Checking the first one deadline element in this queue, you can understand is there other deadlines in this queue occurs or not.

## Usage
#### 1. Create scheduler:
```
scheduler := NewScheduler()
```
#### 2. Insert some elements with it deadline times:
```
now := time.Now()
scheduler.RegisterNewTimer(now.Add(2*time.Second), object2)
scheduler.RegisterNewTimer(now.Add(3*time.Second), object3)
scheduler.RegisterNewTimer(now.Add(3*time.Second), object3)
scheduler.RegisterNewTimer(now.Add(1*time.Second), object1)
```

#### 3. Check the first one deadline.
If it reached, method <b>scheduler.TakeFirstOutdated()</b> returns object, otherwise it returns nil.
```
object := scheduler.TakeFirstOutdatedOrNil()
if object != nil {
  // make some work with object
}
```
#### 4. You can cancel particular object from queue
Several same objects in queue supported
```
scheduler := scheduler.NewScheduler()
scheduler.RegisterNewTimer(time.Now(), 1)
scheduler.RegisterNewTimer(time.Now(), 1)
scheduler.RegisterNewTimer(time.Now(), 2)
scheduler.RegisterNewTimer(time.Now(), 1)
scheduler.RegisterNewTimer(time.Now(), 2)
scheduler.RegisterNewTimer(time.Now(), 2)
scheduler.CancelTimerFor(1)
scheduler.CancelTimerFor(2)
```



## Limitations and specific
Currently insertion of new deadline works for <b>O(n)</b> time (n - is the average size of deadlines queue)<br>
For large schedulers which have more than 1000 timers, it would be nice to use more optimized insertion algorithm with <b>O(log(n))</b> time.
