#!/usr/bin/env python3
# -*- coding: utf-8 -*-


# AlfreDo – a Todoist workflow
# using the todoist sync API
# Sunny ☀️   🌡️+18°F (feels +11°F, 59%) 🌬️↘7mph 🌗 2022-01-26 Wed 8:36AM
# restarted on 
# Partly cloudy ⛅️  🌡️+64°F (feels +64°F, 70%) 🌬️↖9mph 🌕&m Tue Apr  4 18:23:14 2023


import requests
import json
from datetime import datetime, date
import sys
from alfredo_fun import log
from config import TOKEN, MY_DATABASE, RefRate
import re
import os

myFilterLabels = []



def checkingTime ():
## Checking if the database needs to be built or rebuilt
    timeToday = date.today()
    if not os.path.exists(MY_DATABASE):
        log ("Database missing ... building")
        getTodoistData()
        
    else: 
        databaseTime= (int(os.path.getmtime(MY_DATABASE)))
        dt_obj = datetime.fromtimestamp(databaseTime).date()
        time_elapsed = (timeToday-dt_obj).days
        log (str(time_elapsed)+" days from last update")
        if time_elapsed >= RefRate:
            log ("rebuilding database ⏳...")
            getTodoistData()
            log ("done 👍")



def getTodoistData ():
    
    url = 'https://api.todoist.com/sync/v9/sync'
    headers = {
        'Authorization': f'Bearer {TOKEN}'
    }
    data = {
        'sync_token': '*',
        'resource_types': '["all"]'
    }

    resp = requests.post(url, headers=headers, data=data)

    
    with open(MY_DATABASE,'w') as myFile:
         json.dump(resp.json(),myFile,indent=4)

    mydata = resp.json() #getting data from API
    myTasks=mydata['items']
    myProjects=mydata['projects']
    myStats=mydata['stats']
    myUser=mydata['user']
    return myTasks, myProjects, myStats, myUser

def get_project_name(projects, id):
    for project in projects:
        if project["id"] == id:
            return project["name"]
    return None



def main():
    today = datetime.now().strftime("%Y-%m-%d")
    todayDate = datetime.today()
    

    MYOUTPUT = {"items": []}
    allTasks, myProjects, myStats, myUser = getTodoistData()
    
    dueTasks = [task for task in allTasks if task['due']] # selecting tasks with a due date
    dueTodayTasks = [task for task in dueTasks if task['due']['date'] == today]
    overdueTasks = [task for task in dueTasks if task['due']['date'] < today]

    today = datetime.now().strftime("%Y-%m-%d")
    MY_MODE = sys.argv[1]  # source: due today, all, overdue
    MY_INPUT = sys.argv[2]  # source: due today, all, overdue

    #tasks completed today
    todays = [item for item in myStats['days_items'] if item['date'] == today]
    SoFarCompleted = todays[0]['total_completed']
    

    DailyGoal = myUser['daily_goal']
    WeeklyGoal = myUser['weekly_goal']

    TotalWeekCompleted = myStats['week_items'][0]['total_completed']
    

    if SoFarCompleted >= DailyGoal:
        statusDay = "✅"
    else:
        statusDay = "❌"

    if TotalWeekCompleted >= WeeklyGoal:
        statusWeek = "✅"
    else:
        statusWeek = "❌"



    countR=1
    myMatchCount=1
    


    
    for task in overdueTasks:  #counting the total number of tasks due
        if task['due']['date'] <= today:
            myMatchCount+=1
        #print (task)


    if MY_MODE == "today":
        toShow = dueTodayTasks
        if toShow:
            toShow = sorted(toShow, key = lambda i: i['due']['date']) #sorting by due date
        MYICON = 'icons/today.png'
        

    elif MY_MODE == "due":
        toShow = overdueTasks
        if toShow:
            toShow = sorted(toShow, key = lambda i: i['due']['date']) #sorting by due date
        MYICON = 'icons/overdue.png'
        

    elif MY_MODE == "all":
        toShow = allTasks
        MYICON = 'icons/logo1.png'

    dueToday = len(dueTodayTasks)
    if toShow:
        
        myLabels = set()
        for item in toShow:
            myLabels.update(set(item.get('labels', [])))
        myLabels = ['#' + s for s in myLabels]
        
        log (myLabels)

        # extracting any full tags from current input, adding them to the list to filter
        fullTags = re.findall('#[^ ]+ ', MY_INPUT)
        fullTags = [s.strip() for s in fullTags]
        mySearchInput = MY_INPUT.strip()
    
        for currTag  in fullTags:
            if currTag.strip() in myLabels: #if it is a real tag
                myFilterLabels.append(currTag[1:].strip())
                mySearchInput = re.sub(currTag, '', mySearchInput).strip()
                
        
        # check if the user is trying to enter a tag
        MYMATCH = re.search(r'(?:^| )#[^ ]*$', MY_INPUT)
        if (MYMATCH !=None):
            
            MYFLAG = MYMATCH.group(0).lstrip(' ')
            
            MY_INPUT = re.sub(MYFLAG,'',MY_INPUT)
            
            mySubset = [i for i in myLabels if MYFLAG in i]
            
            # adding a complete tag if the user selects it from the list
            if mySubset:
                for thislabel in mySubset:
                    
                    MYOUTPUT["items"].append({
                    "title": thislabel,
                    "subtitle": MY_INPUT,
                    "arg": MY_INPUT+thislabel+" ",
                    "variables" : {
                        
                        },
                    "icon": {
                            "path": f"icons/label.png"
                        }
                    })
            else:
                MYOUTPUT["items"].append({
                "title": "no labels matching",
                "subtitle": "try another query?",
                "arg": " ",
                "icon": {
                        "path": f"icons/Warning.png"
                    }
                })
            print (json.dumps(MYOUTPUT))
        else:
            toShow = [item for item in toShow if (all(label in item.get('labels', []) for label in myFilterLabels)) and all(substring.casefold() in item['content'].casefold() for substring in mySearchInput.split())]

        
            for task in toShow:
                
                myContent = task ['content'] 
                
                if MY_MODE != 'all':
                    dueDate =datetime.strptime(task ['due']['date'] , '%Y-%m-%d')
                    myDue = dueDate if task ['due'] else ""
                    if myDue:
                        dueDays = todayDate - myDue
                    else: 
                        dueDays = ''
                if task['labels']:
                    
                    myLabelsString = f"🏷️ {','.join(task['labels'])}"
                else:
                    myLabelsString = ""
                
                
                myProjectName = get_project_name(myProjects, task['project_id'])
        

                MYOUTPUT["items"].append({
                "title": f"{myContent} – ({myProjectName}) {dueDays.days:,} days overdue❗",
                
                "subtitle": str(dueDate) + "-"+ str(countR)+"/"+str(myMatchCount) + "-" + str(dueToday)+ " due today. Daily: " 
                + str(SoFarCompleted)+"/"+ str(DailyGoal)+statusDay+ " Weekly: " + str(TotalWeekCompleted)+"/"+ str(WeeklyGoal)+statusWeek + myLabelsString, 
                
                "icon": {
                        "path": MYICON
                    },
                
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

if __name__ == '__main__':
    main ()
