# Prometheus Smart Clock

### By Andrew Lee

![Prometheus Clock Cover](assets/newcover.jpg)



### [I JUST WANT TO GET THIS UP AND RUNNING](https://github.com/gilgameshskytrooper/Prometheus/wiki/Quickstart)


### Project Hierarchy

| 	Directory   | Description 	                                                               				|
| ------------- | ----------------------------------------------------------------------------------------- |
| [NCS314/](NCS314)  | Fork with modified version of GRA and AFCH's Nixie Clock Arduino Sketches  |
| [assets/](assets/)  | Content related the the aesthetic presentation of this project such as images  |
| [hberg32/](hberg32)  | Original codebase as well as schematic that hberg32 kindly gave to me to use for this project  |
| [source/expressnode-python3](source/expressnode-python3)  | Where all the original implementation based on [Python3](https://www.python.org/) and [Node](https://nodejs.org/en/) is stored |
| [source/golang](source/golang) | The current implementation of the web server/hardware controller written in [Golang](https://golang.org/) |



Prometheus is the clock that does everything. The 3 main things I want to integrate in this project are:

1. Connect the clock to a bed shaker so my alarm clock can wake me up discretely without waking up my roommate, but also the ability to wake me up in a super noisy fashion by playing an alarm tone through my speaker system for when I really need to wake up.
2. A way to set the alarm via my phone, iPad, or Browser through a polished and intuitive web-interface controller as well as an easy way to set a new alarm sound.
3. Display time using nixie cathode tubes displays.

The current golang program accomplishes all 3 of these things marvelously. Try the [Wiki](https://github.com/gilgameshskytrooper/Prometheus/wiki) if any information is lacking down here.

## Implementation

### Inspiration
hberg32 has already successfully implemented some of the code as well as the hard wiring for a similar project of his. His original code as well as his original schematic is contained in the [hberg32](hberg32/) directory of this repo. His project can be found at: [Merciless Pi Alarm Clock](https://hackaday.io/project/4922-merciless-pi-alarm-clock).

### Hardware

![Custom Wire](assets/barrelplugwire.jpeg)

The hardware in this project is heavily based on hberg32's schematics, which can be found at [Merciless Pi Alarm Schematic](hberg32/PiAlarm.fzz). The schematics of my project can be found at [Prometheus Clock Schematic](/assets/AtomicClockSchematic.fzz).

A few notes on my set-up. My Raspberry Pi has a separate (standard 5V) power source separate from the rest of the Prometheus Clock. The rest of my Project is powered by a 12V @ 2000mA DC Power source, which powers both the Nixie Clock and the Bed Shaker.

As of now, the DC power goes though the breadoard, and goes to both the Nixie Clock, (via a custom barrel plug to breadboard wire I made) as well as a L293D which draws a consistant amount from the main circuit to ensure the clock receives enough V's and the Bed Shaker doesn't fry from too much.

You can read more [here](https://github.com/gilgameshskytrooper/Prometheus/wiki/Hardware-Set-Up)

### [Remote Control Functionality](/source/golang/main.go)

![Demo](assets/AtomicAlarmUI.PNG)

The UI portion of this project consists of a Golang web server running on the Raspberry Pi which can be accessed on any internet capable browser. The web server serves up the intial [index.html](source/golang/public/index.html) page at the root of the webserver (e.g. 111.111.111:3000 where 111.111.111 is the wlan0 IP of the Raspberry Pi). Through heavy use of [Vue.js](http://vuejs.org/), this page acts very must like a single page application. It makes heavy use of the AJAX post requests for the switch buttons, file uploads, and time form submit. It populates the values of the alarm times, sound, and vibration buttons with the values stored in [alarms.json](source/golang/public/json/alarms.json) which holds the configuration data for the 4 alarms. When the user submits the form, the Golang web server handles the request as a put request, reads the data, changes the internally stored values for the 4 alarms, and writes back the values into [alarms.json](source/golang/public/json/alarms.json). The only button on the page that requires a hard refresh is the snooze button at the top because the web server needs to compute the value of +10 minutes on the currenly running alarm.

A working model of the web interface can be found here: [Web Interface Showcase](https://atomicalarmui.herokuapp.com/)

When the requested the root [https://atomicalarmui.herokuapp.com/](https://atomicalarmui.herokuapp.com/), it will send [index.html](https://atomicalarmui.herokuapp.com/index.html), it loads it from the alarm configuration files: [alarm1.json](https://atomicalarmui.herokuapp.com/json/alarm1.json), [alarm2.json](https://atomicalarmui.herokuapp.com/json/alarm2.json), [alarm3.json](https://atomicalarmui.herokuapp.com/json/alarm3.json), and [alarm4.json](https://atomicalarmui.herokuapp.com/json/alarm4.json).

### [Main Alarm Logic](/source/golang/main.go)
Using a 3rd party library, Golang is also able to control the hardware interfaces (bed vibrator and speakers). The main logic for when to start running an alarm is a cron task that runs once a minute: if an alarm time configuration matches the current time, then run the relevant waking methods (i.e. vibration/sound)

More specific implementation information is written [here](/source/golang/README.md).

## [Where's My WiFi?](SetUpEduroamOnPi.md)
Because my school happens to disable ssh and VNC connections for users on the guest network (presumably for security reasons), I needed to set up my Raspberry Pi to work nicely with the school's eduroam. However, getting this to work was quite the struggle, and it seems to be a common issue for aspiring inventors trying to get their Raspberry Pi to work on their school's implementation of eduroam. Therefore, I carefully documented the steps I took to connect my Pi to the encrypted network. For anyone having trouble connecting their Pi (or any single-board computers such as chip) to eduroam, I encourage you to take a look at this document.

[Setting Up RPi to work with Eduroam](SetUpEduroamOnPi.md)

## Contact
Feel free to contact me at [leeas@stolaf.edu](mailto:leeas@stolaf.edu) if you have any suggestions, or want to contribute to this project.

## Special Thanks
[hberg32](https://hackaday.io/hberg32) was super helpful in helping this project become what it is today. I would not even know where to start to build such an alarm clock without his guidance.

Also, AFCH from [GRA & AFCH](https://github.com/afch) who produces the nixie clock kit I bought was also monumental in helping me modify his Arduino Sketch and to add serial USB communication functionality between the Pi and the Clock.

![Desk](assets/desk.jpg)
