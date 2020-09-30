import serial
import os, time
import RPi.GPIO as GPIO

GPIO.setmode(GPIO.BOARD)
port = serial.Serial("/dev/ttyS0", baudrate=115200, timeout=1)

port.write(b'AT\r')
rcv = port.read(10)
print(rcv)
time.sleep(1)

port.write(b"ATE0\r")
rcv = port.read(10)
print(rcv)
time.sleep(1)

port.write(b"AT+CMGF=1\r")
rcv = port.read(10)
print(rcv)
time.sleep(1)
print("Text Mode Enabled...")
time.sleep(1)

port.write(b'AT+CPMS="SM","SM","SM"\r')
rcv = port.read(30)
print(rcv)
time.sleep(1)

port.write(b'AT+CMGR=1\r')
rcv = port.read(100)
print(rcv)
time.sleep(1)

port.write(b'AT+CMGL="ALL"\r')
rcv = port.read(50000)
print(rcv)
time.sleep(1)
