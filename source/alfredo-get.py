#!/usr/bin/env python3

import requests
import json
from requests.structures import CaseInsensitiveDict
from datetime import datetime
import sys

# AlfreDo – a Todoist workflow
# Chappaqua – Sunny ☀️   🌡️+18°F (feels +11°F, 59%) 🌬️↘7mph 🌗 2022-01-26 Wed 8:36AM
# new version of `alfredo-old` using the sync API
today = datetime.utcnow().strftime("%Y-%m-%d")
MY_MODE = sys.argv[1]  
url_sync = "https://api.todoist.com/sync/v8/sync"
url_stats = "https://api.todoist.com/sync/v8/completed/get_stats"

headers = CaseInsensitiveDict()
headers["Authorization"] = "Bearer f919fc636de42e7966cfa27c51742e6e0a1e4ef9"
headers["Content-Type"] = "application/x-www-form-urlencoded"

#data = 'sync_token=* & resource_types=["stats"]'

#data = 'sync_token=*&resource_types=["projects"]'
data = 'sync_token=*&resource_types=["items"]'

resp = requests.post(url_sync, headers=headers, data=data)
resp_stats = requests.get(url_stats, headers=headers)
myStats = resp_stats.json()
mydata = resp.json()
#resp = requests.post(url, headers=headers, data=data)
#print (myData['days_items'])

todays = [item for item in myStats['days_items'] if item['date'] == today]
#print (myData['goals']['daily_goal'])
SoFarCompleted = todays[0]['total_completed']

DailyGoal = myStats['goals']['daily_goal']

if SoFarCompleted >= DailyGoal:
    statusDay = "✅"
else:
    statusDay = "❌"

### WILL ADD STATUS WEEK 


#print(mydata['sync_token'])
myMatchCount=1
mydata=mydata['items']
dueDateItems = [task for task in mydata if task['due']] # selecting tasks with a due date
#dueDateItems = [task for task in dueDateItems if task['due']['date'] == today]


#print(dueDateItems)
#print (len(mydata))
#print(len(dueDateItems))
#print(myMatchCount)
#print(mydata['items'][0]['due']['date'])
#print(resp.status_code)

MYOUTPUT = {"items": []}
countR=1
myMatchCount=1

for task in dueDateItems:  #counting the total number of tasks due
    if task['due']['date'] <= today:
        myMatchCount+=1
      #print (task)



if MY_MODE == "today":
    dueDateItems = [task for task in dueDateItems if task['due']['date'] == today]
else:
    dueDateItems = [task for task in dueDateItems if task['due']['date'] <= today]


dueDateItems = sorted(dueDateItems, key = lambda i: i['due']['date']) #sorting by due date
#print (len(dueDateItems))
dueToday = len(dueDateItems) ## will need to figure this out f I want to show the number left

for task in dueDateItems:
    if 'due' in task and task['due']['date'] <= today:
        myContent = task ['content'] 
        myDue = task ['due']['date']
        MYOUTPUT["items"].append({
        "title": myContent,
        "subtitle": myDue + "-"+ str(countR)+"/"+str(myMatchCount) + "-" + str(dueToday)+ " due today. Daily goal: " 
        + str(SoFarCompleted)+"/"+ str(DailyGoal)+statusDay, 
        "arg": str(task['id']) + ";;" + str(dueToday) 
        })
        countR += 1
    

print (json.dumps(MYOUTPUT))


"""
#### DOWNLOAD USER STATS
import requests
from requests.structures import CaseInsensitiveDict

url = "https://api.todoist.com/sync/v8/completed/get_stats"

headers = CaseInsensitiveDict()
headers["Authorization"] = "Bearer f919fc636de42e7966cfa27c51742e6e0a1e4ef9"


resp = requests.get(url, headers=headers)

print(resp.status_code)

"""