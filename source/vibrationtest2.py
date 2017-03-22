import time

import RPi.GPIO as GPIO
GPIO.setmode(GPIO.BOARD)
GPIO.setup(5, GPIO.OUT)
GPIO.setup(6, GPIO.OUT)

GPIO.output(5, GPIO.HIGH)
GPIO.outputs(6, GPIO.LOW)

time.sleep(2)

GPIO.output(5, GPIO.LOW)
GPIO.outputs(6, GPIO.LOW)
GPIO.cleanup()