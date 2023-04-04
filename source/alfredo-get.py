#!/usr/bin/env python3
# -*- coding: utf-8 -*-

import requests
import json
from requests.structures import CaseInsensitiveDict
from datetime import datetime
import sys
from alfredo_fun import *
from config import TOKEN

# AlfreDo – a Todoist workflow
# using the todoist sync API
# Sunny ☀️   🌡️+18°F (feels +11°F, 59%) 🌬️↘7mph 🌗 2022-01-26 Wed 8:36AM



today = datetime.now().strftime("%Y-%m-%d")
#log (today)
MY_MODE = sys.argv[1]  # source: due today or overdue

url_sync = "https://api.todoist.com/sync/v8/sync"
url_stats = "https://api.todoist.com/sync/v8/completed/get_stats"

headers = CaseInsensitiveDict()
headers["Authorization"] = "Bearer " + TOKEN
headers["Content-Type"] = "application/x-www-form-urlencoded"

data = 'sync_token=*&resource_types=["items"]'

resp = requests.post(url_sync, headers=headers, data=data)
resp_stats = requests.get(url_stats, headers=headers)

log (resp)

myStats = resp_stats.json() #getting stats from API
mydata = resp.json() #getting data from API

log (myStats)

todays = [item for item in myStats['days_items'] if item['date'] == today]
#log (myData['goals']['daily_goal'])
SoFarCompleted = todays[0]['total_completed']

#log (myStats)

DailyGoal = myStats['goals']['daily_goal']
WeeklyGoal = myStats['goals']['weekly_goal']

TotalWeekCompleted = myStats['week_items'][0]['total_completed']
#log ("total week completed: " + str(TotalWeekCompleted))

if SoFarCompleted >= DailyGoal:
    statusDay = "✅"
else:
    statusDay = "❌"

if TotalWeekCompleted >= WeeklyGoal:
    statusWeek = "✅"
else:
    statusWeek = "❌"



myMatchCount=1
mydata=mydata['items']
dueDateItems = [task for task in mydata if task['due']] # selecting tasks with a due date
#dueDateItems = [task for task in dueDateItems if task['due']['date'] == today]



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

if dueDateItems:
    dueDateItems = sorted(dueDateItems, key = lambda i: i['due']['date']) #sorting by due date
    #log (len(dueDateItems))
    dueToday = len(dueDateItems) ## will need to figure this out f I want to show the number left

    for task in dueDateItems:
        if 'due' in task and task['due']['date'] <= today:
            myContent = task ['content'] 
            myDue = task ['due']['date']
            MYOUTPUT["items"].append({
            "title": myContent,
            "subtitle": myDue + "-"+ str(countR)+"/"+str(myMatchCount) + "-" + str(dueToday)+ " due today. Daily: " 
            + str(SoFarCompleted)+"/"+ str(DailyGoal)+statusDay+ " Weekly: " + str(TotalWeekCompleted)+"/"+ str(WeeklyGoal)+statusWeek , 
            "arg": str(task['id']) + ";;" + str(dueToday) 
            })
            countR += 1
        

    print (json.dumps(MYOUTPUT))
else: 
    MYOUTPUT["items"].append({
            "title": "no tasks left to do today 🙌",
            "subtitle": "Daily: " 
            + str(SoFarCompleted)+"/"+ str(DailyGoal)+statusDay+ " Weekly: " + str(TotalWeekCompleted)+"/"+ str(WeeklyGoal)+statusWeek , 
            "mods": {
    "shift": {
        
        "arg": "",
        "subtitle": "nothing to see here"
    },
    
},
            "arg": ""
            })
    print (json.dumps(MYOUTPUT))