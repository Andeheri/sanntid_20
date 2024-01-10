Exercise 1 - Theory questions
-----------------------------

### Concepts

What is the difference between *concurrency* and *parallelism*?
> Parallellism runs on multiple cores at the exact same time, concurrency means dealing with multiple processes on the same CPU (at seemingly same time).

What is the difference between a *race condition* and a *data race*? 
> 
A data race happens when two threads access the same mutable object without synchronization, while a race condition happens when the order of events affects the correctness of the program. A data race can cause a race condition, but not always. 
 
*Very* roughly - what does a *scheduler* do, and how does it do it?
> Switches between and selects which thread to run. Can be random, ordered (preemptive). Can also have priorities on threads.
Or it can be cooperative between threads.


### Engineering

Why would we use multiple threads? What kinds of problems do threads solve?
> Desire to do independent things at the "same" time. 

Some languages support "fibers" (sometimes called "green threads") or "coroutines"? What are they, and why would we rather use them over threads?
> Fibers divide a thread into smaller tasks. Structure?

Does creating concurrent programs make the programmer's life easier? Harder? Maybe both?
> Yes because things from outside world are easier to handle and no because introduces challenges for example race conditions. 

What do you think is best - *shared variables* or *message passing*?
> Depends.


