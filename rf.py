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
from sklearn.ensemble import RandomForestClassifier,BaggingRegressor
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

missingVal = -1000

class RF(object):
    #data = []

    def __init__(self):
        self.size = 0
        self.data = []
        self.nameX = []
        self.trainX = numpy.array([])
        self.testX = numpy.array([])
        self.nameY = []
        self.trainY = []
        self.testY = []
        self.macSet = set()
        self.locationSet = set()
        self.clf = None

    def get_data(self, fname, splitRatio):
        # First go through once and get set of macs/locations
        X = []
        with open("data/" + fname + ".rf.json", 'r') as f_in:
            for fingerprint in f_in:
                try:
                    data = json.loads(fingerprint)
                except:
                    pass
                X.append(data)
                self.locationSet.add(data['location'])
                for signal in data['wifi-fingerprint']:
                    self.macSet.add(signal['mac'])

        if DEBUG:
            print("Loaded %d fingerprints" % len(X))

        # Convert them to lists, for indexing
        self.nameX = list(self.macSet)
        self.nameY = list(self.locationSet)

        # Go through the data again, in a random way
        # shuffle(X)
        # Split the dataset for training / learning
        trainSize = int(len(X) * splitRatio)
        if DEBUG:
            print("Training size is %d fingerprints" % trainSize)
        # Initialize X, Y matricies for training and testing
        self.trainX = numpy.full((trainSize, len(self.nameX)),missingVal)
        self.testX = numpy.full((len(X) - trainSize, len(self.nameX)),missingVal)
        # self.trainY = [0] * trainSize
        # self.trainY = [[0,0] for i in range(trainSize)]
        self.trainY = numpy.zeros(shape=(trainSize,2))
        # self.testY = [0] * (len(X) - trainSize)
        # self.testY = [[0,0] for i in range(len(X) - trainSize)]
        self.testY = numpy.zeros(shape=(len(X) - trainSize,2))

        curRowTrain = 0
        curRowTest = 0

        for i in range(len(X)):
            newRow = numpy.full(len(self.nameX),missingVal)
            newXY = numpy.zeros(2)
            for signal in X[i]['wifi-fingerprint']:
                newRow[self.nameX.index(signal['mac'])] = signal['rssi']
            if i < trainSize:  # do training
                self.trainX[curRowTrain, :] = newRow
                xyList = X[i]['location'].split(",")
                self.trainY[curRowTrain] = numpy.asarray(xyList, dtype=numpy.float32) 
                # self.trainY[curRowTrain, :] = X[i]['location']
                curRowTrain = curRowTrain + 1
            else:
                self.testX[curRowTest, :] = newRow
                # self.testY[curRowTest] = self.nameY.index(X[i]['location'])
                xyList = X[i]['location'].split(",")
                self.testY[curRowTest] = numpy.asarray(xyList, dtype=numpy.float32) 
                curRowTest = curRowTest + 1
        # print(self.trainX)
        # print(self.trainY)
        # print(self.testX)
        # print(self.testY)
        # print(self.nameX)

    def learn(self, dataFile, splitRatio):
        print("Learning...")
        self.get_data(dataFile, splitRatio)
        if DEBUG:
            names = [
                # "Nearest Neighbors",
                # "Linear SVM",
                # "RBF SVM",
                # "Gaussian Process",
                # "Decision Tree",
                # "Random Forest",
                # "Neural Net",
                # "AdaBoost",
                # "Naive Bayes",
                # "QDA",
                "BaggingRegressor"]
            classifiers = [
                # KNeighborsClassifier(3),
                # SVC(kernel="linear", C=0.025),
                # SVC(gamma=2, C=1),
                # GaussianProcessClassifier(1.0 * RBF(1.0), warm_start=True),
                # RandomForestClassifier(max_depth=5, n_estimators=10, max_features=1),
                # MLPClassifier(alpha=1),
                # DecisionTreeClassifier(max_depth=5),
                # AdaBoostClassifier(),
                # GaussianNB(),
                # QuadraticDiscriminantAnalysis(),
                MultiOutputRegressor(BaggingRegressor(random_state=1))]
            # print(self.trainY)
            for name, clf in zip(names, classifiers):
                try:
                    clf.fit(self.trainX, self.trainY)
                    score = clf.score(self.testX, self.testY)
                    print(name, score)
                except:
                    pass

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
        self.clf = MultiOutputRegressor(BaggingRegressor(random_state=1))
        print("BBBBBBBBBBBBBBBBBBBBBBBBBBBBB")
        print(self.trainX)
        print(self.trainY)
        print("FFFFFFFFFFFFFFFFFFFFFFFFFFFFF")
        self.clf.fit(self.trainX, self.trainY)
        # print(test)
        # print(self.nameX)
        # print(self.nameY)
        score = self.clf.score(self.testX, self.testY)
        with open('data/' + dataFile + '.rf.pkl', 'wb') as fid:
            pickle.dump([self.clf, self.nameX, self.nameY], fid)
        return score

    def classify(self, groupName, fingerpintFile):
        print("Classifing...")
        with open('data/' + groupName + '.rf.pkl', 'rb') as pickle_file:
            [self.clf, self.nameX, self.nameY] = pickle.load(pickle_file)


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
        prediction = self.clf.predict(newRow.reshape(1, -1))
        print("Predicition: ",prediction)
        predictStr = str(prediction[0][0])+","+str(prediction[0][1])
        predictionJson = {}
        # for i in range(len(prediction[0])):
        #     predictionJson[self.nameY[i]] = prediction[0][i]
        predictionJson[predictStr]=1
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
        randomF = RF()
        print("fileName:",filename)
        print("group:",group)
        if len(filename) == 0:
            # payload = json.dumps(randomF.learn(group, 1)).encode('utf-8')
            # print("filename length is zero!")
            payload = json.dumps(randomF.learn(group, 0.9)).encode('utf-8')
        else:
            # print("filename length isn't zero!")
            payload = json.dumps(
                randomF.classify(
                    group,
                    filename +
                    ".rftemp")).encode('utf-8')
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
        randomF = RF()
        print(randomF.classify(args.group, args.file))
    elif args.group is not None:
        randomF = RF()
        print(randomF.learn(args.group, 1))
        # print(randomF.learn(args.group, 0.5))
    else:
        print("""Usage:
To just run as TCP server:
	python3 rf.py --port 5009
To just learn:
	python3 rf.py --group GROUP
To classify
	python3 rf.py --group GROUP --file FILEWITHFINGERPRINTS
""")