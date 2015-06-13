Idle
=====

Cross-platform idle time detection in Go (Golang).

<img src="https://lh6.googleusercontent.com/-sm9TUtep2xs/T3R7ZCDrJVI/AAAAAAAAAKQ/jaSnMOyRJGw/w856-h1228-no/2_b%2Bdata%2Bnieznana.jpg" height="400"><br>

Read and run [idle.go](idle) command for an example of the API usage:

```
> go run cmd/idle.go
2015/06/13 22:55:43 Idle for 1 seconds.
2015/06/13 22:55:44 Idle for 2 seconds.
2015/06/13 22:55:45 Idle for 3 seconds.
2015/06/13 22:55:46 Idle for 4 seconds.
2015/06/13 22:55:47 Idle for 5 seconds.

2015/06/13 22:55:50 Idle for 1 seconds.
2015/06/13 22:55:51 Idle for 2 seconds.
2015/06/13 22:55:52 Idle for 3 seconds.
2015/06/13 22:55:53 Idle for 4 seconds.
2015/06/13 22:55:54 Idle for 5 seconds.
```

To do
------

* Docs.
* Windows support.
* Better OSX: Figure out how to collect these values using ```c.go```. (The way I'm currently doing it is brittle as fuck).
* Test Linux.

License
-------

MIT
