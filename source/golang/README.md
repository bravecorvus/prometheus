# Golang Prometheus Software

![Golang](https://68.media.tumblr.com/93601f0c11deeb9189b152096ffe8ec3/tumblr_ormg9e9Zr51s5a4bko1_1280.png)

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

## Set-Up (Pre-Built Binary)
If you install the bindary directly from [https://github.com/gilgameshskytrooper/Prometheus/releases](https://github.com/gilgameshskytrooper/Prometheus/releases), then the only thing you need to do is to fill in the correct email address stored in [public/json/email](public/json/email), so that Prometheus can send you an email notification if the IP on your Pi changes. (Mainly if you have a dynamically assigned IP address from your ISP). If you have a static IP, this part of the set up. Once the email where you want to receive notifications, you are good to go. Start `./main`.


## Set-Up (Built from source Binary)
If you cloned the repo, and built the binary from source, then you will need to replace the following code:
```
//Account from which Prometheus sends an email from.
from := "email@example.com"
pass := "password"
```

Replace the from/pass values with a valid email address from which Prometheus will send the notification about an IP change. (Note, this is only necessary if you built from source. A email is provided if you use the pre-built binary). Then, replace the email stored at [public/json/email](public/json/email) which is the email which Prometheus will send the IP change notification to.

## Initial Start
At program start, you will receive an automatic email notifying you of an IP change since this is a stored value, at [public/json/ip](public/json/ip). Once your IP gets stored, then it will only notify you when it changes.

## User Interface
(***See general [User Interface page](https://github.com/gilgameshskytrooper/Prometheus/wiki/User-Interface-Tutorial)***)
