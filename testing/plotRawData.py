import matplotlib.pyplot as plt
from matplotlib import cm
import random
# resultDic --> location:ap:rssList



resultDic = {}
apList = []

# get input file name and open it, then get maclist from first line
while(1):
    inpt = input(">> Enter csv file name: ")
    try:
        with open(inpt) as f:
            firstLine = f.readline().split(",")
            numAp = len(firstLine)-2-1 #x&y and \n (because there is extra comma at end of each line in csv file)
            for apn in range(numAp):
                apList.append(firstLine[apn+2])
            break
    except:
        print("\tFile not found!")

# open the file and parse lines 
with open(inpt) as f:
    lines = f.readlines()[1:]
    for line in lines:
        elements = line.split(",")
        xy = elements[0]+","+elements[1]
        if xy not in resultDic:
            resultDic[xy] = {}
            for i in range(0,numAp):
                resultDic[xy][apList[i]] = []
        for i in range(0,numAp):
            resultDic[xy][apList[i]].append(int(elements[2+i]))
        # resultDic[xy]["AP0"].append(int(elements[2]))
        # resultDic[xy]["AP1"].append(int(elements[3]))
        # resultDic[xy]["AP2"].append(int(elements[4]))

# print(resultDic)

breakNum = 0


# komeil_sham_oof
# for xy in resultDic:
#     for apNum in range(0,numAp):
#         # breakNum += 1
#         # if(breakNum == 4):
#         #     break
#         y = resultDic[xy]["AP"+str(apNum)]
#         # print(y)
#         fig = plt.figure(xy+":"+str(apNum))
#         fig.canvas.set_window_title(xy+":"+str(apNum))
#         plt.plot(list(range(len(y))), y, linewidth=2, linestyle="-", c=cm.hot(random.randint(40,200)))
#         plt.axis([0, 40, -110,-10])
# plt.show() 


# CMD
history = []
figNum = 0
inpt = "begin"
print("Enter 'help' to show commands ")

while(1):
    inpt = input(">> Enter x,y,apMac: ").strip()
    
    if(inpt == "help"):
        print("\n")
        print("\thelp --> show commands")
        print("\texit --> exit")
        print("\tshow --> show plots")
        print("\tdots --> show dots ")
        print("\tsearchInDots --> search in dots")
        print("\taps --> show num of aps")
        print("\thistory --> load history plots and print the history")
        print("\thistoryClear --> clear history")
        print("\thistoryMath --> min,max,Average of list of rss")
        print("\tbyMac --> load plot of an ap's rss in list of location")
        print("\n")

    elif(inpt=="exit"):
        break

    elif(inpt=="show"):
        figNum = 0
        plt.show()

    elif(inpt=="dots"):
        for xy in resultDic:
            print("\t"+xy)

    elif(inpt=="searchInDots"):
        txt = input(">> Enter text to search: ").strip()
        for xy in resultDic:
            if(txt in xy):
                print("\t"+xy)

    elif(inpt=="aps"):
        print(apList)

    elif(inpt=="historyClear"):
        history = []
        print("\thistory is clear")

    elif(inpt=="history"):
        figNum = 0
        for inpttemp in history:

            inptList = inpttemp.split(",")
            if(len(inptList)!=3):
                print("\tError")
                continue
            xy = inptList[0]+","+inptList[1]
            y = resultDic[xy][inptList[2]]
            print("\t"+inpttemp+": "+str(y))
            fig = plt.figure(xy+":"+inptList[2])
            figNum += 1
            fig.canvas.set_window_title(str(figNum)+":"+xy+":"+inptList[2])
            plt.plot(list(range(len(y))), y, linewidth=2, linestyle="-", c=cm.hot(random.randint(40,200)))
            plt.axis([0, len(y), -110,-10])

    elif(inpt=="historyMath"):
        try:
            for inpttemp in history:
                print("\n")
                inptList = inpttemp.split(",")
                if(len(inptList)!=3):
                    print("\tError")
                    continue
                xy = inptList[0]+","+inptList[1]
                y = resultDic[xy][inptList[2]]
                print("\t"+inpttemp+":")
                print("\t\t"+str(y))
                avgY = sum(y) / float(len(y))
                minY = min(y)
                maxY = max(y)
                print("\t\tAverage: "+str(avgY)+" Min: "+str(minY)+" Max: "+str(maxY))  
            print("\n")
        except Exception as e:
            print("\tError in historyMath mode")
            print(e)
            continue

    elif(inpt=="byMac"):
        try:
            inptMac = input(">> Enter mac: ").strip()
            inptLocList = input(">> Enter loc list (example: x1,y1 x2,y2): ").split(" ")

            figNum = 0
            for xy in inptLocList:
                y = resultDic[xy][inptMac]
                print("\t"+xy+": "+str(y))
                fig = plt.figure(xy+":"+inptMac)
                figNum += 1
                fig.canvas.set_window_title(str(figNum)+":"+xy+":"+inptMac)
                plt.plot(list(range(len(y))), y, linewidth=2, linestyle="-", c=cm.hot(random.randint(40,200)))
                plt.axis([0, len(y), -110,-10])
                history.append(xy+","+inptMac)

        except Exception as e:
            print("\tError in byMac mode")
            print(e)
            continue
    
    # default command is to get input as x,y,apMac format
    else:
        try:
            inptList = inpt.split(",")
            if(len(inptList)!=3):
                print("\tError")
                continue
            xy = inptList[0]+","+inptList[1]
            y = resultDic[xy][inptList[2]]
            print("\t"+str(y))
            fig = plt.figure(xy+":"+inptList[2])
            figNum += 1
            fig.canvas.set_window_title(str(figNum)+":"+xy+":"+inptList[2])
            plt.plot(list(range(len(y))), y, linewidth=2, linestyle="-", c=cm.hot(random.randint(40,200)))
            plt.axis([0, len(y), -110,-10])
            history.append(inpt)
            
        except Exception as e:
            print("\tError in x,y,mac mode")
            print(e)
            continue
    
# fig = plt.figure(1)
# fig.canvas.set_window_title('Window 2')
# plt.plot(list(range(len(y))), y, linewidth=2, linestyle="-", c="g")
# plt.axis([0, 40, -110,-10])
# plt.show() 

plt.close()
