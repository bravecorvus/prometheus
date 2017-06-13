#Node.js Web Server for Atomic Pi Project
##Andrew Lee

###[Main Project](https://github.com/gilgameshskytrooper/AtomicClock)

![Atomic Alarm UI](https://github.com/gilgameshskytrooper/AtomicClock/raw/master/assets/AtomicAlarmUI.PNG)

###Credits
I used the following technologies to make this website:
The HTML is loosely based on Gokul S Krishnan's [simple_alarm](https://github.com/gsk1692/simple_alarm).
[jQuery](http://jquery.com/) is a dependency of Bootstrap.
[node.js](https://nodejs.org/en/) provides the RESTful services needed to run the UI.
[npm](https://www.npmjs.com/) is the package manager for node.js I needed to install various tools for the project.

###Role in the Atomic Clock Project
This server runs the web server that controls the Alarm Configuration files. There are 4 of them: [alarm1.json](/public/json/alarm1.json), [alarm2.json](/public/json/alarm2.json), [alarm3.json](/public/json/alarm3.json), and [alarm4.json](/public/json/alarm4.json). The main alarm clock function will use the configuraton files as the basis for when to run an alarm and to decide when an alarm was enabled or disabled.

###Installation
If you haven't cloned the main Atomic Clock Project
```
> sudo apt-get install git
> git clone https://github.com/gilgameshskytrooper/AtomicClock.git
```

Next, we want to make sure you have node.js and npm:
```
> sudo apt-get install node npm
```

Next we want to install the node app dependencies for ```express```, ```body-parser```, and ```jsonfile```. Furthermore, we will need another node application called ```forever``` to run the web server as a daemon (basically useful for running this in the background without having to keep a terminal window open)

```
> cd AtomicClock/source/webinterface
> sudo node install
> sudo node install forever -g
```

At this point, we will have everything we will need to run the server.

The most basic way to run the server is

```
> cd /path_to_webinterface_root/
> node server.js
```

However, if you ssh'ed into your Pi, or close the terminal window by mistake, this will close the web server since it will be running as a main process on the terminal window.

Instead, I suggest running the server through ```forever``` which will run the web server in the background, if your ssh connection were to time out, or disconnect, it would still be running.

```
> cd /path_to_webinterface_root/
> sudo forever start server.js
```

If you are not running the server for the first time, you will need to add the ```-a``` tag (for append log file rather than create new) to the forever command or else forever will throw an error due to there being a log file already.

```
> sudo forever start -a server.js
```

To test your server, use any device with a web browser and connect to [111.111.111:3000](111.111.111:3000) (where 111.111.111 is the IP of your Pi. If you don't know this value, go to your Pi and run ```sudo ifconfig```. It will be the value listed as inet addr:). **Make sure you add ":3000" after the address in the browser as this specifies you will be sending the get request to port 3000 of your Pi**

Congratulations, you now have a fully functional Atomic Clock User Interface.

###How it all works
When the requested the root [111.111.111:3000](111.111.111:3000), it will send [index.html](/public/index.html), it loads it from the alarm configuration files: [alarm1.json](/public/json/alarm1.json), [alarm2.json](/public/json/alarm2.json), [alarm3.json](/public/json/alarm3.json), and [alarm4.json](/public/json/alarm4.json).

When the user fills out the form, it will update the configuration files and reload the page. **Note: You cannot send the form from the [index.html link](/public/index.html), you must use the [111.111.111:3000](111.111.111:3000)**