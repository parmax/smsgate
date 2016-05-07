This is a work in progress SMS gateway supporting Asterisk+chan_dongle AMI and serial dongle connections.

To run it you must have asterisk with chan_dongle listnening on an AMI port.
After that you create config.ini from a provided example.
do a go build && ./smsgate
