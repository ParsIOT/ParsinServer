# AreIntersect and two inner functions(onSegment & orientation) are provided to checking intersection of two 2D lines
def onSegment(p, q, r):
    if (q[0] <= max(p[0], r[0]) and q[0] >= min(p[0], r[0]) and q[1] <= max(p[1], r[1]) and q[1] >= min(p[1], r[1])):
        return True
    return False


def orientation(p, q, r):
    val = (q[1] - p[1]) * (r[0] - q[0]) - (q[0] - p[0]) * (r[1] - q[1])
    if (val == 0):
        return 0
    return 1 if val > 0 else 2


# Inspired by :https://www.geeksforgeeks.org/check-if-two-given-line-segments-intersect/
def AreIntersect(line1, line2):
    p1 = line1[0]
    q1 = line1[1]
    p2 = line2[0]
    q2 = line2[1]

    o1 = orientation(p1, q1, p2)
    o2 = orientation(p1, q1, q2)
    o3 = orientation(p2, q2, p1)
    o4 = orientation(p2, q2, q1)

    if (o1 != o2 and o3 != o4):
        return True
    if (o1 == 0 and onSegment(p1, p2, q1)):
        return True
    if (o2 == 0 and onSegment(p1, q2, q1)):
        return True
    if (o3 == 0 and onSegment(p2, p1, q2)):
        return True
    if (o4 == 0 and onSegment(p2, q1, q2)):
        return True

    return False

# def cross_wall3(begin, end):
#     pathLine = [(begin[0], begin[1]),(end[0], end[1])]
#     graphMap = [[[2,2],[4,4]]]*20+[[[0,1],[1,0]]]
#     for line in graphMap:
#         wall = line
#         if doIntersect(pathLine, wall):
#             return True
#     return False
