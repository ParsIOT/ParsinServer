from PyQt5.QtWidgets import *
from PyQt5.QtGui import *
from PyQt5.QtCore import *
import sys, random
import math
import numpy
import pickle
import os


class MainWindow(QWidget):

    def __init__(self, data, scaleXY, scaleScreen, bgPath, ShowTrueAndEst):
        super().__init__()
        self.infLocation = -10000000000
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
        self.ShowTrueAndEst = ShowTrueAndEst
        self.trueLocAndEst = numpy.zeros((3, 3))
        self.trueLocAndEst[:, 1:].fill(self.infLocation)

        self.particleColor = Qt.red
        self.meanColor = Qt.blue
        self.trueLocColor = Qt.black
        self.masterEstColor = Qt.darkMagenta
        self.slaveEstColor = Qt.yellow

        self.updateIndicatortate = 0
        self.initUI()


    def initUI(self):
        self.width = self.bg.width()*self.scaleScreen
        self.height = self.bg.height()*self.scaleScreen
        self.setFixedSize(self.width, self.height)
        self.infLocation = 2 * max(self.width, self.height)

        # self.setRandomParticle()
        centerPoint = QDesktopWidget().availableGeometry().center()
        # self.setGeometry(centerPoint.x()-self.width /2, centerPoint.y()-self.height/2, self.width ,  self.height)
        self.setGeometry(centerPoint.x(), centerPoint.y() - self.height / 2, self.width, self.height)
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
        self.drawUpdateIndicator()
        if self.ShowTrueAndEst:
            self.drawTrueAndEst()
        qp.end()

    def drawBackgroundImage(self):

        self.qp.drawPixmap(self.rect(), self.bg)
        self.qp.translate(self.width /2 , self.height/2)
        self.qp.scale(1.0, -1.0)



    def setFirstData(self):
        fline = data[5]
        timestamp = fline[0]
        self.mean = fline[1]
        self.particles = fline[2]

    def drawUpdateIndicator(self):
        indicatorSize = 10
        indicatorCoordinate = QPoint(indicatorSize - self.width / 2, -indicatorSize + self.height / 2)
        brush = QBrush(Qt.SolidPattern)
        if self.updateIndicatortate == 0:
            brush.setColor(Qt.gray)
            self.qp.setPen(Qt.gray)
            self.qp.setBrush(brush)
            self.qp.drawEllipse(indicatorCoordinate, indicatorSize, indicatorSize)
        elif self.updateIndicatortate == 1:
            brush.setColor(Qt.green)
            self.qp.setPen(Qt.green)
            self.qp.setBrush(brush)
            self.qp.drawEllipse(indicatorCoordinate, indicatorSize, indicatorSize)

    def drawPie(self, xyh, size, color=Qt.cyan):
        x = int(xyh[0] * self.scaleXY)
        y = int(xyh[1] * self.scaleXY)
        h = int(xyh[2]) % 360
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

    def drawParticles(self):
        if len(self.particles) != 0:
            for particle in self.particles:
                self.drawPie(particle, 20, self.particleColor)

    def drawMean(self):
        if len(self.mean) != 0:
            # self.drawPie(self.mean, 40, self.meanColor)
            print("Mean")
            print(self.mean)
            brush = QBrush(Qt.SolidPattern)
            brush.setColor(self.meanColor)
            self.qp.setPen(self.meanColor)
            self.qp.setBrush(brush)
            meanDot = QPoint(self.mean[0] * self.scaleXY, self.mean[1] * self.scaleXY)
            self.qp.drawEllipse(meanDot, 8, 8)

    def drawTrueAndEst(self):
        # print(self.trueLocAndEst)
        if len(self.trueLocAndEst) != 0:
            print("EEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEEst")
            print(self.trueLocAndEst[2])
            trueLocCenter = QPoint(self.trueLocAndEst[2][1] * self.scaleXY, self.trueLocAndEst[2][2] * self.scaleXY)
            brush = QBrush(Qt.SolidPattern)
            brush.setColor(self.trueLocColor)
            self.qp.setPen(self.trueLocColor)
            self.qp.setBrush(brush)
            self.qp.drawEllipse(trueLocCenter, 5, 5)

            if (self.trueLocAndEst[0][0]):
                masterEstCenter = QPoint(self.trueLocAndEst[0][1] * self.scaleXY,
                                         self.trueLocAndEst[0][2] * self.scaleXY)
                brush = QBrush(Qt.SolidPattern)
                brush.setColor(self.masterEstColor)
                self.qp.setPen(self.masterEstColor)
                self.qp.setBrush(brush)
                self.qp.drawEllipse(masterEstCenter, 7, 7)
            if (self.trueLocAndEst[1][0]):
                slaveEstCenter = QPoint(self.trueLocAndEst[1][1] * self.scaleXY,
                                        self.trueLocAndEst[1][2] * self.scaleXY)
                brush = QBrush(Qt.SolidPattern)
                brush.setColor(self.slaveEstColor)
                self.qp.setPen(self.slaveEstColor)
                self.qp.setBrush(brush)
                self.qp.drawEllipse(slaveEstCenter, 6, 6)

    def keyPressEvent(self, event):
        if event.key() == Qt.Key_Right:
            self.changeState("next")
        elif event.key() == Qt.Key_Left:
            self.changeState("back")

        event.accept()

    # def moveParticles(self, heading):
    #     if heading == "right":
    #         dx = 4
    #     elif heading == "left":
    #         dx = -4
    #     for particle in self.particles:
    #         particle[0] += dx
    #     self.update()

    def changeState(self, direction="next"):
        if direction == "next":
            if self.LastLine >= self.dataLength - 1:
                return
            self.LastLine += 1
        elif direction == "back":
            if self.LastLine == 0:
                return
            self.LastLine -= 1
        data = self.data[self.LastLine]
        timestamp = data[0]
        mean = data[1]
        particles = data[2]
        trueLocAndEst = data[4]
        print("#################")
        print(mean)
        print(trueLocAndEst)
        # print(particles)
        # print(data[3])
        # print(numpy.sort(data[3]))

        for i in range(len(self.particles)):
            self.particles[i] = particles[i]
        for i in range(len(mean)):
            self.mean[i] = mean[i]


        self.updateIndicatortate = 0
        self.trueLocAndEst[:, 1:].fill(self.infLocation)

        # set truel location and ble estimation
        if len(trueLocAndEst) > 0:
            for i in range(len(trueLocAndEst)):
                for j in range(len(trueLocAndEst[0])):
                    self.trueLocAndEst[i][j] = trueLocAndEst[i][j]

            if trueLocAndEst[0][0] == 1 or trueLocAndEst[1][0] == 1:
                self.updateIndicatortate = 1

        self.update()

    # def previousLine(self):
    #
    #
    #     data = self.data[self.LastLine]
    #     timestamp = data[0]
    #     mean = data[1]
    #     particles = data[2]
    #     trueLocAndBLE = data[4]
    #
    #     for i in range(len(self.particles)):
    #         self.particles[i] = particles[i]
    #     for i in range(len(mean)):
    #         self.mean[i] = mean[i]
    #
    #     # set truel location and ble estimation
    #     if len(trueLocAndBLE) == 0:
    #         for i in range(0, 4):
    #             self.trueLocAndBLE[i] = self.infLocation
    #     elif len(trueLocAndBLE) == 2:
    #         for i in range(2, 4):
    #             self.trueLocAndBLE[i] = self.infLocation
    #     else:
    #         for i in range(0, 4):
    #             self.trueLocAndBLE[i] = trueLocAndBLE[i]
    #
    #     # toggle indicator
    #     self.updateIndicatortate = 0
    #     if len(trueLocAndBLE) == 4:
    #         self.updateIndicatortate = 1
    #
    #     self.update()

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
    ShowTrueAndEst = True
    scaleScreen = scaleXY / 0.25
    bgPath = "./ArmanExactMap.png"
    pklFilePath =  "../results.pkl"

    if (len(sys.argv) == 2):
        newPklFilePath = sys.argv[1]
        if os.path.isfile(newPklFilePath):
            pklFilePath = newPklFilePath
        else:
            print("Err:", newPklFilePath, " doesn't exist")
            exit(0)

    data = ReadData(pklFilePath)


    app = QApplication(sys.argv)
    ex = MainWindow(data, scaleXY, scaleScreen, bgPath, ShowTrueAndEst)
    sys.exit(app.exec_())
