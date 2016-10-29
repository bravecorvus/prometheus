# Atomic Clock

### By Andrew Lee

![Atomic Clock Cover](assets/cover.jpg)

The idea was to have a clock that does everything. The 3 main things I want to integrate in this project are:

1) Connect the clock to a bed shaker so my alarm clock can wake me up discretely without waking up my roomate

2) Utilize the Bluetooth capabilities of my Pi to utilize it as a Bluetooth audio receiver, which would then output to my speakers.

3) A way to set the alarm via my phone. (Whether through internet triggerring applications such as IFFFT, or HomeKit RPi hacks has yet to be determined).



###Extra
4) Display time using nixie cathode tubes.

###Implementation
hberg32 has already successfully implemented some of the code as well as the hard wiring for a similar project of his. My project however, will include a web framework of some sort to be able to pass the remote alarm commands into the main program. At the same time, I am really excited to implement a nixie tube clock to top it all off.

[Merciless Pi Alarm Clock](https://hackaday.io/project/4922-merciless-pi-alarm-clock)

My plan is to improve on his code as well as add the alarm remote capabilities in order create the Atomic Alarm Clock. At the same time, I have worked in implementing a software-hardware solution in a previous class which I entirely wrote in C++. However, this time, I am exited to be able to use the extensive Python libraries to make this a full project.


The following chart describes the various classes, methods, and variables I would need to implement the AtomicClock program.

![Implementation Diagram](assets/implementation.jpg)


###Where's My WiFi?
Because my school happens to disable ssh and VNC connections for users on the guest network (presumably for security reasons), I needed to set up my Raspberry Pi to work nicely with the school's eduroam. However, getting this to work was quite the struggle, and it seems to be a common issue for aspiring inventors trying to get their Raspberry Pi to work on their school's implementation of eduroam. Therefore, I carefully documented the steps I took to connect my Pi to the encrypted network. For anyone having trouble connecting their Pi (or any single-board computers such as chip) to eduroam, I encourage you to take a look at this document.

[Setting Up RPi to work with Eduroam](SetUpEduroamOnPi.md)

###Contact
Feel free to contact me at (leeas@stolaf.edu) if you have any suggestions, or want to contribute to this project.

###Note:
Everything under the pialarm directory is the work of hberg32. I am currently in the process of reading through his code to adapt my own modifications on it.

###Special Thanks
It goes without saying that the real work was done by hberg32, and I am just making improvements to what is already a amazing project.