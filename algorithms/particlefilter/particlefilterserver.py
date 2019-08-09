#!/usr/bin/python3

import grpc
from concurrent import futures
import time
import traceback

# import the generated classes
import particlefilterclasses.particlefilter_pb2 as pf_classes_pb2
import particlefilterclasses.particlefilter_pb2_grpc as pf_classes_pb2_grpc

import particlefilter  # main particle filter functions
import threading

class ParticleFilterServicer(pf_classes_pb2_grpc.ParticleFilterServicer):

    def ConnectionTest(selfs, empty, context):
        print("Connection tested.")
        try:
            initReply = pf_classes_pb2.InitReply(ReturnValue=True)
        except Exception as e:
            print(traceback.format_exc())
            raise e
        return initReply

    def Initialize(self, initRequest, context):
        try:
            ReturnValue = particlefilter.Initialize(initRequest)
        except Exception as e:
            print(traceback.format_exc())
            raise e
        return pf_classes_pb2.InitReply(ReturnValue=ReturnValue)

    def Predict(self, predictRequest, context):
        try:
            ResXY, ReturnValue = particlefilter.Predict(predictRequest)
        except Exception as e:
            print(traceback.format_exc())
            raise e
        return pf_classes_pb2.PredictReply(ResXY=ResXY, ReturnValue=ReturnValue)

    def Update(self, updateRequest, context):
        try:
            ResXY, ReturnValue = particlefilter.Update(updateRequest)
        except Exception as e:
            print(traceback.format_exc())
            raise e
        return pf_classes_pb2.UpdateReply(ResXY=ResXY, ReturnValue=ReturnValue)

# create a gRPC server
server = grpc.server(futures.ThreadPoolExecutor(max_workers=10))

pf_classes_pb2_grpc.add_ParticleFilterServicer_to_server(
    ParticleFilterServicer(), server)

# listen on port 50051
print('Starting server. Listening on port 50051.')
server.add_insecure_port('[::]:50051')
server.start()

# since server.start() will not block,
# a sleep-loop is added to keep alive
try:
    while True:
        time.sleep(86400)
except KeyboardInterrupt:
    server.stop(0)
