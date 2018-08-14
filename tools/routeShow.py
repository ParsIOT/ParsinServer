import json

import requests
url = 'http://104.237.255.199:18003/location?group=arman_28_3_97_ble_1'+"&n=1000"
resp = requests.get(url=url)
bigJson = resp.json()
# print(bigJson)
# file = open("trackExample", "r")
# bigStr = file.read()
# j = json.loads('{"one" : "1", "two" : "2", "three" : "3"}')
# bigJson = json.loads(bigStr)
dots = []
for res in bigJson['users']['hadi']:
    # print(res['location'])
    dots.append(res['location'])

for dot in dots:
    # print(dot)
    print("addMarker("+dot+")")

for i in range(len(dots)-1):
    print("line("+dots[i]+","+dots[i+1]+")")


