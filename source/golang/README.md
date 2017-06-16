# Golang Prometheus Software

### Author: Andrew Lee

## Function
[main](main.go) functions as both the server, which hosts the web-based user interface as well as the hardware controlling mechanisms.

## Build From Source
In order to be able to use the interface, the user will have to install the necessary dependencies for this to run.

First, this program uses VLC to play music (by spawning a shell process).

```
>sudo apt-get install vlc
```

Next, you will need the golang build tools.

```
>sudo apt-get install golang
```

Finally, you need to install the third party libraries.

The following is a cron library that allows the Prometheus to check the current time against the user supplied alarm times, once a minute.
```
>go get github.com/robfig/cron
```

The following is a library that allows golang to easily interact with the GPIO pins on the Pi.
```
>go get github.com/stianeikeland/go-rpio
```

Then, you just have to build the program. Then connect the relevant wires and you are good to go (More of that on our [Wiki](https://github.com/gilgameshskytrooper/Prometheus/wiki/Hardware-Set-Up))
```
>cd Prometheus/source/golang/
>go build main.go
```
