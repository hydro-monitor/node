from picamera import PiCamera
from time import sleep
from datetime import datetime
import os

PICTURES_DIR = "/mnt/usb/"

camera = PiCamera()

while True:
	if os.path.ismount(PICTURES_DIR):
		name = '{0:%Y%m%d_%H%M%S}'.format(datetime.now())
		print("Taking picture: %s" % name)
		camera.capture("%s%s.jpg" % (PICTURES_DIR, name))
	sleep(30)
	#for mode in camera.EXPOSURE_MODES:
	#	camera.exposure_mode = mode
	#	camera.capture("/home/pi/camera/%s.jpg" % mode)