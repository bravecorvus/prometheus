import time

import RPi.GPIO as GPIO
GPIO.setmode(GPIO.BCM)
GPIO.setup(17, GPIO.OUT) #GPIO 17 EN 1, 2
GPIO.setup(5, GPIO.OUT) #GPIO 5 INPUT 1
GPIO.setup(6, GPIO.OUT) #GPIO 6 INPUT 2

#GPIO.output(5, GPIO.HIGH)
#GPIO.output(6, GPIO.LOW)
#GPIO.output(17, GPIO.HIGH)
GPIO.output(5, GPIO.LOW)
GPIO.output(6, GPIO.HIGH)
GPIO.output(17, GPIO.HIGH)
time.sleep(5)

GPIO.output(17, GPIO.LOW)
#GPIO.output(5, GPIO.LOW)
#GPIO.output(6, GPIO.LOW)
GPIO.cleanup()
