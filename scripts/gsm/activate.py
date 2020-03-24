import serial
import RPi.GPIO as GPIO
import os, time

GPIO.setmode(GPIO.BOARD)
GPIO.setup(11, GPIO.OUT)

GPIO.output(11,1)
time.sleep(3)
GPIO.output(11,0)
time.sleep(3)
