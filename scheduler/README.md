# scheduler
This scheduler collects objects to one queue according it's deadline time.
Checking the first one deadline element in this queue, you can understand is there deadline occurs or not.

## Usage
#### 1. Create scheduler:
```
scheduler := NewScheduler()
```
#### 2. Insert some elements with it deadline times:
```
scheduler.RegisterNewTimer(time.Now().Add(2*time.Second), object2)
scheduler.RegisterNewTimer(time.Now().Add(3*time.Second), object3)
scheduler.RegisterNewTimer(time.Now().Add(3*time.Second), object3)
scheduler.RegisterNewTimer(time.Now().Add(1*time.Second), object1)
```

#### 3. Check the first one deadline.
If it reached, method <b>scheduler.TakeFirstOutdated()</b> returns object, otherwise it returns nil.
```
object := scheduler.TakeFirstOutdated()
if object != nil {
  // make some work with object
}
```

## Limitations and specific
1. It is not thread safe (if you use <b>RegisterNewTimer(...) TakeFirstOutdated()</b> methods from different goroutines race condition could occur. Use channel to single goroutine or mutexes to manage it).
2. Currently insertion of new deadline works for <b>O(n)</b> time (n - is the avarage size of deadlines queue)<br>
   For large schedulers which have more than 1000 timers, it would be nice to use more optimized insertion algorith with <b>O(log(n))</b> time.
