# ------------------------------------------------------------------------
# coding=utf-8
# ------------------------------------------------------------------------
#
#  Created by Martin J. Laubach on 2011-11-15
#
# ------------------------------------------------------------------------

import math
import turtle
import random

UPDATE_EVERY = 0
DRAW_EVERY = 2


class Figure(object):
    def __init__(self):
        s = turtle.Screen()
        self.mapScale = 0.25
        width = 494
        height = 522
        s.setup(width, height)
        s.bgpic('ArmanExactMap.png')

        turtle.tracer(50000, delay=0)
        turtle.register_shape("dot", ((-3, -3), (-3, 3), (3, 3), (3, -3)))
        turtle.register_shape("tri", ((-3, -2), (0, 3), (3, -2), (0, 0)))
        turtle.speed(0)
        turtle.title("Poor robbie is lost")

        self.width = width
        self.height = height
        turtle.setworldcoordinates(-self.width / 2, -self.height / 2, self.width / 2, self.height / 2)
        self.blocks = []
        self.update_cnt = 0
        self.one_px = float(turtle.window_width()) / float(self.width) / 2

        self.beacons = []

        self.wn = turtle.Screen()
        # self.blocks.append((10, 30))
        # for y, line in enumerate(self.maze):
        #     for x, block in enumerate(line):
        #         if block:
        #             nb_y = self.height - y - 1
        #             self.blocks.append((x, nb_y))
        # if block == 2:
        #     self.beacons.extend(((x, nb_y), (x+1, nb_y), (x, nb_y+1), (x+1, nb_y+1)))

    def draw(self):
        # for x, y in self.blocks:
        turtle.up()
        turtle.setposition(0, 0)
        turtle.down()
        turtle.setheading(90)
        turtle.begin_fill()
        for _ in range(0, 4):
            turtle.fd(1)
            turtle.right(90)
        turtle.end_fill()
        turtle.up()

        # turtle.color("#00ffff")
        # for x, y in self.beacons:
        #     turtle.setposition(x, y)
        #     turtle.dot()
        turtle.update()

    def weight_to_color(self, weight):
        return "#%02x00%02x" % (int(weight * 255), int((1 - weight) * 255))

    def is_in(self, x, y):
        if x < 0 or y < 0 or x > self.width or y > self.height:
            return False
        return True

    def is_free(self, x, y):
        if not self.is_in(x, y):
            return False

        yy = self.height - int(y) - 1
        xx = int(x)
        return self.maze[yy][xx] == 0

    def show_mean(self, x, y, confident=False):
        turtle.clearstamps()
        turtle.color("blue")
        turtle.shape("circle")
        turtle.setposition([x * self.mapScale, y * self.mapScale])
        turtle.stamp()
        # turtle.update()

    # def show_particles(self, particles):
    #     self.update_cnt += 1
    #     if UPDATE_EVERY > 0 and self.update_cnt % UPDATE_EVERY != 1:
    #         return
    #
    #     turtle.clearstamps()
    #     turtle.shape('tri')
    #
    #     draw_cnt = 0
    #     px = {}
    #     for p in particles:
    #         draw_cnt += 1
    #         if DRAW_EVERY == 0 or draw_cnt % DRAW_EVERY == 1:
    #             # Keep track of which positions already have something
    #             # drawn to speed up display rendering
    #             scaled_x = int(p.x * self.one_px)
    #             scaled_y = int(p.y * self.one_px)
    #             scaled_xy = scaled_x * 10000 + scaled_y
    #             if not scaled_xy in px:
    #                 px[scaled_xy] = 1
    #                 turtle.setposition(*p.xy)
    #                 turtle.setheading(90 - p.h)
    #                 turtle.color(self.weight_to_color(p.w))
    #                 turtle.stamp()

    def show_particles_2(self, particles):
        self.update_cnt += 1
        if UPDATE_EVERY > 0 and self.update_cnt % UPDATE_EVERY != 1:
            return

        turtle.clearstamps()
        turtle.shape('tri')

        draw_cnt = 0
        px = {}
        for i, p in enumerate(particles):

            draw_cnt += 1
            if DRAW_EVERY == 0 or draw_cnt % DRAW_EVERY == 1:
                # Keep track of which positions already have something
                # drawn to speed up display rendering
                scaled_x = int(p[0] * self.one_px)
                scaled_y = int(p[1] * self.one_px)
                scaled_xy = scaled_x * 10000 + scaled_y
                if not scaled_xy in px:
                    px[scaled_xy] = 1

                    turtle.setposition(p[0] * self.mapScale, p[1] * self.mapScale)
                    turtle.setheading(90 - p[2])
                    # turtle.color(self.weight_to_color(weights[i]))
                    turtle.color("#b74d2a")
                    turtle.stamp()
        turtle.update()

    def show_robot(self, xyh):
        turtle.clearstamps()
        turtle.color("green")
        turtle.shape('turtle')
        turtle.setposition([xyh[0] * self.mapScale, xyh[1] * self.mapScale])
        turtle.setheading(90 - xyh[2])
        turtle.stamp()

    def random_place(self):
        x = random.uniform(0, self.width)
        y = random.uniform(0, self.height)
        return x, y

    def random_free_place(self):
        while True:
            x, y = self.random_place()
            if self.is_free(x, y):
                return x, y

    def distance(self, x1, y1, x2, y2):
        return math.sqrt((x1 - x2) ** 2 + (y1 - y2) ** 2)

    def distance_to_nearest_beacon(self, x, y):
        d = 99999
        for c_x, c_y in self.beacons:
            distance = self.distance(c_x, c_y, x, y)
            if distance < d:
                d = distance
                d_x, d_y = c_x, c_y

        return d
