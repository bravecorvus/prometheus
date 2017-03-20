from collections import OrderedDict
from datetime import datetime
from datetime import timedelta
import time
import json
import re
import pyinotify
import os
import signal
import subprocess

currentlyModifying = False
wm = pyinotify.WatchManager()  # Watch Manager
mask = pyinotify.IN_CLOSE_WRITE | pyinotify.IN_OPEN  # watched events
timeWatcher = wm.add_watch('webinterface/public/json/time.json', mask, rec=False)
alarm1Watcher = wm.add_watch('webinterface/public/json/alarm1.json', mask, rec=False)
alarm2Watcher = wm.add_watch('webinterface/public/json/alarm2.json', mask, rec=False)
alarm3Watcher = wm.add_watch('webinterface/public/json/alarm3.json', mask, rec=False)
alarm4Watcher = wm.add_watch('webinterface/public/json/alarm4.json', mask, rec=False)
snoozeWatcher = wm.add_watch('webinterface/public/json/snooze.json', mask, rec=False)

class Alarm:
    def __init__(self, arg):
        with open(arg) as data_file:
            data = json.load(data_file)
        self.name = data["name"]
        # self.time = datetime.strptime(data[time], '%H:%M').time() #If I need to convert to time. For now, using string representation of time
        self.time = data["time"]
        self.sound = data["sound"]
        self.vibration = data["vibration"]
        self.currentStatus = "off"
    def updateAlarm(self, arg):
        with open(arg) as data_file:
            data = json.load(data_file)
        self.name = data["name"]
        # self.time = datetime.strptime(data[time], '%H:%M').time() #If I need to convert to time. For now, using string representation of time
        self.time = data["time"]
        self.sound = data["sound"]
        self.vibration = data["vibration"]

def readTime(arg):
    with open(arg) as data_file:
        data = json.load(data_file)
    return(data["time"])

def readBool(arg):
    with open(arg) as data_file:
        data = json.load(data_file)
    print(data['snooze'])
    return(data['snooze'])


class Alarms:
    def __init__(self): #alarm1-4, time, snooze will be strings of json filenames
        currentlyModifying = True
        with open('webinterface/public/json/snooze.json', 'w') as outfile:
            json.dump({"snooze":"off"}, outfile)
        currentlyModifying = False
        self.snooze = "off"
        self.alarm1 = Alarm("webinterface/public/json/alarm1.json")
        self.alarm2 = Alarm("webinterface/public/json/alarm2.json")
        self.alarm3 = Alarm("webinterface/public/json/alarm3.json")
        self.alarm4 = Alarm("webinterface/public/json/alarm4.json")
        self.time = readTime("webinterface/public/json/time.json")

    def updateAlarms(self):
        self.alarm1.updateAlarm("webinterface/public/json/alarm1.json")
        self.alarm2.updateAlarm("webinterface/public/json/alarm2.json")
        self.alarm3.updateAlarm("webinterface/public/json/alarm3.json")
        self.alarm4.updateAlarm("webinterface/public/json/alarm4.json")

    def updateSnooze(self):
        self.snooze = readBool("webinterface/public/json/snooze.json")
        if self.snooze == "on":
            for i in [self.alarm1, self.alarm2, self.alarm3, self.alarm4]:
                if i.currentStatus == "on":
                    i.currentStatus = "off"
                    snoozetime = (datetime.strptime(i.time, '%H:%M') + timedelta(minutes=5)).strftime('%H:%M')
                    currentlyModifying = True
                    with open('webinterface/public/json/'+i.name+'.json', 'w') as outfile:
                        json.dump(OrderedDict([("name",i.name),("time",snoozetime),("sound",i.sound),("vibration",i.vibration)]), outfile)
                    currentlyModifying = False
                    break
            currentlyModifying = True
            with open('webinterface/public/json/snooze.json', 'w') as outfile:
                json.dump({"snooze":"off"}, outfile)
            currentlyModifying = False
            self.snooze = "off"
    def updateTime(self):
        self.time = readTime("webinterface/public/json/time.json")
        if self.alarm1.time == self.time:
            self.alarm1.currentStatus = "on"
        elif self.alarm2.time == self.time:
            self.alarm2.currentStatus = "on"
        elif self.alarm3.time == self.time:
            self.alarm3.currentStatus = "on"
        elif self.alarm4.time == self.time:
            self.alarm4.currentStatus = "on"
        for i in [self.alarm1, self.alarm2, self.alarm3, self.alarm4]:
            if i.currentStatus == "on":
                if i.vibration == "on" and i.sound == "off":    
                    while i.currentStatus == "on" and i.vibration == "on":
                        self.updateSnooze()
                        self.updateAlarms()
                        self.runVibration()
                        time.sleep(2)
                elif i.vibration == "on" and i.sound == "on":    
                    playsound = subprocess.Popen("exec mpg123 webinterface/public/assets/alarm.m4a", stdout=subprocess.PIPE, shell=True)
                    while i.currentStatus == "on" and i.vibration == "on":
                        self.updateSnooze()
                        self.updateAlarms()
                        self.runVibration()
                        time.sleep(2)
                    playsound.kill()
                elif i.vibration == "off" and i.sound == "on":    
                    playsound = subprocess.Popen("exec mpg123 webinterface/public/assets/alarm.m4a", stdout=subprocess.PIPE, shell=True)
                    while i.currentStatus == "on" and i.vibration == "on":
                        self.updateSnooze()
                        self.updateAlarms()
                        self.runVibration()
                        time.sleep(2)
                    playsound.kill()
    def runVibration(self):
        print("Vibrating for 2 seconds")



main = Alarms()#main Alarms events

class EventHandler(pyinotify.ProcessEvent):
    def process_IN_CLOSE_WRITE(self, event):
        stringpath = list(event.pathname)[::-1]
        counter = 0
        for i in stringpath:
            if i == "/":
                break; 
            counter+=1
        filename = ''.join(stringpath[0:counter][::-1])
        if currentlyModifying == False and filename == "snooze.json":
            print("snooze update")
            # snoozeWatcher = wm.del_watch(5)
            main.updateSnooze()
            # snoozeWatcher = wm.add_watch('webinterface/public/json/snooze.json', mask, rec=False)
        elif re.search(r'alarm*', filename):
            print("alarm update")
            main.updateAlarms()
        elif filename == "time.json":
            main.updateTime()


handler = EventHandler()
notifier = pyinotify.Notifier(wm, handler)
notifier.loop()