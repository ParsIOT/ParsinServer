# ./usr/bin/python

import grpc
from concurrent import futures
import time

# import the generated classes
import particlefilterclasses.particlefilter_pb2 as pf_classes_pb2
import particlefilterclasses.particlefilter_pb2_grpc as pf_classes_pb2_grpc

import particlefilter  # main particle filter functions
import threading

class ParticleFilterServicer(pf_classes_pb2_grpc.ParticleFilterServicer):

    def ConnectionTest(selfs, empty, context):
        print("Connection tested.")
        return pf_classes_pb2.InitReply(ReturnValue=True)

    def Initialize(self, initRequest, context):
        # threading.Thread(target=particlefilter.Initialize,args =(initRequest,), daemon=True).start()
        ReturnValue = particlefilter.Initialize(initRequest)
        # ReturnValue = True
        return pf_classes_pb2.InitReply(ReturnValue=ReturnValue)

    def Predict(self, predictRequest, context):
        ResXY, ReturnValue = particlefilter.Predict(predictRequest)
        return pf_classes_pb2.PredictReply(ResXY=ResXY, ReturnValue=ReturnValue)

    def Update(self, updateRequest, context):
        ResXY, ReturnValue = particlefilter.Update(updateRequest)
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
