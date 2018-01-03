# Golang Prometheus Software

![Golang](https://68.media.tumblr.com/93601f0c11deeb9189b152096ffe8ec3/tumblr_ormg9e9Zr51s5a4bko1_1280.png)

### Author: Andrew Lee

If you want documentation, check out the [godocs](https://godoc.org/github.com/gilgameshskytrooper/prometheus) for the project.

## Functionality

- [prometheus](https://godoc.org/github.com/gilgameshskytrooper/prometheus/prometheus) functions as both the server, which hosts the web-based user interface as well as the hardware controlling mechanisms
- [utils](https://godoc.org/github.com/gilgameshskytrooper/prometheus/utils) is the package that contains all the utility functions used by Prometheus
- [gpio](https://godoc.org/github.com/gilgameshskytrooper/prometheus/gpio) contains the specific functions used to execute GPIO.
- [structs](https://godoc.org/github.com/gilgameshskytrooper/prometheus/structs) contains the structs used by Prometheus to store alarms and to unmarshal JSON files
- [nixie](https://godoc.org/github.com/gilgameshskytrooper/prometheus/nixie) contains the code used to interact with the [Arduino nixie clock](https://gra-afch.com/product-category/shield-nixie-clock-for-arduino/). Specifically, the code implemented in [prometheus main package](https://github.com/gilgameshskytrooper/prometheus/blob/master/prometheus.go) sends the current time (using the Go time library) to the Arduino via serial USB.
- [public](public/) contains all the static assets such as index.html, css, and javascript. The front-end functionality heavily utilizes [Vue.js](https://vuejs.org/) and I highly recommend it to anyone who is interested in a front-end framework.

## Software Installation

### Install Released Binary

The easiest way to install Prometheus is by downloading the zip of the executable and static files, unzipping them, and then putting them in `rc.local`

First, you need VLC which prometheus uses to play sound files.

```
sudo apt install vlc-nox
```

However, since you are using the command `cvlc` from `rc.local`, you will also need to enable root to be able to execute it (since by default, VLC does not allow the root use to start the app)
```
sudo apt install bless
sudo bless $(which cvlc)
sed -i 's/geteuid/getppid/' $(which cvlc)
```


Then grab the latest executable.

```
wget https://github.com/gilgameshskytrooper/prometheus/releases/download/v2.2.0/prometheus.v2.2.0.zip
unzip prometheus.v2.2.0.zip
rm prometheus.v2.2.0.zip
```

***the above link should be accurate, but check the [releases page](https://github.com/gilgameshskytrooper/prometheus/releases) to ensure that you are getting the most recent version***

Then add the path to the executable to `rc.local` to run prometheus every time you reboot.

***This is an example configuration, it may be different depending on the location you downloaded the program.***

```
/home/pi/prometheus/prometheus &
```

The `&` at the end is to ensure that bash runs prometheus, then sends it to a background in order to finish other boot processes.

### Build From Source
In order to be able to use the interface, the user will have to install the necessary dependencies for this to run.

First, this program uses VLC to play music (by spawning a shell process). Install via `apt`

```
sudo apt install vlc-nox
```

Next, you will need the golang build tools.

```
sudo apt-get install golang
```

Then you have to add `GOPATH` to your environment variables. In `.bashrc`, this will be accomplished by adding the following lines:

```
export GOPATH=~/go
export PATH=$PATH:$(go env GOPATH)/bin
```

and then running

```
source ~/.bashrc
```

Finally, you need to install the third party libraries.

The following is a cron library that allows the Prometheus to check the current time against the user supplied alarm times, once a minute.

```
go get github.com/robfig/cron
```

The following is a library that allows golang to easily interact with the GPIO pins on the Pi.

```
go get github.com/stianeikeland/go-rpio
```

Then get Prometheus source. and build it
```
go get github.com/gilgameshskytrooper/prometheus
cd $GOPATH/src/github.com/gilgameshskytrooper/prometheus
go build
```


To start the program at boot, you need add the following line to `/etc/rc.local`
```
cd $GOPATH/src/github.com/gilgameshskytrooper/prometheus/prometheus &
```

Then you will need to ensure root can execute the `cvlc` command (since by default, VLC does not allow the root use to start the app).
```
sudo apt install bless
sudo bless $(which cvlc)
sed -i 's/geteuid/getppid/' $(which cvlc)
```


## Sound
To get good sound, you will need to output the sound through a external sound card. The built-in TRS connector are terrible in sound. However, if this is sufficient for your needs, you can skip this section.

You need to set up `~/.asoundrc` to read your card.

For USB, it might look like the following.

```
vi ~/.asoundrc

pcm.!default {
    type hw
    card 1
}

ctl.!default {
    type hw
    card 1
}
```

For a custom sound card, it might look like the following:

```
pcm.!default {
	type hw card 0
}
ctl.!default {
	type hw card 0
}
```

## [Hardware Installation](Quickstart.md#hardware-stuff)

Note, if you opt to use a custom card, just change the setting in the bottom right hand corner of the front-end user interface.

## Initial Start
At program start, you will receive an automatic email notifying you of an IP change since this is a stored value, at [public/json/ip](public/json/ip). Once your IP gets stored, then it will only notify you when it changes. Change whether or not you want Prometheus to send you emails regarding changes to your IP in the front-end interface as well as the email you want to receive notifications on.

## User Interface
See general [user interface page](https://github.com/gilgameshskytrooper/Prometheus/wiki/User-Interface-Tutorial)
