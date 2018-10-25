import json
import sys
import os
import pickle
import sklearn
import random
import numpy
import socket
import threading
import argparse
from random import shuffle

import socketserver

from sklearn.neural_network import MLPClassifier
from sklearn.multioutput import MultiOutputClassifier
from sklearn.ensemble import RandomForestClassifier

from sklearn.ensemble import RandomForestClassifier,BaggingRegressor
from sklearn.tree import DecisionTreeRegressor
from sklearn.linear_model import LassoCV
from sklearn.feature_extraction import DictVectorizer
from sklearn.pipeline import make_pipeline
from sklearn.multioutput import MultiOutputRegressor
from sklearn.linear_model import *
from sklearn.tree import *
from sklearn.svm import *
from sklearn.kernel_ridge import *
from sklearn.neural_network import MLPClassifier
from sklearn.neighbors import KNeighborsClassifier
from sklearn.svm import SVC
from sklearn.gaussian_process import GaussianProcessClassifier
from sklearn.gaussian_process.kernels import RBF
from sklearn.tree import DecisionTreeClassifier
from sklearn.ensemble import *
from sklearn.neural_network import MLPRegressor
from sklearn.neighbors import KNeighborsRegressor
from sklearn.naive_bayes import GaussianNB
from sklearn.discriminant_analysis import QuadraticDiscriminantAnalysis


DEBUG = False

random.seed(123)

missingVal = -100

class Scikit(object):
    #data = []



    def __init__(self):
        self.size = 0
        self.data = []
        self.nameX = []
        self.trainX = numpy.array([])
        self.nameY = []
        self.nameY2 = []
        # trainY1 is used for classifiers --> each fp is a class
        self.trainY1 = []
        # trainY2 is used for regression --> each location is a class
        self.trainY2 = []
        self.macSet = set()
        self.locationSet = set()
        self.locationList = []
        self.clfClassifiers = []
        self.clfRegressors = []
        self.mainK = 10

        self.regressorsNames = [
            "BaggingRegression",
            "DecisionTreeRegression",
            "Lasso"
        ]
        self.regressors = [
            MultiOutputRegressor(BaggingRegressor(random_state=1)),
            MultiOutputRegressor(DecisionTreeRegressor(random_state=1)),
            MultiOutputRegressor(LassoCV())
        ]

        self.classifiersNames = [
            "scikitKNN",
            # "mlp",
            "rf",
        ]
        self.classifiers = [
            KNeighborsClassifier(n_neighbors=10,weights='distance',n_jobs=-1),
            # MLPClassifier(alpha=1),
            RandomForestClassifier(),
        ]

    def get_data(self, fname):
        # First go through once and get set of macs/locations
        X = []
        with open("../data/" + fname + ".scikit.json", 'r') as f_in:
            for fingerprint in f_in:
                try:
                    data = json.loads(fingerprint)
                except:
                    pass
                X.append(data)
                self.locationSet.add(data['location'])
                self.locationList.append(data['location'])
                for signal in data['wifi-fingerprint']:
                    self.macSet.add(signal['mac'])

        # print("macSet : ")
        # print(list(self.macSet))

        if DEBUG:
            print("Loaded %d fingerprints" % len(X))

        # Convert them to lists, for indexing
        self.nameX = list(self.macSet)
        self.nameY = list(self.locationSet)
        # print(self.locationList)
        # print(len(self.locationList))
        # print(len(self.locationList))

        # Go through the data again, in a random way
        # shuffle(X)
        # Split the dataset for training / learning
        trainSize = int(len(X))
        if DEBUG:
            print("Training size is %d fingerprints" % trainSize)
        # Initialize X, Y matricies for training and testing
        # self.trainX = numpy.full((trainSize, len(self.nameX)),missingVal)
        # self.testX = numpy.full((len(X) - trainSize, len(self.nameX)),missingVal)
        # # self.trainY = [0] * trainSize
        # # self.trainY = [[0,0] for i in range(trainSize)]
        # self.trainY = numpy.zeros(shape=(trainSize,2))
        # # self.testY = [0] * (len(X) - trainSize)
        # # self.testY = [[0,0] for i in range(len(X) - trainSize)]
        # self.testY = numpy.zeros(shape=(len(X) - trainSize,2))
        self.trainX = numpy.full((trainSize, len(self.nameX)),missingVal)
        self.trainY1 = [0] * trainSize
        self.trainY2 = numpy.zeros(shape=(trainSize,2))
        curRowTrain = 0
        curRowTest = 0

        for i in range(len(X)):
            newRow = numpy.full(len(self.nameX),missingVal)
            newXY = numpy.zeros(2)
            for signal in X[i]['wifi-fingerprint']:
                newRow[self.nameX.index(signal['mac'])] = signal['rssi']
            # if i < trainSize:  # do training
            self.trainX[curRowTrain, :] = newRow
            xyList = X[i]['location'].split(",")
            self.trainY2[curRowTrain] = numpy.asarray(xyList, dtype=numpy.float32)
            # self.trainY[curRowTrain, :] = X[i]['location']
            self.trainY1[curRowTrain] = curRowTrain
            #self.trainY2[curRowTrain] = self.nameY.index(X[i]['location'])
            curRowTrain = curRowTrain + 1
            # else:
            #     self.testX[curRowTest, :] = newRow
            #     # self.testY[curRowTest] = self.nameY.index(X[i]['location'])
            #     #xyList = X[i]['location'].split(",")
            #     #self.testY[curRowTest] = numpy.asarray(xyList, dtype=numpy.float32)
            #     self.testY1[curRowTest] = curRowTest
            #     self.testY2[curRowTest] = self.nameY.index(X[i]['location'])
            #     curRowTest = curRowTest + 1
            # print(self.trainX)
            # print(self.trainY)
            # print(self.testX)
            # print(self.testY)
            # print(self.nameX)
            # print(len(self.trainY1))
    def learn(self, dataFile):
        print("Learning...")
        self.get_data(dataFile)
        # if DEBUG:
        # print(self.trainY)
        # for name, clf in zip(self.names, self.classifiers):
        #     try:
        #         clf.fit(self.trainX, self.trainY1)
        #         # score = clf.score(self.testX, self.testY1)
        #         # print(name, score)
        #     except Exception as ex:
        #         print("ERROR:",ex)
        for name, clf in zip(self.classifiersNames, self.classifiers):
            try:
                self.clfClassifiers.append(clf)
                print(name)
                print(clf)
            except Exception as ex:
                print(ex)
        for name, clf in zip(self.regressorsNames, self.regressors):
            try:
                self.clfRegressors.append(clf)
                print(name)
                print(clf)
            except Exception as ex:
                print(ex)
        # for max_feature in ["auto","log2",None,"sqrt"]:
        # 	for n_estimator in range(1,30,1):
        # 		for min_samples_split in range(2,10):
        # 			clf = RandomForestClassifier(n_estimators=n_estimator,
        # 				max_features=max_feature,
        # 				max_depth=None,
        # 				min_samples_split=min_samples_split,
        # 				random_state=0)
        # 			clf.fit(self.trainX, self.trainY)
        # 			print(max_feature,n_estimator,min_samples_split,clf.score(self.testX, self.testY))

        # clf = RandomForestClassifier(
        #     n_estimators=10,
        #     max_depth=None,
        #     min_samples_split=2,
        #     random_state=0)

        # self.clf = MultiOutputRegressor(BaggingRegressor(base_estimator=SVC(probability=True, kernel='linear'),random_state=1))
        # adaboost and baggingresgressor that are regression and classifier are good too!
        # self.clf = MultiOutputRegressor(BaggingRegressor(base_estimator=SVC(probability=True, kernel='linear'),random_state=1))
        # self.clf.append(MultiOutputRegressor(BaggingRegressor(random_state=1)))
        # self.clf.append(MultiOutputRegressor(BaggingRegressor(random_state=1)))
        # self.clf.append(MultiOutputRegressor(BaggingRegressor(base_estimator=SVC(probability=True, kernel='linear'),random_state=1)))

        # print(self.clf)
        print("BBBBBBBBBBBBBBBBBBBBBBBBBBBBB")
        # print(self.trainX)
        # print(self.trainY)

        print("FFFFFFFFFFFFFFFFFFFFFFFFFFFFF")
        try:
            for i in range(len(self.clfClassifiers)):
                self.clfClassifiers[i].fit(self.trainX, self.trainY1)
            for i in range(len(self.clfRegressors)):
                self.clfRegressors[i].fit(self.trainX, self.trainY2)
        except Exception as ex:
            print(ex)
            # print("error in fitting")
        #     print("Trainx : ")
        #     print(self.trainX)
        #     print("TrainY : ")
        #     print(self.trainY)
        #
        # # self.clf.fit(self.trainX, self.trainY)
        # # print(test)
        # print(self.nameX)
        # print(self.nameY)
        # score = {}
        # for i in range(len(self.clf)):
        #     score[self.names[i]] = self.clf[i].score(self.testX, self.testY1)
        # score["bagging"] = self.clf[0].score(self.testX, self.testY)
        # score["svc"] =self.clf[1].score(self.testX, self.testY)
        # score = self.clf.score(self.testX, self.testY)
        # print("score")
        # print(score)
        with open('../data/' + dataFile + '.scikit.pkl', 'wb') as fid:
            pickle.dump([self.clfClassifiers,self.clfRegressors, self.nameX, self.nameY, self.locationList], fid)
        return {'learn': 1}


    def predictions2XY(self, prdWithName):

        prdWithName = sorted(prdWithName.items(), key=lambda x: x[1], reverse=True)
        print(prdWithName)
        i=0
        wigthSum = 0
        xyResult = [0,0]
        for keyval in prdWithName:
            i+=1
            if(i>self.mainK):
                break
            xy = keyval[0][0].split(",")
            wigth = keyval[1]
            wigthSum += wigth
            xyResult[0] += float(xy[0])*wigth
            xyResult[1] += float(xy[1])*wigth
        xyResult[0] = xyResult[0] / wigthSum
        xyResult[1] = xyResult[1] / wigthSum
        return str(xyResult[0])+","+str(xyResult[1])

    def classify(self, groupName, fingerpintFile):
        print("Classifing...")
        with open('../data/' + groupName + '.scikit.pkl', 'rb') as pickle_file:
            [self.clfClassifiers,self.clfRegressors, self.nameX, self.nameY, self.locationList] = pickle.load(pickle_file)


        # print(self.nameX)
        # print(self.nameY)
        # As before, we need a row that defines the macs
        newRow = numpy.full(len(self.nameX),missingVal)
        data = {}
        with open(fingerpintFile, 'r') as f_in:
            for line in f_in:
                data = json.loads(line)
        if len(data) == 0:
            return
        for signal in data['wifi-fingerprint']:
            # print(signal)
            # Only add the mac if it exists in the learning model
            if signal['mac'] in self.nameX:
                newRow[self.nameX.index(signal['mac'])] = signal['rssi']
        #Notice: missing mac must be handled
        # prediction = clf.predict_proba(newRow.reshape(1, -1))
        # print(newRow)
        if(len(newRow) == 0):
            return {}
        # print(newRow.reshape(1, -1))
        print("Classifiers:")
        print(self.clfClassifiers)
        print("Regressors: ")
        print(self.clfRegressors)
        prediction = []

        #Add classifier results
        for i in range(len(self.clfClassifiers)):
            prdWithName = {}
            prdWithNameSorted= {}
            # prediction.append(self.clf[i].predict(newRow.reshape(1, -1)))
            prd = self.clfClassifiers[i].predict_proba(newRow.reshape(1, -1))
            print(prd)
            for i in range(len(prd[0])):
                # prdWithName[self.nameY[i]] = prd[0][i]
                prdWithName[(self.locationList[i],i)] = prd[0][i]
            #print(prdWithName)
            prediction.append(self.predictions2XY(prdWithName))

            # print(self.predictions2XY(prdWithName))
            # print(sorted(prdWithName.items(), key=lambda x: x[1], reverse=True))
            # prediction.append(prdWithName)
        # prediction.append(self.clf[0].predict(newRow.reshape(1, -1)))
        # prediction.append(self.clf[1].predict(newRow.reshape(1, -1)))
        print("Prediction: ",prediction)

        # predictStr = str(prediction[0][0])+","+str(prediction[0][1])
        predictStr = []
        for i in range(len(prediction)):
            # predictStr.append(str(prediction[i][0][0])+","+str(prediction[i][0][1]))
            predictStr.append(str(prediction[i][0]))
        # predictStr1 = str(prediction[0][0][0])+","+str(prediction[0][0][1])
        # predictStr2 = str(prediction[1][0][0])+","+str(prediction[1][0][1])

        predictionJson = {}
        # for i in range(len(prediction[0])):
        #     predictionJson[self.nameY[i]] = prediction[0][i]

        for i in range(len(prediction)):
            predictionJson[self.classifiersNames[i]]=prediction[i]

        #Add regressions results
        predictionRg = []
        for i in range(len(self.clfRegressors)):
            predictionRg.append(self.clfRegressors[i].predict(newRow.reshape(1, -1)))

        predictStr = []
        for i in range(len(predictionRg)):
            predictStr.append(str(predictionRg[i][0][0])+","+str(predictionRg[i][0][1]))

        for i in range(len(predictStr)):
            predictionJson[self.regressorsNames[i]]=predictStr[i]
        # predictionJson["bagging"]=predictStr1
        # predictionJson["svc"]=predictStr2

        return predictionJson


class EchoRequestHandler(socketserver.BaseRequestHandler):

    def handle(self):
        # Echo the back to the client
        data = self.request.recv(1024)
        data = data.decode('utf-8').strip()
        print("received data:'%s'" % data)
        group = data.split('=')[0].strip()
        filename = data.split('=')[1].strip()
        payload = "error".encode('utf-8')
        if len(group) == 0:
            self.request.send(payload)
            return
        randomF = Scikit()
        print("fileName:",filename)
        print("group:",group)
        if len(filename) == 0:
            # payload = json.dumps(randomF.learn(group, 1)).encode('utf-8')
            # print("filename length is zero!")
            payload = json.dumps(randomF.learn(group)).encode('utf-8')
        else:
            # print("filename length isn't zero!")
            payload = json.dumps(
                randomF.classify(
                    group,
                    "../data/" + filename +
                    ".scikittemp")).encode('utf-8')
            print("Payload: ",payload)
        self.request.send(payload)
        return

if __name__ == '__main__':
    parser = argparse.ArgumentParser()
    parser.add_argument(
        "-p",
        "--port",
        type=int,
        help="select the port to run on")
    parser.add_argument("-g", "--group", type=str, help="select a group")
    parser.add_argument(
        "-f",
        "--file",
        type=str,
        help="select a file with fingerprints")
    parser.add_argument("-d", "--debug", help="debug mode")
    args = parser.parse_args()
    DEBUG = args.debug
    if args.port is not None:
        socketserver.TCPServer.allow_reuse_address = True
        address = ('localhost', args.port)  # let the kernel give us a port
        server = socketserver.TCPServer(address, EchoRequestHandler)
        ip, port = server.server_address  # find out what port we were given
        server.serve_forever()
    elif args.file is not None and args.group is not None:
        randomF = Scikit()
        print(randomF.classify(args.group, args.file))
    elif args.group is not None:
        randomF = Scikit()
        print(randomF.learn(args.group))
        # print(randomF.learn(args.group, 0.5))
    else:
        print("""Usage:
To just run as TCP server:
	python3 scikit.py --port 5009
To just learn:
	python3 scikit.py --group GROUP
To classify
	python3 scikit.py --group GROUP --file FILEWITHFINGERPRINTS
""")

