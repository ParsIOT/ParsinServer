import time
import serial
import signal 

resultFileName = "result.log"

open(resultFileName, 'w').close()


class SIGINT_handler():
    def __init__(self):
        self.SIGINT = False
    def signal_handler(self, signal, frame):
        print('Closed!')
        ser.write(str.encode('quit\r'))
        out = ''
        while ser.inWaiting() > 0:
            out += ser.read(1).decode()

        if out != '':
            print(out)
        time.sleep(1)
        ser.close()
        f.close()
        time.sleep(1)
        exit()
        self.SIGINT = True


handler = SIGINT_handler()
signal.signal(signal.SIGINT, handler.signal_handler)

# configure the serial connections (the parameters differs on the device you are connecting to)
ser = serial.Serial(
    port='/dev/ttyACM0',
    baudrate=115200
    # parity=serial.PARITY_NONE,
    # stopbits=serial.STOPBITS_ONE,
    # bytesize=serial.EIGHTBITS
)

ser.isOpen()

print('Enter your commands below.\r\nInsert "exit" to leave the application.')

ser.write(str.encode('\r\r'))
out = ''
while ser.inWaiting() > 0:
    out += ser.read(1).decode()

if out != '':
    print(out)

time.sleep(1)
inputCMD=1
ser.write(str.encode('lep\r'))

locations = []
writeCount = 0
while 1:
    writeCount += 1
    # send the character to the device
    # (note that I happend a \r\n carriage return and line feed to the characters - this is requested by my device)
    out = ''
    # let's wait one second before reading output (let's give device time to answer)
    time.sleep(0.2)
    timestamp = int(time.time()*10**3)
    # print(timestamp)
    while ser.inWaiting() > 0:
        out += ser.read(1).decode()
    
    if out != '':
        # print(out)
        out = out.strip()
        if out[:3]=="POS":
            outList = out.split(",")
            tag,x,y,z = str(outList[2]),float(outList[3])*100,float(outList[4])*100,float(outList[5])*100
            print(timestamp,tag,x,y,z)
            locations.append([timestamp,tag,x,y,z])
    else:
        print(timestamp,",",",",",")
        locations.append([timestamp,None,None,None,None])
    if (writeCount==1):
        writeCount = 0
        f = open(resultFileName, "a")
        for loc in locations:
            f.write(str(loc)[1:-1]+"\r\n")
        f.close()
        locations = []
   
    print(locations)
