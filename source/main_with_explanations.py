import RPi.GPIO as GPIO
from collections import OrderedDict
from datetime import datetime
from datetime import timedelta
import time
import json
import re
import pyinotify #you are going to have to install this via Pip or from source
import os
import signal
import subprocess

GPIO.setmode(GPIO.BCM)
GPIO.setup(17, GPIO.OUT) #GPIO 17 EN 1, 2
GPIO.setup(5, GPIO.OUT) #GPIO 5 INPUT 1
GPIO.setup(6, GPIO.OUT) #GPIO 6 INPUT 2

currentlyModifying = False #Since the event based stuff will look at when a JSON file is written/closed, we want to be able to modify the JSON files within this Python Program without setting off the event handler. Hence, This will be modified so that we don't get a weird recurive structure that traps the program in an infinite loop
wm = pyinotify.WatchManager()  # Watch Manager
mask = pyinotify.IN_CLOSE_WRITE | pyinotify.IN_OPEN  # watched events
timeWatcher = wm.add_watch('webinterface/public/json/time.json', mask, rec=False) #This adds a watcher that watches the time.json file
alarm1Watcher = wm.add_watch('webinterface/public/json/alarm1.json', mask, rec=False) #This adds a watcher that watches the alarm1.json file
alarm2Watcher = wm.add_watch('webinterface/public/json/alarm2.json', mask, rec=False) #This adds a watcher that watches the alarm2.json file
alarm3Watcher = wm.add_watch('webinterface/public/json/alarm3.json', mask, rec=False) #This adds a watcher that watches the alarm3.json file
alarm4Watcher = wm.add_watch('webinterface/public/json/alarm4.json', mask, rec=False) #This adds a watcher that watches the alarm4.json file
snoozeWatcher = wm.add_watch('webinterface/public/json/snooze.json', mask, rec=False) #This adds a watcher that watches the snooze.json file

class Alarm:
    def __init__(self, arg): #where arg is the name of a configuration file
        with open(arg) as data_file:
            data = json.load(data_file)
        self.name = data["name"] #Since the json.load will construct a dictionary, we can use the dictionary key to get the dictionary value
        self.time = data["time"] #Since the json.load will construct a dictionary, we can use the dictionary key to get the dictionary value
        self.sound = data["sound"] #Since the json.load will construct a dictionary, we can use the dictionary key to get the dictionary value
        self.vibration = data["vibration"] #Since the json.load will construct a dictionary, we can use the dictionary key to get the dictionary value
        self.currentStatus = "off" #When initialized, we don't want the alarm to be on
    def updateAlarm(self, arg): #identical to the class contructor __init__, but the constructor can only be called when the class is initialized
        with open(arg) as data_file:
            data = json.load(data_file)
        self.name = data["name"] #Since the json.load will construct a dictionary, we can use the dictionary key to get the dictionary value
        self.time = data["time"] #Since the json.load will construct a dictionary, we can use the dictionary key to get the dictionary value
        self.sound = data["sound"] #Since the json.load will construct a dictionary, we can use the dictionary key to get the dictionary value
        self.vibration = data["vibration"] #Since the json.load will construct a dictionary, we can use the dictionary key to get the dictionary value

def readTime(arg): #specifically to get time from the Node updated file time.json
    with open(arg) as data_file:
        data = json.load(data_file)
    return(data["time"])

def readBool(arg): #specifically to get the snooze file from the snooze.json file
    with open(arg) as data_file:
        data = json.load(data_file)
    return(data['snooze'])


class Alarms:
    def __init__(self): #alarm1-4, time, snooze will be strings of json filenames
        currentlyModifying = True
        with open('webinterface/public/json/snooze.json', 'w') as outfile: #We don't want the snooze to be on when we initialize the program
            json.dump({"snooze":"off"}, outfile)
        currentlyModifying = False
        self.snooze = "off"
        self.alarm1 = Alarm("webinterface/public/json/alarm1.json") #see __init__ in class Alarm
        self.alarm2 = Alarm("webinterface/public/json/alarm2.json") #see __init__ in class Alarm
        self.alarm3 = Alarm("webinterface/public/json/alarm3.json") #see __init__ in class Alarm
        self.alarm4 = Alarm("webinterface/public/json/alarm4.json") #see __init__ in class Alarm
        self.time = readTime("webinterface/public/json/time.json") #gets the current time in string format and stores it into self.time

    def updateAlarms(self): #Pretty much same as __init__, but we don't need to make sure snooze is off since the alarms are already initialized, and updateSnooze() might be what is triggering this function anyways
        self.alarm1.updateAlarm("webinterface/public/json/alarm1.json")
        self.alarm2.updateAlarm("webinterface/public/json/alarm2.json")
        self.alarm3.updateAlarm("webinterface/public/json/alarm3.json")
        self.alarm4.updateAlarm("webinterface/public/json/alarm4.json")

    def updateSnooze(self): #Called when the snooze.json file is modified
        self.snooze = readBool("webinterface/public/json/snooze.json")
        if self.snooze == "on": #only run if snooze is set to "on" rather than "off"
            for i in [self.alarm1, self.alarm2, self.alarm3, self.alarm4]:
                if i.currentStatus == "on": #Loop through the current alarms and see which one is currently on
                    i.currentStatus = "off" #Turn alarm off
                    snoozetime = (datetime.strptime(i.time, '%H:%M') + timedelta(minutes=5)).strftime('%H:%M') #Complicated, but basically saying take the time of i, add 5 minutes to it, and save it back snoozetime
                    currentlyModifying = True #Make sure we don't set off the alarm#.json write/close handler
                    with open('webinterface/public/json/'+i.name+'.json', 'w') as outfile:
                        json.dump(OrderedDict([("name",i.name),("time",snoozetime),("sound",i.sound),("vibration",i.vibration)]), outfile) #update that alarm's configuration file to have the snoozetime (time + 5 minutes)
                    currentlyModifying = False #Then return currentlyModifying back to false so the program can continue its normal operations
                    break #Since we already found the alarm which was currently on, we have no need to look through the other alarms
            currentlyModifying = True #Make sure we don't set off the snooze.json write/close handler
            with open('webinterface/public/json/snooze.json', 'w') as outfile:  #Since we already executed what needed to be done when snooze happened, set snooze back to its original state which is false
                json.dump({"snooze":"off"}, outfile)
            currentlyModifying = False
            self.snooze = "off" #Return original event handler capabilities
    def updateTime(self): #Called when time.json is modified once a minute by the Node server
        self.time = readTime("webinterface/public/json/time.json")
        if self.alarm1.time == self.time: #The next 4 clauses will set the alarm's currentStatus to be "on" if any of them share the same time with the current time
            self.alarm1.currentStatus = "on"
        elif self.alarm2.time == self.time:
            self.alarm2.currentStatus = "on"
        elif self.alarm3.time == self.time:
            self.alarm3.currentStatus = "on"
        elif self.alarm4.time == self.time:
            self.alarm4.currentStatus = "on"
        for i in [self.alarm1, self.alarm2, self.alarm3, self.alarm4]: #Using the currentStatus variable we set above, we will see loop through an run the alarm sequence for the first one that is on:
            if i.currentStatus == "on":
                if i.vibration == "on" and i.sound == "off": #self explanatory
                    while i.currentStatus == "on" and i.vibration == "on": #These will only get reset by the updateSnooze() function or updateAlarm() function but not the time function (e.g. we don't want the alarm to stop running just because it's been one minute)
                        self.updateSnooze() #making sure if the user pressed the snooze button, the necessary steps will be run
                        self.updateAlarms() #Making sure if the user turned off the alarm itself on the interface, then the Alarms class variables will reflect the new values
                        self.runVibration() #Run the vibration sequence
                        time.sleep(2) #Pause program for 2 seconds
                    GPIO.output(17, GPIO.LOW) #Takes off the Enable off of the EN1, 2 (Pin 1 on the L293D) (Halts the Vibration )
                    GPIO.output(5, GPIO.LOW) #Sets In 1 to 0
                    GPIO.output(6, GPIO.LOW) #Sets In 2 to 0
                elif i.vibration == "on" and i.sound == "on":    
                    playsound = subprocess.Popen("exec mpg123 webinterface/public/assets/alarm.m4a", stdout=subprocess.PIPE, shell=True) #basically, run this command (the exec in the front is to ensure we can kill the process later on)
                    while i.currentStatus == "on" and i.vibration == "on":
                        self.updateSnooze()
                        self.updateAlarms()
                        self.runVibration()
                        time.sleep(2)
                    playsound.kill() #Once the loop ends (either by snooze, or an alarm update) then kill the sound
                    GPIO.output(17, GPIO.LOW) #Kill the vibration
                    GPIO.output(5, GPIO.LOW)
                    GPIO.output(6, GPIO.LOW)
                elif i.vibration == "off" and i.sound == "on": 
                    playsound = subprocess.Popen("exec mpg123 webinterface/public/assets/alarm.m4a", stdout=subprocess.PIPE, shell=True)
                    while i.currentStatus == "on" and i.vibration == "on":
                        self.updateSnooze()
                        self.updateAlarms()
                        time.sleep(2)
                    playsound.kill()
    def runVibration(self):
        GPIO.output(5, GPIO.HIGH) #Enable In 1
        GPIO.output(6, GPIO.LOW)   #Disable In 2
        GPIO.output(17, GPIO.HIGH) #Enable Enable 1, 2 (The top two are just the input variables, nothing happens until this enable is set to 1)
        time.sleep(2) #Run the vibration for 2 seconds
        GPIO.output(17, GPIO.LOW) #Turn off the vibration and wait for the next loop iteration to see if the vibration should keep on running



main = Alarms()#main Alarms events

class EventHandler(pyinotify.ProcessEvent): #Lotta magic I don't really understand, but I still kinda understand
    def process_IN_CLOSE_WRITE(self, event): # the class event has the variable pathname which is the full Linux path to the file
        stringpath = list(event.pathname)[::-1] #Grab the string pathname, and reverse it: ("/home/pi/Projects/Personal/AtomicClock/Macbook/source/webinterface/public/json/alarm1.json" would become "nosj.1mrala/nosj/cilbup/ecafretnibew/ecruos/koobcaM/kcolCcimotA/lanosreP/stcejorP/ip/emoh/"")
        counter = 0
        for i in stringpath:
            if i == "/": #Basically, count the # of characters we pass before getting the first "/" and the string up until that point reversed will be the correct file name
                break; 
            counter+=1
        filename = ''.join(stringpath[0:counter][::-1])
        if currentlyModifying == False and filename == "snooze.json": #If the file name is snooze.json and this program isn't explicitly modifying the file
            main.updateSnooze() #snooze.json must have been updated. Therefore, run updateSnooze()
        elif re.search(r'alarm*', filename): #Match the regular expression alarm* (i.e. alarm1.json alarm2.json alarm3.json alarm4.json)
            main.updateAlarms() #alarm#.json must have been updated. Therefore, run updateAlarms()
        elif filename == "time.json": #Time updated by Node Server
            main.updateTime() #Run updateTime()


handler = EventHandler()
notifier = pyinotify.Notifier(wm, handler)
notifier.loop()
GPIO.cleanup() #This will really never run since I never put an explicit user exiting function (the only way to exit this program is via Control-C), but it looks nice anyways