import arrow
import sys
from observable import Observable
import threading
import time
import os
import signal
import subprocess
import json
from pprint import pprint
from datetime import datetime
from datetime import timedelta
from watchdog.observers import Observer
from watchdog.events import FileSystemEventHandler

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
    def play(self, sound, vibration):
        if self.sound:
            self.p = subprocess.Popen("mpg123 webinterface/public/assets/alarm.m4a", stdout=subprocess.PIPE, shell=False)
        if self.vibration:
            print("BZZZ")

    def killSound(self, sound):
        if self.sound == False:
            self.p.kill()
            
    def stopVibration(self, vibration):
        if self.vibration == False:
            print("vibration stopped")

    def run(self):
        while True:
            self.event.wait(self.update_interval)
            if self.event.isSet():
                break
            now = datetime.now()
            if self._alarm_thread and self._alarm_thread.is_alive():
                alarm_symbol = '+'
            else:
                alarm_symbol = ' '
            sys.stdout.write("\r%02d:%02d:%02d %s" 
                % (now.hour, now.minute, now.second, alarm_symbol))
            sys.stdout.flush()

    def set_alarm(self):
        now = datetime.now()
        alarm = now.replace(hour=int(datetime.strptime(self.alarmtime, '%H:%M').time().strftime('%H')), minute=int(datetime.strptime(self.alarmtime, '%H:%M').time().strftime('%M')))
        delta = int((alarm - now).total_seconds())
        if delta <= 0:
            alarm = alarm.replace(day=alarm.day + 1)
            delta = int((alarm - now).total_seconds())
        if self._alarm_thread:
            self._alarm_thread.cancel()
        self._alarm_thread = threading.Timer(delta, self.play)
        self._alarm_thread.daemon = True
        self._alarm_thread.start()




class MyHandler(FileSystemEventHandler):
    def on_modified(self, event):
        snoozed = getsnoozeconfig("webinterface/public/json/snooze.json")
        if snoozed == True:
            now = arrow.now()
            alarmdeltas = []
            alarm1delta = now - arrow.get(str(datetime.now().date())+'-'+alarm1.alarmtime, 'YYYY-MM-DD-HH:mm:ss')
            alarm1delta = alarm1delta.format('HH:mm:ss')
            alarm2delta = now - arrow.get(str(datetime.now().date())+'-'+alarm2.alarmtime, 'YYYY-MM-DD-HH:mm:ss')
            alarm2delta = alarm2delta.format('HH:mm:ss')
            alarmdeltas.append(alarm2delta)
            alarm3delta = now - arrow.get(str(datetime.now().date())+'-'+alarm1.alarmtime, 'YYYY-MM-DD-HH:mm:ss')
            alarm3delta = alarm3delta.format('HH:mm:ss')
            alarmdeltas.append(alarm3delta)
            alarm4delta = now - arrow.get(str(datetime.now().date())+'-'+alarm1.alarmtime, 'YYYY-MM-DD-HH:mm:ss')
            alarm4delta = alarm4delta.format('HH:mm:ss')
            alarmdeltas.append(alarm4delta)
            minimum = datetime.strptime(str(arrow.get('00:00:00', 'HH:mm:ss').format('HH:mm:ss')),'%H:%M:%S').time()-datetime.strptime(str(arrow.get(alarm1delta, 'HH:mm:ss').format('HH:mm:ss')),'%H:%M:%S').time()
            for i in alarm4deltas:
                posmin = datetime.strptime(str(arrow.get('00:00:00', 'HH:mm:ss').format('HH:mm:ss')),'%H:%M:%S').time()-datetime.strptime(str(arrow.get(i, 'HH:mm:ss').format('HH:mm:ss')),'%H:%M:%S').time()
                print(posmin)

        else:
            alarm1 = ALARM("alarm1")
            alarm1.set_alarm()
            alarm1.run()
            alarm2 = ALARM("alarm2")
            alarm2.set_alarm()
            alarm2.run()
            alarm3 = ALARM("alarm3")
            alarm3.set_alarm()
            alarm3.run()
            alarm4 = ALARM("alarm4")
            alarm4.setalarm()
            alarm4.run()

if __name__ == "__main__":
    alarm1 = ALARM("alarm1")
    alarm1.set_alarm()
    alarm1.run()
    alarm2 = ALARM("alarm2")
    alarm2.set_alarm()
    alarm2.run()
    alarm3 = ALARM("alarm3")
    alarm3.set_alarm()
    alarm3.run()
    alarm4 = ALARM("alarm4")
    alarm4.setalarm()
    alarm4.run()
    obs = Observable()
    event_handler = MyHandler()
    observer = Observer()
    observer.schedule(event_handler, path='webinterface/public/json/', recursive=False)
    observer.start()
    try:
        while True:
            time.sleep(1)
    except KeyboardInterrupt:
        observer.stop()
    observer.join()