from draw import Figure
import time
import turtle
import pickle

resultDataFileName = "results.pkl"

world = Figure()
world.draw()

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
lastLine = 0
line = lines[0]
print(line)
timestamp = line[0]
xy = line[1]
world.show_mean(xy[0], xy[1])


def nextLine():
    global lastLine, world

    if lastLine >= len(lines) - 1:
        return
    lastLine += 1
    line = lines[lastLine]
    print(line)
    timestamp = line[0]
    xy = line[1]
    particles = line[2]
    world.show_particles_2(particles)
    world.show_mean(xy[0], xy[1])


def previousLine():
    global lastLine, world

    if lastLine == 0:
        return
    lastLine -= 1
    line = lines[lastLine]
    print(line)
    timestamp = line[0]
    xy = line[1]
    particles = line[2]
    world.show_particles_2(particles)
    world.show_mean(xy[0], xy[1])


# world.wn.onclick(nextLine)
turtle.onkey(nextLine, "Right")
turtle.onkey(previousLine, "Left")

turtle.listen()

input()
# while 1:
#     for i in range(0,100,10):
#         print(i,i)
#         world.show_mean(-2*i,-2*i)
#
#         world.show_particles_2([[i,i,1],[0,i,0],[i,0,0]])
#
#         # world.show_robot([2*i,2*i,i])
#         time.sleep(0.1)

# def draw_fig(init_loc):
#     world = Figure()
#     world.draw()
#     # world.show_mean(init_loc[0], init_loc[1])
#     world.show_mean(init_loc[0], init_loc[1])
#
#     while(1):
#         pass
