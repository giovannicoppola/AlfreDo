#!/usr/bin/env python3

import requests
import json
from datetime import datetime, date, timedelta
import sys
import os
from config import TOKEN, MY_DATABASE, RefRate
import uuid
import re

"""

alfredo_ops
a library of functions for the AlfreDo Workflow
April 2023

"""



def generate_uuid():
    return str(uuid.uuid4())

# AlfreDo – a Todoist workflow
# Partly cloudy ⛅️  🌡️+31°F (feels +28°F, 82%) 🌬️↘4mph 🌗 2022-01-25 Tue 9:05AM

MY_COMMAND = sys.argv[1]  
MY_INPUT = sys.argv[2]


def log(s, *args):
    if args:
        s = s % args
    print(s, file=sys.stderr)

def get_project_name(projects, id):
    for project in projects:
        if project["id"] == id:
            return project["name"]
    return None

def get_project_id(projects, nameP):
    for project in projects:
        if project["name"] == nameP:
            return project["id"]
    return None

def fetchLabels(toShow):
     # getting all the tags (and counts) from the current subset of tasks
    label_counts = {}
    for item in toShow:
        for label in item.get('labels', []):
            if label in label_counts:
                label_counts[label] += 1
            else:
                label_counts[label] = 1
    
    # Define a function to return the value associated with the key in the label_counts dictionary
    def get_count(key):
        return label_counts[key]

    myLabels = list(label_counts.keys())
    # Sort the myLabels list based on the count associated with each label
    myLabels = sorted(myLabels, key=get_count, reverse=True)

    myLabels = ['@' + s for s in myLabels]
    #log (myLabelsAll)
    return label_counts, myLabels

def fetchProjects (toShow,myProjects):
     # # getting all the projects (and counts) from the current subset of tasks
    project_counts = {}
    for item in toShow:
        projN = [x['name'] for x in myProjects if x['id'] == item['project_id']][0]
        if projN in project_counts:
            project_counts[projN] += 1
        else:
            project_counts[projN] = 1
    
    def get_count(key):
        return project_counts[key]

    myProjectList = list(project_counts.keys())
    # Sort the myLabels list based on the count associated with each label
    myProjectList = sorted(myProjectList, key=get_count, reverse=True)

    
    
    myProjectList = ['#' + s for s in myProjectList]
    
    #log (project_counts)
    return project_counts,myProjectList

def completeTask (myTaskID):
    url = "https://api.todoist.com/sync/v9/sync"
    MY_UUID = generate_uuid()
    headers = {
        "Authorization": f"Bearer {TOKEN}",
    }

    data = {
        "commands": json.dumps([
            {
                "type": "item_close",
                "uuid": MY_UUID,
                "args": {
                    "id": myTaskID
                }
            }
        ])
    }

    response = requests.post(url, headers=headers, data=data)

    log(response.content)
    myResponse = response.json()
       
    if (myResponse['sync_status'][MY_UUID]) == "ok":
        print ("🎯 task completed!\nWell done 💪")
        getTodoistData()
    else:
        print ("❌ server error\ncheck debugger")

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

    myData = resp.json()
    with open(MY_DATABASE,'w') as myFile:
         json.dump(myData,myFile,indent=4)

    # prepare label and project file

def readTodoistData ():
    with open(MY_DATABASE,'r') as myFile:
         mydata = json.load(myFile)
    
    myTasks=mydata['items']
    myProjects=mydata['projects']
    myStats=mydata['stats']
    myUser=mydata['user']
    return myTasks, myProjects, myStats, myUser

def createNewTask (taskContent,taskLabels,taskProjectID,myDueDate):
    
    url = 'https://api.todoist.com/sync/v9/sync'
    MY_UUID = generate_uuid()
    MY_TEMPID = generate_uuid()
    taskLabels = taskLabels.split(',')
    
    headers = {
        "Authorization": f"Bearer {TOKEN}",
    }

    data = {
        "commands": json.dumps([
            {
                "type": "item_add",
                "uuid": MY_UUID,
                "temp_id": MY_TEMPID,
                "args": {
                    "content": taskContent,
                    "labels": taskLabels,
                    "project_id": taskProjectID,
                    "due": {"date": myDueDate}
                }
            }
        ])
    }

    response = requests.post(url, headers=headers, data=data)

    log(response.content)
    myResponse = response.json()
       
    if (myResponse['sync_status'][MY_UUID]) == "ok":
        print ("🎯 task created!\nWell done.")
        getTodoistData()
    else:
        print ("❌ server error\ncheck debugger")
    

def parseNewTask (myInput):
    # fetching label data from todoist
    allTasks, myProjects, myStats, myUser = readTodoistData()
    label_counts, myLabelsAll = fetchLabels(allTasks)
    project_counts,myProjectListAll = fetchProjects(allTasks,myProjects)

    myTags = []
    
    
    MYOUTPUT = {"items": []}
    myInputElements = myInput.split()
    #finalInputItems = myInputElements
    #FINAL_INPUT = " ".join(finalInputItems)
    
    
    for myInputItem in myInputElements:
        
        if myInputItem.strip() in myLabelsAll: # is this a real tag? 
            
            myTags.append (myInputItem[1:])    
        
        elif myInputItem.strip() in myProjectListAll: # is this a real project? 
            taskProjectName = myInputItem
            taskProjectID = get_project_id (myProjects,myInputItem[1:])
            
        elif myInputItem.startswith('@'): #user trying to add a label
            
            mySubset = [i for i in myLabelsAll if myInputItem.casefold() in i.casefold()]
            
            
            # adding a complete tag if the user selects it from the list
            if mySubset:
                myInputElements.remove(myInputItem)
                myInput = " ".join(myInputElements)
                
                for thislabel in mySubset:
                    if myInput:
                        MY_ARG = f"{myInput} {thislabel} "
                        
                    else:
                        MY_ARG = f"{thislabel} "
                    
                    MYOUTPUT["items"].append({
                    "title": f"{thislabel} ({label_counts[thislabel[1:]]})",
                    "subtitle": MY_ARG,
                    "arg": MY_ARG,
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
                "arg": MY_INPUT+" ",
                "icon": {
                        "path": f"icons/Warning.png"
                    }
                })
            print (json.dumps(MYOUTPUT))
            
            exit()
    
        elif myInputItem.startswith('#'): #user trying to add a project
            
            mySubset = [i for i in myProjectListAll if myInputItem.casefold() in i.casefold()]
            
            
            # adding a complete project name if the user selects it from the list
            if mySubset:
                myInputElements.remove(myInputItem)
                myInput = " ".join(myInputElements)
                
                for thisProj in mySubset:
                    if myInput:
                        MY_ARG = f"{myInput} {thisProj} "
                    else:
                        MY_ARG = f"{thisProj} "
                    MYOUTPUT["items"].append({
                    "title": f"{thisProj} ({project_counts[thisProj[1:]]})",
                    "subtitle": MY_ARG,
                    "arg": MY_ARG,
                    "variables" : {
                        
                        },
                    "icon": {
                            "path": f"icons/project.png"
                        }
                    })
            else:
                MYOUTPUT["items"].append({
                "title": "no projects matching",
                "subtitle": "try another query?",
                "arg": "",
                 "variables" : {
                    
                    "myArg": MY_INPUT+" "
                    },
                "icon": {
                        "path": f"icons/Warning.png"
                    }
                })
            print (json.dumps(MYOUTPUT))
            exit()
        
        
        
        elif myInputItem.startswith ('due:'):
            # check first if there is a due date already 
            patternDue = r'due:(\d+)d'
            matchDue = re.search(patternDue, myInputItem)
                        
            if matchDue:
               dueString = getNewDate (int(matchDue.group(1)),F=False)
               
               
            else:
                customDays = myInputItem.split(':')[1]
                myInput = " ".join(myInputElements)
                dueMenu(customDays,myInput)
                exit()
            

    MYOUTPUT = {"items": []}
    myTaskElements = myInput.split()
    
    for xxx in myTaskElements[:]:
        if xxx.startswith('@') or xxx.startswith('#') or xxx.startswith('due:'):
            myTaskElements.remove(xxx)
        
    
    MY_TASK_TEXT = " ".join (myTaskElements)
    
    myTagString = ",".join(myTags)
    if myTagString:
        myTagStringF = f"🏷️{myTagString}"
    else:
        myTagStringF = ''
    
    try:
        dueStringF = f"🗓️ due:{dueString}"
    except NameError:
        dueString = ""
        dueStringF = ""

    try:
        myProjStringF = f"📋{taskProjectName}"
        
    except NameError:
        taskProjectName = "#Inbox"
        myProjStringF = f"📋{taskProjectName}"
        taskProjectID = get_project_id (myProjects,taskProjectName[1:])


    MYOUTPUT["items"].append({
                "title": MY_TASK_TEXT,
                "subtitle": f"{myProjStringF} {myTagStringF} {dueStringF} ⇧↩️ to create",
                "arg": myInput,
                "variables" : {
                "myTaskText": MY_TASK_TEXT,
                "myTagString": myTagString,
                "myProjectID": taskProjectID,
                "myDueDate": dueString
                    },
                "icon": {
                        "path": f"icons/newTask.png"
                    }
                })
   
    print (json.dumps(MYOUTPUT))
    

def checkingTime ():
## Checking if the database needs to be built or rebuilt
    timeToday = date.today()
    if not os.path.exists(MY_DATABASE):
        log ("Database missing ... building ⏳")
        getTodoistData()
        
    else: 
        databaseTime= (int(os.path.getmtime(MY_DATABASE)))
        dt_obj = datetime.fromtimestamp(databaseTime).date()
        time_elapsed = (timeToday-dt_obj).days
        log (f"{time_elapsed} days from last update")
        if time_elapsed >= RefRate:
            log ("rebuilding database ⏳...")
            getTodoistData()
            log ("done 👍")
            

def getNewDate(myDays, F = True):
    timeToday = date.today()
    delta = timedelta(days=myDays)
    finalTime = timeToday+delta
    finalTimeF = finalTime.strftime("%A, %B %d, %Y")
    finalTimeFS = finalTime.strftime("%Y-%m-%d")
    if F == True:
        return finalTimeF
    else:
        return finalTimeFS

def reschMenu(customDays):
    taskContent = os.getenv('myTaskContent')
    pattern = r'^(\d+)([wm]?)$'
    match = re.match(pattern, customDays)

    
    MYOUTPUT = {"items": []}
    
    if customDays == '':
        MYOUTPUT["items"].extend([{
                        "title": f"Reschedule to today 🗓️ {getNewDate(0)} 🔥",
                        "subtitle": taskContent,
                        "arg": 0,
                        "variables" : {
                            },
                        "icon": {
                                "path": f"icons/today.png"
                            }
                        },
                        {
                        "title": f"Reschedule to tomorrow 🗓️ {getNewDate(1)} 🧨",
                        "subtitle": taskContent,
                        "arg": 1,
                        "variables" : {
                            },
                        "icon": {
                                "path": f"icons/today.png"
                            }
                        },
                        {
                        "title": f"Reschedule in a week 🗓️ {getNewDate(7)} 🍹",
                        "subtitle": taskContent,
                        "arg": 7,
                        "variables" : {
                            },
                        "icon": {
                                "path": f"icons/today.png"
                            }
                        },
                        {
                        "title": f"Reschedule in a month 🗓️ {getNewDate(30)} 🏖️",
                        "subtitle": taskContent,
                        "arg": 30,
                        "variables" : {
                            },
                        "icon": {
                                "path": f"icons/today.png"
                            }
                        }]
                        
                        )
        
    elif match:
        num_str, letter = match.groups()
        num = int(num_str)
        if letter == 'w':
            num *= 7
        elif letter == 'm':
            num *= 30
        log(f'Valid input:{num}')
        if num ==1:
            dayString = 'day'
        else:
            dayString = 'days'
        MYOUTPUT["items"].append({
                        "title": f"Reschedule in {num:,} {dayString} 🗓️ {getNewDate(num)}",
                        "subtitle": taskContent,
                        "arg": num,
                        "variables" : {
                            },
                        "icon": {
                                "path": f"icons/today.png"
                            }
                        })
    
    else:
        MYOUTPUT["items"].append({
                        "title": f"Incorrect format!",
                        "subtitle": "enter an integer (days) or add 'w' (weeks) or 'm' (months)",
                        "arg": '',
                        "variables" : {
                            },
                        "icon": {
                                "path": f"icons/warning.png"
                            }
                        })
    print (json.dumps(MYOUTPUT))


def dueMenu(customDays,inputThrough):
    
    pattern = r'^(\d+)([wm]?)$'
    match = re.match(pattern, customDays)

    patternDue = r'(?<=due:)\d*[wm]?'
    
    inputThroughF = re.sub(patternDue, '', inputThrough)
    

    MYOUTPUT = {"items": []}
    
    if customDays == '':
        MYOUTPUT["items"].extend([{
                        "title": f"Due today 🗓️ {getNewDate(0)} 🔥",
                        "subtitle": '',
                        "arg": f"{inputThroughF}0d ",
                        "variables" : {
                            },
                        "icon": {
                                "path": f"icons/today.png"
                            }
                        },
                        {
                        "title": f"Due tomorrow 🗓️ {getNewDate(1)} 🧨",
                        "subtitle": '',
                        "arg": f"{inputThroughF}1d ",
                        "variables" : {
                            },
                        "icon": {
                                "path": f"icons/today.png"
                            }
                        },
                        {
                        "title": f"Due in a week 🗓️ {getNewDate(7)} 🍹",
                        "subtitle": '',
                        "arg": f"{inputThroughF}7d ",
                        "variables" : {
                            },
                        "icon": {
                                "path": f"icons/today.png"
                            }
                        },
                        {
                        "title": f"Due in a month 🗓️ {getNewDate(30)} 🏖️",
                        "subtitle": '',
                        "arg": f"{inputThroughF}30d ",
                        "variables" : {
                            },
                        "icon": {
                                "path": f"icons/today.png"
                            }
                        }]
                        
                        )
        
    elif match:
        num_str, letter = match.groups()
        num = int(num_str)
        if letter == 'w':
            num *= 7
        elif letter == 'm':
            num *= 30
        log(f'Valid input:{num}')
        if num ==1:
            dayString = 'day'
        else:
            dayString = 'days'
        MYOUTPUT["items"].append({
                        "title": f"Due in {num:,} {dayString} 🗓️ {getNewDate(num)}",
                        "subtitle": '',
                        "arg": f"{inputThroughF}{num}d ",
                        "variables" : {
                            },
                        "icon": {
                                "path": f"icons/today.png"
                            }
                        })
    
    else:
        MYOUTPUT["items"].append({
                        "title": f"Incorrect format!",
                        "subtitle": "enter an integer (days) or add 'w' (weeks) or 'm' (months)",
                        "arg": '',
                        "variables" : {
                            },
                        "icon": {
                                "path": f"icons/warning.png"
                            }
                        })
    print (json.dumps(MYOUTPUT))


def rescheduleTask (days,taskID):
    newDate = getNewDate(days,F=False)
    log (f"days to reschedule: {days}, {taskID}, {newDate}")

    url = "https://api.todoist.com/sync/v9/sync"
    MY_UUID = generate_uuid()
    headers = {
        "Authorization": f"Bearer {TOKEN}",
    }

    data = {
        "commands": json.dumps([
            {
                "type": "item_update",
                "uuid": MY_UUID,
                "args": {
                    "id": taskID,
                    "due": {"date": newDate }
                }
            }
        ])
    }

    response = requests.post(url, headers=headers, data=data)

    log(response.content)
    myResponse = response.json()
       
    if (myResponse['sync_status'][MY_UUID]) == "ok":
        print ("🎯 task rescheduled!\nGet to work!😅")
        getTodoistData()
    else:
        print ("❌ server error\ncheck debugger")

def forceRebuild():  
    log ("rebuilding database ⏳...")
    log ("done 👍")
    getTodoistData()
    result= {"items": [{
        "title": "Done!" ,
        "subtitle": "ready to use AlfreDo now ✅",
        "arg": "",
        "icon": {

                "path": "icons/done.png"
            }
        }]}
    print (json.dumps(result))

    

def main():
    if MY_COMMAND == "complete":
        completeTask (MY_INPUT)

    if MY_COMMAND == "findDate":
        customDays = sys.argv[3] 
        reschMenu (customDays)

    if MY_COMMAND == "reschedule":
        daysReschedule = int(sys.argv[2] )
        myTaskID = os.getenv('myTaskID')
        rescheduleTask(daysReschedule,myTaskID)

    if MY_COMMAND == "rebuild":
        forceRebuild()


    if MY_COMMAND == "newTask":
        taskText = os.getenv('myTaskText')
        taskLabels = os.getenv('myTagString')
        taskProjectID = os.getenv('myProjectID')
        myDueDate = os.getenv('myDueDate')
        
        createNewTask (taskText,taskLabels,taskProjectID, myDueDate)

    if MY_COMMAND == "parse":
     
        parseNewTask (MY_INPUT)


if __name__ == '__main__':
    main ()
