#motor pin 22 & 23
#snooze pin 25

import pygame
import time
import datetime
from pygame.locals import *
import RPi.GPIO as GPIO
from Adafruit_ADS1x15 import ADS1x15
import signal
import sys
from subprocess import call

class alarmEntry:
	name = ""
	start = 0
	end = 0
	snoozesAllowed = 0
	buttonAction = ""
	daysOfWeek = ""
	halted = 0

#This method returns a struct_time representing the current time of day but with day/month/year fields zeroed out
def get_time():
	return time.strptime(datetime.datetime.today().strftime("%H:%M"),"%H:%M")

def shutdown(signal, frame):
	global running
	running = 0;
	log("Shutting down ")
	bleep.stop
	
	GPIO.cleanup()
	sys.exit(0)

def hup(signal, frame):
	global running
	running = 0;
	
def read_schedule():
	global alarms
	alarms = []
	
	log("reading alarm schedule")
	with open('/home/pi/pialarm/schedule.conf', 'r') as f:
		for line in f:
			line = line.strip()
			thisEntry = alarmEntry()
			(thisEntry.name, thisEntry.start, thisEntry.end, thisEntry.snoozesAllowed, thisEntry.buttonAction, thisEntry.daysOfWeek) = line.split(',')
			thisEntry.start = time.strptime(thisEntry.start, "%H:%M")
			thisEntry.end = time.strptime(thisEntry.end, "%H:%M")
			thisEntry.snoozesAllowed = int(thisEntry.snoozesAllowed)
			thisEntry.daysOfWeek = thisEntry.daysOfWeek.split(':')
			if(thisEntry.snoozesAllowed < 0):
				thisEntry.snoozesAllowed = 0
			alarms.append(thisEntry)
	f.closed

def print_schedule(signal, frame):
	for alarm in alarms:
		print(alarm.name + ": " + time.strftime("%H:%M", alarm.start) + " - " + time.strftime("%H:%M", alarm.end) + ", " + alarm.buttonAction);
	
def inBed():
	global inBedVal
	global lastBedCheckTime
	currentTime = time.time()
	
	#we need to limit how often we spam the ADC because the clock app needs it for setting the display brightness, so if we've already done a bed check in the last second just return that val
	if(lastBedCheckTime > 0) :
		if(int(currentTime - lastBedCheckTime) < 1) :
			return inBedVal
		
	lastBedCheckTime = currentTime
	try :
		#pressure sensor is not exempt from debouncing
		rawVal1 = adc.readADCSingleEnded(0, gain, sps) / 1000
		time.sleep(1)
		rawVal2 = adc.readADCSingleEnded(0, gain, sps) / 1000
		inBedVal = (rawVal1 > 2) & (rawVal2 > 2) & (rawVal1 < 8) & (rawVal2 < 8)
	except :
		pass
			
	return inBedVal

def alarm_on(why):
	if(not pygame.mixer.get_busy()):
		log(why)
		bleep.set_volume(1)
		bleep.play(loops=-1)
		GPIO.output(23, 1)

def alarm_off(why):
	if(pygame.mixer.get_busy()):
		log(why)
		bleep.stop()
		bleep.set_volume(0)
		GPIO.output(23, 0)

def log(message):
	print datetime.datetime.today().strftime("%D %H:%M") + ": " + message

def run_alarm(alarm):
	log("Beginning alarm " + alarm.name)
	alarm.halted = 0
	snoozeCount = 0
	now = get_time()
	while(now < alarm.end and running):
		if not alarm.halted:
			if(inBed()):
				alarm_on("in bed during alarm period - start sound")
			else:
				alarm_off("out of bed - stop sound")
					
			if(GPIO.input(25)):
				if(alarm.buttonAction == "halt"):
					alarm.halted = 1
					alarm_off("alarm halted - stop sound")
				elif(snoozeCount < alarm.snoozesAllowed):
					alarm_off("snooze pressed - stop sound ")
					announcement = "The time is now " + datetime.datetime.today().strftime("%H:%M")
					command = "/usr/bin/testtts '" + announcement + "'  | aplay --rate=16000 --channels=1 --format=S16_LE"
					call(command, shell=True)  
					time.sleep(300)
					snoozeCount += 1
		now = get_time()
		
	alarm_off("end of alarm period " + alarm.name)
	log("Ending alarm " + alarm.name)

def run_alarms():
	global running
	
	log("main alarm loop starting")
	
	GPIO.setmode(GPIO.BCM)
	GPIO.setup(25, GPIO.IN)
	GPIO.setup(22, GPIO.OUT)
	GPIO.setup(23, GPIO.OUT)
	
	while(running):
		for alarm in alarms:
			now = get_time()
			if(now >= alarm.start and now < alarm.end and (datetime.datetime.today().strftime("%A") in alarm.daysOfWeek)):
				run_alarm(alarm)
					
		#Pressing the button outside an alarm time will be considered an alarm test
		time.sleep(1)
		if(GPIO.input(25)):
			alarm = alarmEntry()
			alarm.name = "test"
			alarm.start = time.localtime()
			alarm.end = time.localtime(time.time() + 60*60*2)
			alarm.snoozesAllowed = 0
			alarm.buttonAction = "halt"
			
			run_alarm(alarm)
		
	GPIO.cleanup()
	log("main alarm loop ending")
	
#init hardware
pygame.init()

#define signals
signal.signal(signal.SIGTERM, shutdown)
signal.signal(signal.SIGHUP, hup)
signal.signal(signal.SIGUSR1, print_schedule)

#define vars
alarms = []
inBedVal = 0
lastBedCheckTime = 0
gain = 4096  # +/- 4.096V
sps = 250  # 250 samples per second
ADS1015 = 0x00  # 12-bit ADC
adc = ADS1x15(ic=ADS1015)
bleep = pygame.mixer.Sound("/home/pi/pialarm/police_s.wav")
bleep.set_volume(0)

while(1):
	running = 1
	read_schedule()
	run_alarms()

log("oops, got to the exit")
