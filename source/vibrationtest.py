import pygame
import time
import datetime
from pygame.locals import *
import RPi.GPIO as GPIO
from Adafruit_ADS1x15 import ADS1x15
import signal
import sys
from subprocess import call

GPIO.setmode(GPIO.BCM)
GPIO.setup(25, GPIO.IN)
GPIO.setup(22, GPIO.OUT)
GPIO.setup(23, GPIO.OUT)

# signal.signal(signal.SIGTERM, shutdown)
# signal.signal(signal.SIGHUP, hup)
# signal.signal(signal.SIGUSR1, print_schedule)

# alarms = []
# inBedVal = 0
# lastBedCheckTime = 0

GPIO.output(23, 1)