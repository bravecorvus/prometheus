import threading
import json
import time
import os
# from datetime import datetime
from datetime import timedelta          #for the time on the rpi end
from observable import Observable
from datetime import datetime
from datetime import date
from watchdog.observers import Observer
from watchdog.events import FileSystemEventHandler
from apscheduler.scheduler import Scheduler

def getsnoozeconfig(arg):
    with open(arg) as data_file:
        data = json.load(data_file)
    snoozing = data['snooze']
    if snoozing == "on":
        return(True)
    else:
        return(False)

def gettimeconfig(arg):
    with open(arg) as data_file:    
        data = json.load(data_file)
    return(datetime.strptime(data["time"], '%H:%M').time().strftime('%H:%M'))

def getsoundconfig(arg):
    with open(arg) as data_file:    
        data = json.load(data_file)
    soundconfigstring = data["sound"]
    if soundconfigstring == "on":
        return(True)
    else:
        return(False)

def getsoundconfigstring(arg):
    with open(arg) as data_file:    
        data = json.load(data_file)
    return(data["sound"])

def getvibrationconfig(arg):
    with open(arg) as data_file:    
            data = json.load(data_file)
    vibrationconfigstring = data["vibration"]
    if vibrationconfigstring == "on":
        return(True)
    else:
        return(False)

def getvibrationconfigstring(arg):
    with open(arg) as data_file:    
            data = json.load(data_file)
    vibrationconfigstring = data["vibration"]
    if vibrationconfigstring == "on":
        return(True)
    else:
        return(False)


def snooze(arg):
    arg.alarmtime = datetime.now() + timedelta(minutes = 10)
    with open("webinterface/public/json/"+arg.id+".json", "w") as outfile:
        json.dump({'id':arg.id,'time':str(arg.alarmtime.time().strftime('%H:%M')),'sound':arg.soundstring, 'vibration':arg.vibrationstring}, outfile)
class ALARM:
    def __init__(self, id):
        self.id = id
        self.path = "webinterface/public/json/"+id+".json"
        self.alarmtime = gettimeconfig(self.path)
        self.soundstring = getsoundconfig(self.path) #boolean for easier if else behavior
        self.sound = getsoundconfig(self.path) #boolean for easier if else behavior
        self.vibration = getvibrationconfig(self.path)
        self.vibrationstring = getvibrationconfigstring(self.path)
        self._alarm_thread = None
        self.update_interval = 1
        self.event = threading.Event()

    def updateclock(self, path):
        self.alarmtime = gettimeconfig(self.path)
        self.sound = getsoundconfig(self.path)
        self.soundstring = getsoundconfigstring(self.path)
        self.vibration = getvibrationconfig(self.path)
        self.vibrationstring = getvibrationconfigstring(self.path)
    def play(self):
        if self.sound:
            self.p = subprocess.Popen("mpg123 webinterface/public/assets/alarm.m4a", stdout=subprocess.PIPE, shell=False)
        if self.vibration:
            print("BZZZ")

    def killSound(self):
        if self.sound == False:
            try:
                self.p.kill()
            except:
                "o well"
            
    def killVibration(self):
        if self.vibration == False:
            print("vibration stopped")



def nextAlarm(alarmarray):
    delta = []
    for i in alarmarray:
        delta.append(datetime.strptime(i.alarmtime, '%H:%M').time())
    minimum = 0
    for count, data in enumerate(delta):
        deltanow = datetime.combine(date.min, data)-datetime.combine(date.min, datetime.now().time())
        if count == 0:
            continue
        elif deltanow.total_seconds() < 0:
            continue
        elif deltanow.total_seconds() < minimum:
            minimum = count
    return(minimum)

def AlarmRun(arg):
    start_min = datetime.strptime(arg.alarmtime, '%H:%M').time()       #  calling date to set the beginning of query range for the present day
    start_max = datetime.strptime(arg.alarmtime, '%H:%M').time() + datetime.timedelta(minutes = 10)    #  calling endDate to limit the query range to the next 14 days. change tmedelta(days) to set the range
    current = datetime.now().time()
    deltamin = datetime.combine(date.min, start_min)-datetime.combine(date.min, current)
    deltamax = datetime.combine(date.min, start_max)-datetime.combine(date.min, current)
    if deltamin < 600 and deltamin > 0 or deltamax < 600 and deltamax > 0:
        arg.play()

def callable_func(allalarms):
    if len(allalarms) == 4:
        for i in allalarms:
            del i
    os.system("clear")
    alarm1 = ALARM("alarm1")
    allalarms.append(alarm1)
    alarm2 = ALARM("alarm2")
    allalarms.append(alarm2)
    alarm3 = ALARM("alarm3")
    allalarms.append(alarm3)
    alarm4 = ALARM("alarm4")
    allalarms.append(alarm4)
    AlarmRun(allalarms[nextAlarm(allalarms)])

class MyHandler(FileSystemEventHandler):
    def on_modified(self, event):
        snoozed = getsnoozeconfig("webinterface/public/json/snooze.json")
        if snoozed == True:
            alarm1.killSound()
            alarm1.killVibration()
            alarm1.snooze()
            alarm2.killSound()
            alarm2.killVibration()
            alarm2.snooze()
            alarm3.killSound()
            alarm3.killVibration()
            alarm3.snooze()
            alarm4.killSound()
            alarm4.killVibration()
            alarm4.snooze()


if __name__ == "__main__":
    allalarms = []
    sched = Scheduler(standalone=True)
    sched.add_interval_job(callable_func(allalarms),seconds=10)  #  define refresh rate. Set to every 10 seconds by default
    sched.start()
    try:
        while True:
            time.sleep(1)
    except KeyboardInterrupt:
        observer.stop()
    observer.join()