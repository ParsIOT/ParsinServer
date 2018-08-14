# import json,urllib
# data = urllib.urlopen("").read()
# output = json.loads(data)
# print (output)
import requests
import math
import operator
import numpy as np
import matplotlib.pyplot as plt

As = []
memNum = 3
print("Each location group contains "+str(memNum)+" locations.")
print("Url Example: "+'http://127.0.0.1:8003/locations?group=arman_4_11_96_ble_1')
url = input("Enter url: ")

for i in range(101):
    As.append((i)/100)

As = [0.04]


for a in As :


    def dist(d1,d2):
        xy1 = d1.split(",")
        xy1 = float(xy1[0]),float(xy1[1])
        xy2 = d2.split(",")
        xy2 = float(xy2[0]),float(xy2[1])
        return math.sqrt((xy1[0]-xy2[0])**2 + (xy1[1]-xy2[1])**2)

    rj = requests.get(url, auth=('admin', 'admin')).json()
    locs = []
    for loc in rj['locations']:
        locs.append(loc.strip())

    # print(len(locs))

    groupedLocs = []
    passedLoc = []
    for loc1 in locs:
        if loc1 in passedLoc:
            continue
        passedLoc.append(loc1)
        distList = {}
        locs.remove(loc1)
        for loc2 in locs:
            if loc1 != loc2:
                distList[loc2] = dist(loc1,loc2)    
        sortedDistList = sorted(distList.items(), key=operator.itemgetter(1)) #sort by val(dist)
        
        # Saving two nearer dots to loc1
        i = 0
        tempGroup = [loc1]
        # print(loc1)
        for dot,_ in sortedDistList:
            # print(dot)
            tempGroup.append(dot)
            # locs.remove(dot)
            passedLoc.append(dot)
            i+=1
            if i == memNum-1:
                # print(tempGroup)
                break
        groupedLocs.append(tempGroup) # add 3 member group to groupLocs


    # print(len(groupedLocs))
    # for gp in groupedLocs:
    #     print(gp)

    # sort groups according to Y val
    # mediumDotDict = {}
    # for gp in groupedLocs:
    #     y = 0.0
    #     for loc in gp :
    #         xy = loc.split(",")
    #         y += float(xy[1])
    #     y /= 3
    #     mediumDotDict[y] = gp

    # xySumList = list(mediumDotDict.keys())
    # xySumList.sort()

    # for xySum in xySumList:
    #     print(xySum)
    #     print(mediumDotDict[xySum])
        
    def mediumDot(group):
        x = 0.0
        y = 0.0
        for loc in group :
            xy = loc.split(",")
            x += float(xy[0])
            y += float(xy[1])
        x /= 3
        y /= 3
        return list((x,y))

    def mediumDotStr(group):
        x = 0.0
        y = 0.0
        for loc in group :
            xy = loc.split(",")
            x += float(xy[0])
            y += float(xy[1])
        x /= 3
        y /= 3
        return str(x)+","+str(y)


    mediumDotDict = {}
    for gp in groupedLocs:
        x,y = mediumDot(gp)
        mediumDotDict[x+y] = gp

    xySumList = list(mediumDotDict.keys())
    xySumList.sort()

    # print("########")
    # print(xySumList)
    refGroup = mediumDotDict[xySumList[0]]
    lastGroup = mediumDotDict[xySumList[-1]]
    minDistConst = dist(mediumDotStr(refGroup),mediumDotStr(lastGroup))

    # print(minDist)

    # print(refGroup)
    # print(mediumDot(refGroup))
    # print(lastGroup)
    # print(mediumDot(lastGroup))
    # print("########")

    newGpLocs = groupedLocs[:]

    midX,midY = 0,0

    clusterdGPs = []
    lastGP = refGroup
    clusterdGPs.append(refGroup)
    newGpLocs.remove(refGroup)
    refGroupMed = mediumDot(refGroup)
    while len(newGpLocs) > 0:

        minDist = minDistConst
        # print("#############")
        # print(refGroup)
        # print("-------------")
        # print(newGpLocs)
        # print("#############")
        # newGpLocs.remove(refGroup)
        # clusterdGPs.append(refGroup)

        refGroupMed = [0,0]
       
        for gp in clusterdGPs:
            gpMed = mediumDot(gp)
            refGroupMed[0] =refGroupMed[0] + gpMed[0]
            refGroupMed[1] =refGroupMed[1] + gpMed[1]
        
        refGroupMed[0] /= len(clusterdGPs)
        refGroupMed[1] /= len(clusterdGPs)

        for gp in newGpLocs:
            d = a*dist(mediumDotStr(gp),str(refGroupMed)[1:-1]) + (1-a)*dist(mediumDotStr(gp),mediumDotStr(lastGP))
            if minDist > d :
                minDist = d
                nextGP = gp
        
        
        
        # print(nextGP)
        # nextGPmed = mediumDot(nextGP)
        # refGroupMed[0] += nextGPmed[0]
        # refGroupMed[0] /= 2
        # refGroupMed[1] += nextGPmed[1]
        # refGroupMed[1] /= 2
        clusterdGPs.append(nextGP)
        newGpLocs.remove(nextGP)
        lastGP = nextGP
        

    mediumDots = []
    allList = []
    for gp in clusterdGPs:
        print(gp)
        for loc in gp:
            allList.append(loc)
        mediumDots.append(mediumDot(gp))
    # for xySum in xySumList:
    #     print(xySum)
    #     print(mediumDotDict[xySum])

    # print(mediumDots)
    print("\n\n")
    print(allList)
    print("\n\n")
    for dot in allList:
        print(dot)

    # N = 50
    # x = np.random.rand(N)
    # y = np.random.rand(N)


    # print(len(mediumDots))
    Xs = []
    Ys = []
    for d in mediumDots:
        Xs.append(d[1])
        Ys.append(d[0])
    plt.plot(Xs,Ys,'-')
    plt.scatter(Xs,Ys)
    plt.title("a:"+str(a))
    plt.show()
    
    