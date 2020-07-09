import numpy
import pyparticleest.utils.kalman as kalman
import pyparticleest.interfaces as interfaces
import matplotlib.pyplot as plt
import pyparticleest.simulator as simulator
import pyparticleest.filter as filter
import math
import random
import time, threading
import os
import pickle
from particlefilterUtils import *

timeScale = 0.001
distScale = 0.01
LastTimestamp = 0

pfRunner = None
resultData = {}
resultLogFileName = "results.pkl"
try:
    os.remove(resultLogFileName)
except OSError:
    pass


########################################################################
###### Algorithm configs are set in 'init_particle_filter' function ######
# Static configs : capitalized syntax or with underscore
# Variables : camelcase syntax (first letter is lowercase)
# Functions : underscore case (first letter is capital)
########################################################################

class PredictionAndObservationModel(interfaces.ParticleFiltering):
    def __init__(self, Q, R, P0, human_speed, initLoc, mapGraph):
        self.P0 = P0
        self.initLoc = initLoc
        self.Q = numpy.copy(Q)
        self.R = numpy.copy(R)
        self.human_speed = human_speed
        self.mapGraph = mapGraph
        self.LastObserveDist = 0
        self.oldParticles = None

    def create_initial_estimate(self, N):
        # x,y,h
        init_x = numpy.random.normal(self.initLoc[0], self.P0, (N,)).reshape((N, 1))
        init_y = numpy.random.normal(self.initLoc[1], self.P0, (N,)).reshape((N, 1))
        init_h = numpy.random.uniform(0.0, 360.0, (N,)).reshape((N, 1))

        particles = numpy.concatenate((init_x, init_y, init_h), axis=1)

        # xy = list(self.world.random_free_place())
        # xy.append(random.uniform(0, 360))
        # particles = numpy.array([xy],dtype=float)
        # for _ in range(1,N):
        #     xy = list(self.world.random_free_place())
        #     xy.append(random.uniform(0, 360))
        #     particles = numpy.concatenate((particles,[xy]))
        return particles

    def sample_process_noise(self, particles, u, t):
        """ Return process noise for input u and t """
        N = len(particles)
        heading_noise = numpy.random.normal(0, self.Q[0], (N,)).reshape((N, 1))
        speed_noise = numpy.random.normal(self.human_speed / 2, self.human_speed / 2 + self.Q[1], (N,)).reshape((N, 1))
        noise = numpy.concatenate((heading_noise, speed_noise), axis=1)
        return noise

    def cross_wall(self, begin, end):
        """Check that the dot(that cross from 'begin' to 'end') crosses from map walls or don't"""
        pathLine = [(begin[0], begin[1]), (end[0], end[1])]
        for wall in self.mapGraph:  # wall or line
            if AreIntersect(pathLine, wall):
                return True
        return False

    def update(self, particles, u, t, noise):
        """ Update estimate using 'u, t and noise' as input """

        particleNum = len(particles)
        timespan = u * timeScale  # Convert to second

        h = (particles[:, 2] + noise[:, 0]) % 360
        # print(noise[:, 0])
        # h = particles[:, 2] + noise[:, 0]
        speed = noise[:, 1]

        # speed = numpy.random.uniform(-self.human_speed,self.human_speed,((N,))).reshape(1,N)
        # speed = numpy.random.uniform(-self.human_speed/2,self.human_speed,((N,))).reshape(1,N)

        h_rad = numpy.radians(h)
        dxy = numpy.zeros((particleNum, 2))
        dxy[:, 0] = numpy.array(numpy.cos(h_rad) * speed * timespan)
        dxy[:, 1] = numpy.array(numpy.sin(h_rad) * speed * timespan)

        ################################################################
        ### Normal:
        # particles[:, :2] += dxy
        ################################################################

        ################################################################
        ## Map constraint in prediction :
        oldParticles = numpy.copy(particles)

        particles[:, 2] = h
        particles[:, :2] += dxy

        for i in range(len(particles[:, :2])):
            count = 0
            altSpeed = speed[i]
            orgn = oldParticles[:, :2][i]
            dst = particles[:, :2][i]
            # while (numpy.linalg.norm(particles[:, :2][i] - lastMeanXY) > maxDist):
            # while (self.cross_wall(orgn, dst)):
            #     # print(orgn)
            #     if count > 10:
            #         particles[:, :2][i] = oldParticles[:, :2][i]
            #         break
            #     count += 1
            #     # altSpeed /= 2.0
            #     # print(numpy.random.normal(0, self.Q[0], 1)[0])
            #     h[i] += numpy.random.normal(0, 3 * self.Q[0], 1)[0]
            #     h_rad[i] = numpy.radians(h[i])
            #     dx = numpy.cos(h_rad[i]) * altSpeed * timespan
            #     dy = numpy.sin(h_rad[i]) * altSpeed * timespan
            #     particles[:, 2][i] = h[i]
            #     particles[:, :2][i] = oldParticles[:, :2][i] + numpy.array([dx, dy])
        ################################################################

        ################################################################
        ### Map constraint in prediction (other part is in update step):
        # self.oldParticles = numpy.copy(particles)
        # particles[:, :2] += dxy
        ################################################################

        ################################################################
        ### Avoid particle to go more further than maxDist algorithm
        # oldParticles = numpy.copy(particles)
        # particles[:, :2] += dxy
        #
        # for i in range(len(particles[:, :2])):
        #     count = 0
        #     altSpeed = speed[i]
        #     orgn = oldParticles[:, :2][i]
        #     while (numpy.linalg.norm(particles[:, :2][i] - lastMeanXY) > maxDist):
        #         # print(orgn)
        #         if count > 3:
        #             particles[:, :2][i] = oldParticles[:, :2][i]
        #             break
        #         count += 1
        #         altSpeed /= 2.0
        #         dx = numpy.cos(h_rad[i]) * altSpeed * timespan
        #         dy = numpy.sin(h_rad[i]) * altSpeed * timespan
        #         particles[:, :2][i] = oldParticles[:, :2][i] + numpy.array([dx, dy])
        ################################################################


        # oldParticles = numpy.copy(particles)
        # particles[:, :2] += dxy
        # particles[:, 2] = h
        #
        #
        # for i in range(len(particles[:, :2])):
        #     orgn = oldParticles[:, :2][i]
        #     dst = particles[:, :2][i]
        #     if self.cross_wall(orgn,dst):
        #         print("Cross_Wall: ", orgn," --> ",dst)


    def measure(self, particles, y, t):
        """ Return the log-pdf value of the measurement """

        N = len(particles)
        # If there isn't any observation, return zero(particles weight don't change)
        if len(y) <= 1:
            return numpy.zeros(len(particles), dtype=float)

        masterEstimation = y[0]
        slaveEstimation = y[1]

        # todo: handle situations that both of master and slave are present(it's so rare)
        coefficient = 1.0
        if masterEstimation[0] == 1:
            guess = numpy.array(masterEstimation[1:])
            print("############# Master")
            print(guess)
        elif slaveEstimation[0] == 1:
            coefficient = 2.25
            guess = numpy.array(slaveEstimation[1:])
            print("############# Slave")
            print(guess)

        ################################################################
        ### Restart algorithm
        # logyprob = numpy.empty(len(particles), dtype=float)
        # for k in range(len(particles)):
        #     particle = particles[k]
        #     dist = numpy.linalg.norm(particle[:2] - guess)
        #     if self.LastObserveDist > 2.0:
        #         print("self.LastObserveDist:", self.LastObserveDist)
        #         logyprob[k] = kalman.lognormpdf(dist, self.R * 0.1)
        #     else:
        #         logyprob[k] = kalman.lognormpdf(dist, self.R * coefficient)
        ################################################################

        ################################################################
        ### Normal
        logyprob = numpy.empty(len(particles), dtype=float)
        wights = numpy.empty(len(particles), dtype=float)

        for k in range(len(particles)):
            particle = particles[k]
            # if self.cross_wall(particle[:2], guess):
            #     dist = numpy.linalg.norm(particle[:2] - guess)
            # else:
            #     dist = numpy.linalg.norm(particle[:2] - guess)
            dist = numpy.linalg.norm(particle[:2] - guess)

            # dist = dist ** 3
            # weight = 1 / (dist + 0.0000001) * 1/N
            # wights[k] = weight
            # print(self.LastObserveDist)
            #########################################################
            ###Restarting
            # if self.LastObserveDist > 2.0 and masterEstimation[0] == 1:
            # logyprob[k] = kalman.lognormpdf(dist, self.R * coefficient / 10)
            # else:
            # logyprob[k] = kalman.lognormpdf(dist, self.R * coefficient)
            ##########################################################
            logyprob[k] = kalman.lognormpdf(dist, self.R * coefficient)

        ################################################################

        ################################################################
        ### Map constraint in update step (other part is in prediction step):
        # logyprob = numpy.empty(len(particles), dtype=float)
        # for i in range(len(particles[:, :2])):
        #     orgn = self.oldParticles[:, :2][i]
        #     particle = particles[:, :2][i]
        #
        #
        #     if (self.cross_wall(orgn,particle)):
        #         logyprob[i] = -math.inf
        #     else:
        #         dist = numpy.linalg.norm(particle[:2] - guess) / 100
        #         logyprob[i] = kalman.lognormpdf(dist, self.R * coefficient)

        ################################################################

        ################################################################
        ### Backup
        # logyprob = numpy.empty(len(particles), dtype=float)
        #
        # for k in range(len(particles)):
        #     particle = particles[k]
        #     dist = numpy.linalg.norm(particle[:2] - guess) / 100
        #
        #     # print("#########")
        #     # print(particle)
        #     # print(dist)
        #     # weight = 1/(dist +1)
        #     # print(weight)
        #     # print(kalman.lognormpdf(dist, self.R))
        #     if self.LastObserveDist > 2.0:
        #         print("self.LastObserveDist:", self.LastObserveDist)
        #         logyprob[k] = kalman.lognormpdf(dist, self.R * 0.1)
        #     else:
        #         logyprob[k] = kalman.lognormpdf(dist, self.R * coefficient)
        #     # logyprob[k] = numpy.log(numpy.array([[weight]]))
        #
        #     #
        #     # if self.world.is_free(p[0],p[1]):
        #     #     p_d_val = self.world.distance_to_nearest_beacon(p[0],p[1])
        #     #     p_d = numpy.array([p_d_val])
        #     #     if self.USE_BEACON :
        #     #         logyprob[k] = kalman.lognormpdf(numpy.linalg.norm(particles[k][:2] - y), self.R)
        #     #     else:
        #     #         logyprob[k] = kalman.lognormpdf(numpy.array([0]), self.R)
        #     # else:
        #     #     logyprob[k] = numpy.array([-100])
        ################################################################



        return logyprob


def append_data(obj):
    """Append new obj to result log file"""
    with open(resultLogFileName, 'ab') as f:
        # f.write(str(timestamp)+":"+str(loc[0])+","+str(loc[1]) + os.linesep)
        pickle.dump(obj, f)


def init_particle_filter(timestamp, initLoc, mapGraph):
    """Initialize model and configs of particle filter """
    global LastTimestamp, pfRunner, resultData
    # init_loc = numpy.array([-298.0, -772.0])

    # Set static random
    numpy.random.seed(1)
    random.seed(1)

    #Initialize LastTimestamp
    LastTimestamp = timestamp

    ########################################################################
    ############### Algorithm Configs ################
    NUM_OF_PARTICLES = 1000
    P0 = 20.0
    HUMAN_SPEED = 1.0 * 100.0
    HUMAN_MAX_HEADING_CHANGE = 30
    Q = numpy.asarray((HUMAN_MAX_HEADING_CHANGE, HUMAN_SPEED * 0.1))  # heading, speed variances
    # R = numpy.asarray(((0.5,),))
    # R = numpy.asarray(((10,),))
    R = numpy.asarray(((20000,),))
    # R = numpy.asarray(((0 ** 2,),))
    RESAMPLING_THRESHOLD = 2.0 / 3.0
    ########################################################################

    # Create the model from set configs
    model = PredictionAndObservationModel(Q, R, P0, HUMAN_SPEED, initLoc, mapGraph)

    # Create pf runner
    pfRunner = simulator.Simulator(model, u=None, y=initLoc)
    pfRunner.pt = filter.ParticleTrajectory(pfRunner.model, NUM_OF_PARTICLES, resample=RESAMPLING_THRESHOLD)

    # Write initialization data to result log file
    threading.Thread(target=append_data, args=([[timestamp, initLoc]])).start()


def predict_particle_filter(timestamp):
    """With Empty observation, the filter just predicts particles next dimensions"""
    global LastTimestamp, pfRunner

    # Calculate time difference with last
    timeDiff = timestamp - LastTimestamp
    u = numpy.array(timeDiff)
    LastTimestamp = timestamp

    # It ignores measurement and just runs the prediction step
    y = numpy.array([None])

    # Forward filter
    # pfRunner.pt.forward(u, y)
    if pfRunner.pt.forward(u, y):
        print("Resampling occured ")

    resultXY = pfRunner.get_filtered_mean()[-1]

    # Prepare data to append to the result log file
    parts, ws = pfRunner.get_filtered_estimates()
    particles = parts[-1]

    # Write result to result log file
    output = [timestamp, resultXY, particles, ws[-1], []]
    threading.Thread(target=append_data, args=([output])).start()

    # return result x,y to golang server
    return [resultXY[0], resultXY[1]]


def predict_and_update_particle_filter(timestamp, updateRequest):
    """With non-Empty observation, the filter first predicts particles next dimensions and then update their weights"""
    global LastTimestamp, pfRunner

    # Calculate time difference with last
    timeDiff = timestamp - LastTimestamp
    u = numpy.array(timeDiff)
    LastTimestamp = timestamp

    masterEstimation = updateRequest.MasterEstimation[:]
    slaveEstimation = updateRequest.SlaveEstimation[:]
    trueLocation = updateRequest.TrueLocation[:]

    y = numpy.array([
        [0, math.inf, math.inf],  # first one is Master observation
        [0, math.inf, math.inf],  # second one is Slave observation
    ])

    if len(masterEstimation) != 0:  # Master Observation
        y[0] = [1, masterEstimation[0], masterEstimation[1]]
    if len(slaveEstimation) != 0:  # Slave Observation
        y[1] = [1, slaveEstimation[0], slaveEstimation[1]]

    # Forward filter
    # pfRunner.pt.forward(u, y)
    if pfRunner.pt.forward(u, y):
        print("Resampling occured ")

    # Get weighted mean of particles as the result
    resultXY = pfRunner.get_filtered_mean()[-1]

    # Compute distance between the observation and resultXY
    if len(masterEstimation) != 0:  # Master Observation
        pfRunner.model.LastObserveDist = numpy.linalg.norm(y[0][1:] - resultXY[:2]) * distScale
        print(pfRunner.model.LastObserveDist)
    if len(slaveEstimation) != 0:  # Slave Observation
        pfRunner.model.LastObserveDist = numpy.linalg.norm(y[1][1:] - resultXY[:2]) * distScale
        print(pfRunner.model.LastObserveDist)

    # Prepare data to append to the result log file
    EstimationAndTrueLocation = numpy.copy(y)
    EstimationAndTrueLocation = numpy.vstack([EstimationAndTrueLocation, [1, trueLocation[0], trueLocation[1]]])

    # particles = sim.pt.traj[-1].pa.part
    parts, ws = pfRunner.get_filtered_estimates()
    particles = parts[-1]

    # Write result to result log file
    output = [timestamp, resultXY, particles, ws[-1], EstimationAndTrueLocation]
    threading.Thread(target=append_data,
                     args=([output])).start()

    # return result x,y to golang server
    return [resultXY[0], resultXY[1]]


def convet_map_graph_to_float_list(mapGraph):
    """Convert Protobuf map format to 3 dimensions float list( each line contains 2 dots and each dot contains X and Y )"""
    newFloatGraph = []
    for line in mapGraph.Lines:
        newFloatLine = []
        for dot in line.Dots:
            newFloatDot = [dot.XY[0], dot.XY[1]]
            newFloatLine.append(newFloatDot)
        newFloatGraph.append(newFloatLine)
    return newFloatGraph


###############################
# Server functions:
def Initialize(initRequest):
    """Get initRequest, create map according to map Graph and init_particlefilter"""
    timestamp = initRequest.Timestamp
    initLoc = numpy.array(initRequest.XY)
    mapGraph = convet_map_graph_to_float_list(initRequest.MapGraph)

    print("Initialization: ", timestamp)
    print("Initial Location:", initLoc)
    print("Map Graph :", mapGraph)
    init_particle_filter(initRequest.Timestamp, initLoc, mapGraph)

    return True

def Predict(predictRequest):
    """Get predictRequest and run predict_particlefilter"""
    timestamp = predictRequest.Timestamp
    print("Prediction: ", timestamp)
    return predict_particle_filter(timestamp), True

def Update(updateRequest):
    """Get updateRequest and run predict_and_update_particlefilter"""
    timestamp = updateRequest.Timestamp
    print("Update: ", timestamp)
    return predict_and_update_particle_filter(timestamp, updateRequest), True
