# I Just Want to Get This Up and Running

If this is you, then this section will describe the minimum number of steps to get this up and running (I suggest against this since my setup is pretty specific to my needs.

### Software Stuff

Going into the command line, do the following

```
$sudo apt-get install git python3 pip3 mpg123 node npm
$sudo pip3 install pyinotify
$git clone https://github.com/gilgameshskytrooper/AtomicClock.git
$cd AtomicClock/source/webinterface
$sudo node install
$sudo node install forever -g
```


### Hardware Stuff
This part is specific to the hardware you will connect, but here I will share the setup assuming you are using the same setup as me. (e.g. you already own a bed vibrator that runs 12V @ 0.5A and center pin positive, you have a separate power source that is 12V @1+A, and a sound system)

You will need the following Items
| ITEMS | DESCRIPTION | QUANTITY|
|-------|---------------------------------|--------|
| Raspberry Pi| Runs all the logic and supplies 5V to the motor driver | x1 |
| L293D | Motor Driver that is will either allow the circuit to the vibrator, or disallow | x1 (I would bulk buy more) |
| Breadboard | Prototyping and connecting everything | x1 |
| Wires | Breadboard wires. Make sure they can withstand some amount of current, if you don't have a Pi breadboard ribbon, get female to male in addition to male to male| x10 |
| USB Audio Interface | Default Pi audio out is not good, so this provides a way to get clean audio | x1 |
| SonicAlert SS12V Bed Shaker | Bed shaker, runs 12V @ 0.5A (peak 12V @ 1.0A) Center Pin Positive | x1 |
| 12V @ 1+A DC Power Supply | The main point here is that it is DC, and that it can supply 12V @ 1A or greater than 1A | x1 |

Connect a circuit using the schematic I uploaded under /root/assets/AtomicClockSchematic.fzz as a model.

If you are having trouble creating the circuit, maybe the following notes I have on how the L293D is layed out can help.

The schematic for the L293D is as follows:

```
Enable 1, 2 Driver Channels      [1   u   16]  Chip Power (5V)
                  Driver Input 1 [2       15] Driver Input 4
                 Driver Output 1 [3       14] Driver Output 4
                        Ground 1 [4       13] Ground 4
                        Ground 2 [5       12] Ground 3
                 Driver Output 2 [6       11] Driver Output 3
                  Driver Input 2 [7       10] Driver Input 3
               Motor Power (12V) [8        9] Enable 3, 4 Driver Channels
```

source: [TexasInstruments](http://www.ti.com/lit/ds/symlink/l293.pdf)

Where the u at the top center represents the divet in the chip to show which side is up. We only need to use one side of the chip.
This chip was initially created to be able to run two stepper motors (stepper meaning bidirectional), one on the left and one on the right.

In this conventional set-up, (as most tutorials online and on Youtube are doing), the left motor will connect to Output 1 (3) and Output 2 (6) while the right motor will connect to Output 3 (11) and output 4 (14). Both motors are powered by the V+ going in from (8) and the chip itself is powered by 16. The next part was hard to understand for me, so bear with me, but Let's say you want to run the left motor "forward" e.g. V+ goes out from Output 1 and the ground (-) goes from the motor to Output 2, then through Ground 1/2 Then you would set 

```
Input 1 = True | Input 2 = False | then Enable 1, 2 = True
```

(Without the Enable 1, 2 = True, nothing would happen). Then the positive 12V would go from 8 to 3 to the motor, then the ground would be motor to 6, to 5 then grounded. To reverse direction of the same motor, you would send

```
Input 1 = False | Input 2 = True | then Enable 1, 2 = True.
```

However, since we are not running a stepper motor, and we only want "center pin positive" current to be flowing, we will only use the first inputs: 

```
Input 1 = True | Input 2 = False | then Enable 1, 2 = True.
```

Furthermore, after looking at many tutorials, I could not figure out why my circuit was not working, and I realized that I had to ground 12 and 13  as well as 4 and 5 since the Chip's 5V also needed to ground somewhere.

### Sound
To get good sound, you will need to output the sound through the USB sound interface.

I don't remember the specific steps to get Pi to output sound through the USB, but I believe you just need to create a file via

```
$nano ~/.asoundrc

pcm.!default {
    type hw
    card 1
}

ctl.!default {
    type hw
    card 1
}
```

[CTRL-O, RETURN, CTRL-X to Save and exit]

### Run The Alarm Clock

You will need 2 terminal windows open. If you are SSH'ing, I suggest you open a VNC Window as this will allow for the programs to continue running even if you close the session. (SSH will close the processes running if you exit the SSH shell)

```
$cd /rootofAtomicClockProject/source/webinterface/
$forever server.js
```

The above steps will start the web server

Next get your IP address via the following command

```
$ifconfig
```

Copy down the address that is listed as wlan0 inet addr. This address is how you will use another device on the same WiFi network to access the website. (i.e. http:111.11.111.111:3000 where you will change 111.11.111.111 to be your IP but keep :3000 at the end to specify you will access the Pi via your web browser at port 3000)

Next, to start the main program,

```
$cd /rootofAtomicClockProject/source/
$python3 main.py
```

## Have Fun!