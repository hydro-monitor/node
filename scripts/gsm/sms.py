import serial
import RPi.GPIO as GPIO
import os, time

GPIO.setmode(GPIO.BOARD)
GPIO.setup(11, GPIO.OUT)

GPIO.output(11,1)
time.sleep(3)
GPIO.output(11,0)
time.sleep(3)

# Enable Serial Communication

port = serial.Serial("/dev/ttyS0", baudrate=19200, timeout=1)

print "Arranca"

# Transmitting AT Commands to the Modem
# '\r\n' indicates the Enter key

port.write('AT'+'\r\n')
print "write AT"
rcv = port.read(10)
print "read"
print rcv
time.sleep(1)


port.write('ATE0'+'\r\n')      # Disable the Echo
rcv = port.read(10)
print rcv
time.sleep(1)


port.write('AT+CMGF=1'+'\r\n')  # Select Message format as Text mode
rcv = port.read(10)
print rcv
time.sleep(1)

#port.write('AT+CNMI=2,1,0,0,0'+'\r\n')   # New SMS Message Indications
#rcv = port.read(10)
#print rcv
#time.sleep(1)

# Sending a message to a particular Number

port.write('AT+CMGS="+5491161004695"'+'\r\n')
rcv = port.read(10)
print rcv
time.sleep(1)

port.write('Hello User'+'\r\n')  # Message
rcv = port.read(10)
print rcv

port.write("\x1A") # Enable to send SMS
for i in range(10):
    rcv = port.read(10)
    print rcv
