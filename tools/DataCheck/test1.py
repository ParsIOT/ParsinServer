import numpy as np
import matplotlib.pyplot as plt

# x = [102.2395906430023, 289.86942955834525, 497.50935704959943, 688.7344027686181, 891.2222786656959, 1087.6561975854852, 1295.553232945087, 1487.035329186388, 1682.8249379169683]
# y = [-64.79, -66.53, -70.66, -83.54, -73.75, -77.1, -82.86, -81.75, -81.93]

x = [1, 2, 3, 4, 5, 6, 7, 8, 9]
y = [-43, -45, -47, -48, -49, -51, -52, -52, -53]


y1 = []
for i in y:
    y1.append(i)

y= y1

fit = np.polyfit(x,y,1)
fit_fn = np.poly1d(fit) 
# fit_fn is now a function which takes in x and returns an estimate for y

plt.plot(x,y, x, fit_fn(x))
# plt.xlim(0, 5)
# plt.ylim(0, 12)
# plt.xscale('log')
plt.axis([0, x[-1], -80 ,-30])
plt.show()