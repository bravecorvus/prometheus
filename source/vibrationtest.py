import pygame
import time
import datetime
from pygame.locals import *
import RPi.GPIO as GPIO
import signal
import sys
from subprocess import call

GPIO.setmode(GPIO.BCM)
# GPIO.setup(25, GPIO.IN)
GPIO.setup(22, GPIO.OUT)
GPIO.setup(23, GPIO.OUT)

while True:
    userinput = input("on or off\n")
    if userinput == "on":
        GPIO.output(22, 1)
    elif userinput == "off":
        GPIO.output(22, 0)
    else:
        break

GPIO.cleanup()