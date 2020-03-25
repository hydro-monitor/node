import serial
import os, time

# Enable Serial Communication
port = serial.Serial("/dev/ttyS0", baudrate=115200, timeout=1)

# Transmitting AT Commands to the Modem
# '\r\n' indicates the Enter key

port.write('AT+CSQ'+'\r\n')
rcv = port.read(100)
print rcv
