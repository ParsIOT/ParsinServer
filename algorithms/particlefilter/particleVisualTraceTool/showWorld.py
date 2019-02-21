from PyQt5.QtWidgets import *
from PyQt5.QtGui import *
from PyQt5.QtCore import *
import sys, random
import math
import numpy
import pickle

class Example(QWidget):

    def __init__(self,data,scaleXY, scaleScreen, bg):
        super().__init__()
        self.particles = None
        self.mean = None
        self.data = data
        self.dataLength = len(data)
        self.scaleXY = scaleXY
        self.scaleScreen = scaleScreen
        self.bg = QPixmap(bgPath)
        self.qp = None
        self.LastLine = 4
        self.width = 0
        self.height = 0
        self.initUI()


    def initUI(self):
        self.width = self.bg.width()*self.scaleScreen
        self.height = self.bg.height()*self.scaleScreen
        self.setFixedSize(self.width, self.height)

        # self.setRandomParticle()
        centerPoint = QDesktopWidget().availableGeometry().center()
        self.setGeometry(centerPoint.x()-self.width /2, centerPoint.y()-self.height/2, self.width ,  self.height)
        self.setWindowTitle('Points')
        self.show()

    def paintEvent(self, e):
        qp = QPainter()
        self.qp = qp
        self.setFirstData()

        qp.begin(self)
        self.drawBackgroundImage()
        self.drawParticles()
        self.drawMean()
        qp.end()

    def drawBackgroundImage(self):

        self.qp.drawPixmap(self.rect(), self.bg)
        self.qp.translate(self.width /2 , self.height/2)
        self.qp.scale(1.0, -1.0)

    # def setRandomParticle(self):
    #     size = self.size()
    #
    #     nums = 100
    #     self.particles = numpy.zeros(shape=(nums, 3), dtype=float)
    #     for i in range(nums):
    #         x = random.randint(-size.width()/2, size.width()/2)
    #         # x = random.randint(1, size.width()-1)
    #         y = random.randint(-size.height()/2, size.height()/2)
    #         # y = random.randint(1, size.height()-1)
    #         h = random.randint(0, 360)
    #         self.particles[i] = [x,y,h]

    def setFirstData(self):
        fline = data[5]
        timestamp = fline[0]
        self.mean = fline[1]
        self.particles = fline[2]


    def drawParticles(self):
        if len(self.particles) != 0:
            for particle in self.particles:
                self.drawPie(particle,20,Qt.red)


    def drawPie(self, xyh, size, color=Qt.black):
        x = int(xyh[0] * self.scaleXY)
        y = int(xyh[1] * self.scaleXY)
        h = int(xyh[2])
        rectangle = QRect(x - size / 2, y - size / 2, size, size)
        # print(xyh[0])
        # rectangle = QRect(x, 100 , size, size)

        arrowH = (h + 180) % 360
        startAngle = (arrowH - size/2) * 16
        spanAngle = size * 16
        brush = QBrush(Qt.SolidPattern)
        brush.setColor(color)
        self.qp.setPen(color)
        self.qp.setBrush(brush)
        self.qp.drawPie(rectangle, startAngle, spanAngle)

    # def drawParticle(self, qp, particle):
    #     x = particle[0]
    #     y = particle[1]
    #     h = particle[2]
    #
    #     #Drawing
    #     size = 12
    #     # dx = math.cos(math.radians(h)) * size / 4
    #     # dy = math.sin(math.radians(h)) * size / 4
    #     rectangle = QRect(x - size / 2, y - size / 2, size, size)
    #
    #     h1 = (h+180)%360
    #     startAngle = (h1-10) * 16
    #     spanAngle = 20 * 16
    #     brush = QBrush(Qt.SolidPattern)
    #     brush.setColor(Qt.red)
    #     qp.setPen(Qt.red)
    #     qp.setBrush(brush)
    #     qp.drawPie(rectangle, startAngle, spanAngle)
    #
    # def drawRect(self,qp):
    #     rectangle = QRect(0, 0, 10, 50)
    #     qp.drawRect(rectangle)

    def drawMean(self):
        if len(self.particles) != 0:
            self.drawPie(self.mean,40,Qt.blue)

    # def drawMean(self, qp):
    #     if self.mean == None:
    #         return
    #
    #     x = self.mean[0]
    #     y = self.mean[1]
    #     h = self.mean[2]
    #
    #     #Drawing
    #     size = 40
    #     # dx = math.cos(math.radians(h)) * size /4
    #     # dy = math.sin(math.radians(h)) * size /4
    #     rectangle = QRect(x-size/2, y-size/2, size, size)
    #
    #     h1 = (h + 180) % 360
    #     startAngle = (h1 - 20) * 16
    #     spanAngle = 40 * 16
    #     brush = QBrush(Qt.SolidPattern)
    #     brush.setColor(Qt.blue)
    #     qp.setPen(Qt.blue)
    #     qp.setBrush(brush)
    #     qp.drawPie(rectangle, startAngle, spanAngle)

    def keyPressEvent(self, event):
        if event.key() == Qt.Key_Right:
            # self.moveParticles("right")
            # self.moveParticles("right")
            self.nextLine()
        elif event.key() == Qt.Key_Left:
            # self.moveParticles("left")
            self.previousLine()

        event.accept()

    # def moveParticles(self, heading):
    #     if heading == "right":
    #         dx = 4
    #     elif heading == "left":
    #         dx = -4
    #     for particle in self.particles:
    #         particle[0] += dx
    #     self.update()

    def nextLine(self):
        if self.LastLine >= self.dataLength - 1:
            return
        self.LastLine += 1
        data = self.data[self.LastLine]
        timestamp = data[0]
        mean = data[1]
        particles = data[2]
        for i in range(len(self.particles)):
            self.particles[i] = particles[i]
        for i in range(len(mean)):
            self.mean[i] = mean[i]

        self.update()

    def previousLine(self):
        if self.LastLine == 0:
            return
        self.LastLine -= 1

        data = self.data[self.LastLine]
        timestamp = data[0]
        mean = data[1]
        particles = data[2]
        for i in range(len(self.particles)):
            self.particles[i] = particles[i]
        for i in range(len(mean)):
            self.mean[i] = mean[i]

        self.update()

def ReadData(pklFilePath):
    import pickle

    resultDataFileName = pklFilePath

    lines = []
    with open(resultDataFileName, 'rb', 1) as f:
        # lines = f.readlines()
        while 1:
            try:
                obj = pickle.load(f)
                lines.append(obj)
            except:
                break

    print(len(lines))
    # print(lines)
    return lines





if __name__ == '__main__':
    scaleXY = 0.33
    scaleScreen = scaleXY / 0.25
    bgPath = "./ArmanExactMap.png"
    pklFilePath =  "../results.pkl"

    data = ReadData(pklFilePath)


    app = QApplication(sys.argv)
    ex = Example(data, scaleXY, scaleScreen, bgPath)
    sys.exit(app.exec_())
