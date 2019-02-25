import numpy
import pyparticleest.utils.kalman as kalman
import pyparticleest.interfaces as interfaces
import matplotlib.pyplot as plt
import pyparticleest.simulator as simulator
import pyparticleest.filter as filter
import math
import random
import time, threading
from draw import Figure
import os
import pickle

LastTimestamp = int(time.time() * 1000)
sim = None
resultData = {}
resultDataFileName = "results.pkl"
try:
    os.remove(resultDataFileName)
except OSError:
    pass


class Model(interfaces.ParticleFiltering):
    """ x_{k+1} = x_k + v_k, v_k ~ N(0,Q)
        y_k = x_k + e_k, e_k ~ N(0,R),
        x(0) ~ N(0,P0) """

    def __init__(self, Q, R, P0, init_loc, human_speed):
        self.P0 = P0
        self.init_loc = init_loc
        self.Q = numpy.copy(Q)
        self.R = numpy.copy(R)
        self.human_speed = human_speed
        self.LastObserveDist = 0

    def create_initial_estimate(self, N):
        # x,y,h
        init_x = numpy.random.normal(self.init_loc[0], self.P0, (N,)).reshape((N, 1))
        init_y = numpy.random.normal(self.init_loc[1], self.P0, (N,)).reshape((N, 1))
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
        """ Return process noise for input u """
        N = len(particles)
        heading_noise = numpy.random.normal(0, self.Q[0], (N,)).reshape((N, 1))
        speed_noise = numpy.random.normal(0, self.human_speed + self.Q[1], (N,)).reshape((N, 1))
        noise = numpy.concatenate((heading_noise, speed_noise), axis=1)
        return noise

    def update(self, particles, u, t, noise):
        """ Update estimate using 'data' as input """

        N = len(particles)
        # print(u)
        timespan = u / 1000.0  # second
        # noisy_u = noise + u
        # print(noisy_u)

        # h = noisy_u[:,0]
        # speed = noisy_u[:,1]

        # print(h)
        # print(speed)
        # print(h)
        # print(speed)
        # h = u[0] + noise[:,0]
        # speed = u[1] + noise[:,1]
        h = (particles[:, 2] + noise[:, 0]) % 3600
        # h = particles[:, 2] + noise[:, 0]
        speed = noise[:, 1]

        # speed = numpy.random.uniform(-self.human_speed,self.human_speed,((N,))).reshape(1,N)
        # speed = numpy.random.uniform(-self.human_speed/2,self.human_speed,((N,))).reshape(1,N)

        h_rad = numpy.radians(h)
        dxy = numpy.zeros((N, 2))
        dxy[:, 0] = numpy.array(numpy.sin(h_rad) * speed * timespan)
        dxy[:, 1] = numpy.array(numpy.cos(h_rad) * speed * timespan)

        particles[:, :2] += dxy
        particles[:, 2] = h

    def measure(self, particles, y, t):
        """ Return the log-pdf value of the measurement """


        if len(y) <= 1:
            return numpy.zeros(len(particles), dtype=float)

        masterEstimation = y[0]
        slaveEstimation = y[1]

        # todo: handle situations that both of master and slave are present
        coefficient = 1.0
        if masterEstimation[0] == 1:
            guess = numpy.array(masterEstimation[1:])
            print("############# Master")
            print(guess)
        elif slaveEstimation[0] == 1:
            coefficient = 1.2
            guess = numpy.array(slaveEstimation[1:])
            print("############# Slave")
            print(guess)

        logyprob = numpy.empty(len(particles), dtype=float)


        for k in range(len(particles)):
            particle = particles[k]
            dist = numpy.linalg.norm(particle[:2] - guess) / 100


            # print("#########")
            # print(particle)
            # print(dist)
            # weight = 1/(dist +1)
            # print(weight)
            # print(kalman.lognormpdf(dist, self.R))
            # if self.LastObserveDist > 2.0:
            #     logyprob[k] = kalman.lognormpdf(dist/5, self.R * 0.01)
            # else:
            logyprob[k] = kalman.lognormpdf(dist, self.R * coefficient)
            # logyprob[k] = numpy.log(numpy.array([[weight]]))

            #
            # if self.world.is_free(p[0],p[1]):
            #     p_d_val = self.world.distance_to_nearest_beacon(p[0],p[1])
            #     p_d = numpy.array([p_d_val])
            #     if self.USE_BEACON :
            #         logyprob[k] = kalman.lognormpdf(numpy.linalg.norm(particles[k][:2] - y), self.R)
            #     else:
            #         logyprob[k] = kalman.lognormpdf(numpy.array([0]), self.R)
            # else:
            #     logyprob[k] = numpy.array([-100])
        return logyprob


# def draw_fig(init_loc):
#     world = Figure()
#     world.draw()
#     # world.show_mean(init_loc[0], init_loc[1])
#     world.show_mean(init_loc[0], init_loc[1])
#
#     while(1):
#         pass


def AppendData(obj):
    with open(resultDataFileName, 'ab') as f:
        # f.write(str(timestamp)+":"+str(loc[0])+","+str(loc[1]) + os.linesep)
        pickle.dump(obj, f)


def init_particlefilter(timestamp, init_loc):
    global LastTimestamp, sim, resultData
    # init_loc = numpy.array([-298.0, -772.0])

    numpy.random.seed(1)
    random.seed(1)
    LastTimestamp = timestamp

    # threading.Thread(target=draw_fig,args =(init_loc,), daemon=True).start()
    threading.Thread(target=AppendData, args=([[timestamp, init_loc]])).start()

    num = 1000
    P0 = 20.0
    human_speed = 1.0 * 100.0
    human_heading_change = 180
    Q = numpy.asarray((human_heading_change, human_speed * 0.1))  # heading, speed variances
    # R = numpy.asarray(((0.5,),))
    # R = numpy.asarray(((10,),))
    R = numpy.asarray(((8,),))
    # R = numpy.asarray(((0 ** 2,),))

    # init_loc = numpy.array([0.0,0.0])
    model = Model(Q, R, P0, init_loc, human_speed)
    sim = simulator.Simulator(model, u=None, y=init_loc)

    resamplings = 0

    sim.pt = filter.ParticleTrajectory(sim.model, num)
    # # result_history = []
    # for i in range(1000):
    #     # Run PF using noise corrupted input signal
    #     # sensor measurement
    #     ble_location = [0,.0,0.0]
    #     y = numpy.array(ble_location)
    #
    #     # forward filter
    #     if (sim.pt.forward(None, y)):
    #         resamplings = resamplings + 1
    #
    #     meanXY= sim.get_filtered_mean()[-1]
    #     model.x = meanXY[0]
    #     model.y = meanXY[1]
    #
    #     # ---------- Show current state ----------
    #     # model.world.show_mean(meanXY[0],meanXY[1])
    #     # model.world.show_robot(robot)
    #     # model.world.show_particles_2(sim.pt.traj[-1].pa.part)


def predict_particlefilter(timestamp):
    global LastTimestamp, sim

    # return [1.0,1.0]
    u = numpy.array(timestamp - LastTimestamp)
    LastTimestamp = timestamp
    y = numpy.array([None])  # It's ignore measurement

    # forward filter
    sim.pt.forward(u, y)

    meanXY = sim.get_filtered_mean()[-1]
    parts, ws = sim.get_filtered_estimates()
    particles = parts[-1]
    threading.Thread(target=AppendData, args=([[timestamp, meanXY, particles, ws[-1], []]])).start()


    return [meanXY[0], meanXY[1]]


def setLastObserveDist(meanXY, y):
    if y[0][0] == 1:
        return numpy.linalg.norm(y[0][1:] - meanXY[:2]) / 100
    elif y[1][0] == 1:
        return numpy.linalg.norm(y[1][1:] - meanXY[:2]) / 100


def update_particlefilter(timestamp, updateRequest):
    global LastTimestamp, sim

    # return [1.0,1.0]
    u = numpy.array(timestamp - LastTimestamp)
    LastTimestamp = timestamp

    # y = ble_predict  # It's ignore measurement
    masterEstimation = updateRequest.MasterEstimation[:]
    slaveEstimation = updateRequest.SlaveEstimation[:]
    trueLocation = updateRequest.TrueLocation[:]

    y = numpy.array([
        [0, math.inf, math.inf],  # master
        [0, math.inf, math.inf],  # slave
    ])
    if len(masterEstimation) != 0:
        y[0] = [1, masterEstimation[0], masterEstimation[1]]
    if len(slaveEstimation) != 0:
        y[1] = [1, slaveEstimation[0], slaveEstimation[1]]



    # forward filter
    sim.pt.forward(u, y)

    meanXY = sim.get_filtered_mean()[-1]

    # sim.model.LastObserveDist = setLastObserveDist(meanXY, y)
    # print(sim.model.LastObserveDist)
    # threading.Thread(target=AppendData,args =(timestamp,meanXY)).start()
    EstimationAndTrueLocation = numpy.copy(y)
    EstimationAndTrueLocation = numpy.vstack([EstimationAndTrueLocation, [1, trueLocation[0], trueLocation[1]]])
    # particles = sim.pt.traj[-1].pa.part
    parts, ws = sim.get_filtered_estimates()
    particles = parts[-1]
    threading.Thread(target=AppendData,
                     args=([[timestamp, meanXY, particles, ws[-1], EstimationAndTrueLocation]])).start()

    # threading.Thread(target=world.show_mean,args =(meanXY[0],meanXY[1],), daemon=True).start()
    # try :
    #     threading.Thread(target=show_mean,args =(meanXY,), daemon=True).start()
    # except Exception as e:
    #     print('Error: ' + str(e))

    # world.show_mean(meanXY[0], meanXY[1])
    # print(meanXY)
    # world.show_particles_2(sim.pt.traj[-1].pa.part)

    return [meanXY[0], meanXY[1]]


# server functions
def Initialize(initRequest):
    timestamp = initRequest.Timestamp
    init_loc = numpy.array(initRequest.XY)

    print("Initialization: ", timestamp)
    print("Initial Location:", init_loc)
    init_particlefilter(initRequest.Timestamp, init_loc)

    return True


def Predict(predictRequest):
    timestamp = predictRequest.Timestamp
    print("Prediction: ", timestamp)
    return predict_particlefilter(timestamp), True


def Update(updateRequest):
    timestamp = updateRequest.Timestamp
    print("Update: ", timestamp)
    return update_particlefilter(timestamp, updateRequest), True

#
# def test_particlefilte():
#     init_particlefilter(int(time.time() * 1000), numpy.array([0.0, 0.0]))
#     predict_particlefilter(int(time.time() * 1000) + 1000)
#     update_particlefilter(int(time.time() * 1000) + 1000, numpy.array([1.0, 1.0]))

# def priodic_predict(period):
#     predict_particlefilter(int(time.time() * 1000))
#     threading.Timer(period, priodic_predict).start()
# test_particlefilte()
