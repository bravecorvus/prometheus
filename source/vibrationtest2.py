import time

import RPi.GPIO as GPIO
GPIO.setmode(GPIO.BOARD)
GPIO.setup(29, GPIO.OUT) #GPIO 5
GPIO.setup(31, GPIO.OUT) #GPIO 6

GPIO.output(29, GPIO.HIGH)
GPIO.output(31, GPIO.LOW)

time.sleep(2)

GPIO.output(29, GPIO.LOW)
GPIO.output(31, GPIO.LOW)
GPIO.cleanup()