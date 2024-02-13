#!/usr/bin/env python3

import requests
import json
from datetime import datetime, date, timedelta
import sys
import os
from config import TOKEN, MY_DATABASE, RefRate, MY_LABEL_COUNTS, MY_PROJECT_COUNTS, PARTIAL_MATCH
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

def get_parent_project_name(sections, projects, id):
    for section in sections:
        if section["id"] == id:
            myProjectID = section["project_id"]
    get_project_name (projects,myProjectID)
    return None


def get_project_id(projects, nameP):
    for project in projects:
        if project["name"] == nameP:
            return project["id"]
    return None

def get_section_id(projects, sections, nameS):
    myProj = nameS.split("/")[0][1:]
    log (myProj)
    myProjID = get_project_id(projects, myProj)
    mySect = nameS.split("/")[1]
    for section in sections:
        if section["name"] == mySect and section["project_id"] == myProjID:
            return section["id"]
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

def fetchAllActiveLabels(myTasks,myLabels):
     # getting all the tags (and counts) from the complete list of active tags
    label_counts = {}
    for item in myTasks:
        for label in item.get('labels', []):
            if label in label_counts:
                label_counts[label] += 1
            else:
                label_counts[label] = 1
    for mySingleLabel in myLabels:
        if mySingleLabel['is_deleted'] == False:
            if mySingleLabel['name'] not in label_counts: # assigning 0 to all other labels 
                    label_counts[mySingleLabel['name']] = 0
    

    # Define a function to return the value associated with the key in the label_counts dictionary
    #def get_count(key):
    #    return label_counts[key]

    #myLabels = list(label_counts.keys())
    # Sort the myLabels list based on the count associated with each label
    #myLabels = sorted(myLabels, key=get_count, reverse=True)
    label_counts = dict(sorted(label_counts.items(), key=lambda x: x[1], reverse=True))

    #myLabels = ['@' + s for s in myLabels]
    #log (myLabelsAll)
    return label_counts

def fetchAllActiveProjects (myTasks,myProjects,mySections):
     # # getting all the projects (and counts) from the current subset of tasks
    project_counts = {}
    for item in myTasks:
        projN = [x['name'] for x in myProjects if x['id'] == item['project_id']][0]
        if item['section_id']:
            sectN = [x['name'] for x in mySections if x['id'] == item['section_id']][0]
            projN = f"{projN}/{sectN}"
            #log (f"found section, {item['section_id']}: {sectN}")
        if projN in project_counts:
            project_counts[projN] += 1
        else:
            project_counts[projN] = 1
    
    for mySingleProject in myProjects:
        if mySingleProject['is_deleted'] == False and mySingleProject['is_archived'] == False:
            if mySingleProject['name'] not in project_counts: # assigning 0 to all other labels 
                    project_counts[mySingleProject['name']] = 0
    for mySingleSection in mySections:
        myProjectName = get_project_name (myProjects,mySingleSection['project_id'])
        completeName = f"{myProjectName}/{mySingleSection['name']}"
        if completeName not in project_counts:
            project_counts[completeName] = 0

    
    project_counts = dict(sorted(project_counts.items(), key=lambda x: x[1], reverse=True))

    
    # myProjectList = list(project_counts.keys())
    # # Sort the myLabels list based on the count associated with each label
    # myProjectList = sorted(myProjectList, key=get_count, reverse=True)

    
    
    # myProjectList = ['#' + s for s in myProjectList]
    
    #log (project_counts)
    return project_counts


def fetchSections (toShow,mySections,myProjects):
     # # getting all the projects (and counts) from the current subset of tasks
    section_counts = {}
    section_parentProjects = {}
    for item in toShow:
        if item['section_id']:
            sectN = [x['name'] for x in mySections if x['id'] == item['section_id']][0]
            sectpID = [x['project_id'] for x in mySections if x['id'] == item['section_id']][0]
            parentProjectName = get_project_name(myProjects,sectpID)
            section_parentProjects[sectN] = parentProjectName
            if sectN in section_counts:
                section_counts[sectN] += 1
            else:
                section_counts[sectN] = 1
    
    def get_count(key):
        return section_counts[key]

    mySectionList = list(section_counts.keys())
    # Sort the myLabels list based on the count associated with each label
    mySectionList = sorted(mySectionList, key=get_count, reverse=True)

    
    
    mySectionList = ['^' + s for s in mySectionList]
    
    #log (project_counts)
    return section_counts,mySectionList,section_parentProjects


def fetchProjects (toShow,myProjects,mySections):
     # # getting all the projects (and counts) from the current subset of tasks
    project_counts = {}
    for item in toShow:
        projN = [x['name'] for x in myProjects if x['id'] == item['project_id']][0]
        if item['section_id']:
            sectN = [x['name'] for x in mySections if x['id'] == item['section_id']][0]
            projN = f"{projN}/{sectN}"
            #log (f"found section, {item['section_id']}: {sectN}")
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

    # prepare label file
    myAllLabelCounts = fetchAllActiveLabels (myData['items'],myData['labels'])

    with open(MY_LABEL_COUNTS,'w') as myFile:
         json.dump(myAllLabelCounts,myFile,indent=4)

    # prepare project file
    myAllProjectCounts = fetchAllActiveProjects (myData['items'],myData['projects'],myData['sections'])
    with open(MY_PROJECT_COUNTS,'w') as myFile:
         json.dump(myAllProjectCounts,myFile,indent=4)

def readTodoistData ():
    with open(MY_DATABASE,'r') as myFile:
         mydata = json.load(myFile)
    
    myTasks=mydata['items']
    myProjects=mydata['projects']
    myStats=mydata['stats']
    mySections=mydata['sections']
    myUser=mydata['user']
    return myTasks, mySections, myProjects, myStats, myUser

def createNewTask (taskContent,taskLabels,taskProjectID,taskSectionID,myDueDate, myPriority):
    
    url = 'https://api.todoist.com/sync/v9/sync'
    MY_UUID = generate_uuid()
    MY_TEMPID = generate_uuid()
    
    log (f"task labels: {taskLabels}")
    
    taskLabels = taskLabels.split(',,..,,')
    log (f"task labels: {taskLabels}")

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
                    "section_id": taskSectionID,
                    "due": {"date": myDueDate},
                    "priority": myPriority
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
    

def createLabel (myNewLabel):
    """
    a function to create a new todoist label
    """

    url = 'https://api.todoist.com/sync/v9/sync'
    MY_UUID = generate_uuid()
    MY_TEMPID = generate_uuid()
    
    
    headers = {
        "Authorization": f"Bearer {TOKEN}",
    }

    data = {
        "commands": json.dumps([
            {
                "type": "label_add",
                "uuid": MY_UUID,
                "temp_id": MY_TEMPID,
                "args": {
                    "name": myNewLabel
                    
                }
            }
        ])
    }

    response = requests.post(url, headers=headers, data=data)

    log(response.content)
    myResponse = response.json()
       
    if (myResponse['sync_status'][MY_UUID]) == "ok":
        log ("🎯 label added")
        
        # adding the new label to the label list to allow task creation (the list will be overwritten when the task is created)

        # Read the labels file
        with open(MY_LABEL_COUNTS, 'r') as file:
            data = json.load(file)

        # Add the new label
        data[myNewLabel] = 0

        # Write the updated data back to the file
        with open(MY_LABEL_COUNTS, 'w') as file:
            json.dump(data, file, indent=4)

        

    else:
        log ("❌ server error")
        exit()
    
    
    

def parseNewTask (myInput):
    # fetching saved data and counts 
    allTasks, mySections, myProjects, myStats, myUser = readTodoistData()
    
    
    ## LABELS
    with open(MY_LABEL_COUNTS,'r') as myFile:
         label_counts = json.load(myFile)
    myLabelListAll = list (label_counts)
    myLabelListAll = ['@' + s for s in myLabelListAll]
    
    
    section_counts,mySectionListAll, section_ParentProjects = fetchSections(allTasks,mySections,myProjects)
    myTags = []
    taskProjectName = ''
    
    ## PROJECTS
    with open(MY_PROJECT_COUNTS,'r') as myFile:
         project_counts = json.load(myFile)
    myProjectListAll = list (project_counts)
    myProjectListAll = ['#' + s for s in myProjectListAll]
    

    # log (f"mySections all: {mySections}")
    # log (f"myProjects all: {myProjects}")
    MYOUTPUT = {"items": []}
    
    def parseInput(MY_INPUT):
    
        pattern = r'\s*([@#]\([^)]+\)|\S+)\s*' #keeps together elements with space if they are in parenthese and preceded by # or @
        result = re.findall(pattern, MY_INPUT)
    
        return (result)
    
    myInputElements = parseInput (myInput)
    log (f"input elements: {myInputElements}")
    
    for myInputItem in myInputElements:
        #log (f" input item: {myInputItem}")
        
        if myInputItem.startswith('@'): #user trying to add a label
            if myInputItem.startswith("@(") and myInputItem.endswith(")") and " " in myInputItem: #there is a space and AlfreDO introduced parentheses 
                    myInputItem = myInputItem.replace("(","",1)
                    myInputItem = myInputItem.replace(")","")
                    myInputItem = myInputItem.strip()
            
            if myInputItem.strip() in myLabelListAll: # is this a real label? 
                
                myTags.append (myInputItem[1:]) 
        
            else:
                if PARTIAL_MATCH == 1:
                    mySubset = [i for i in myLabelListAll if myInputItem[1:].casefold() in i.casefold()]
                else:
                    mySubset = [i for i in myLabelListAll if myInputItem.casefold() in i.casefold()]    
                    
                # adding a complete tag if the user selects it from the list
                if mySubset:
                    myInputElements.remove(myInputItem)
                    myInput = " ".join(myInputElements)
                    
                    for thislabel in mySubset:
                        if " " in thislabel: #adding parentheses if there is a space in the project name
                            thisLabel_string = f"@({thislabel[1:]})"
                        else:
                            thisLabel_string = thislabel
                        if myInput:
                            MY_ARG = f"{myInput} {thisLabel_string} "
                            
                        else:
                            MY_ARG = f"{thisLabel_string} "
                        
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
                    "title": f"no labels matching, create a new label named '{myInputItem[1:]}'?",
                    "subtitle": "↩️ to create a new label 🏷️",
                    "arg": f"{myInput} ",
                    "variables" : {
                        "mySource": "createLabel",
                        "myNewLabel": myInputItem[1:],
                            },
                    "icon": {
                            "path": f"icons/newLabel.png"
                        }
                    })
                print (json.dumps(MYOUTPUT))
                
                exit()
        
            
        
        elif myInputItem.startswith('#'): # user trying to enter a project 
            # at the time of the first version of the workflow, parentheses and other special characters were not allowed in project names, and I used them to allow spaces in Alfred's window
            # in July 2023 they were allowed.
            
            if myInputItem.startswith("#(") and myInputItem.endswith(")") and " " in myInputItem: #there is a space and AlfreDO introduced parentheses 
                myInputItem = myInputItem.replace("(","",1)
                myInputItem = myInputItem.replace(")","") #there will be one closing parenthesis only
                myInputItem = myInputItem.strip()
            
            
            if myInputItem.strip() in myProjectListAll: # is this a real project? 
                if "/" in myInputItem:
                    taskProjectName = myInputItem.split("/")[0]
                    taskSectionName = myInputItem.split("/")[1]
                    taskProjectID = get_project_id (myProjects,taskProjectName[1:])
                    taskSectionID = get_section_id (myProjects,mySections, myInputItem)
                    log (f"project name: {taskProjectName}, section name: {taskSectionName}, projectID: {taskProjectID}, sectionID: {taskSectionID}")
                else:
                    taskProjectName = myInputItem
                    taskProjectID = get_project_id (myProjects,myInputItem[1:])
                    
       
            else: #user trying to add a project 
                
                if PARTIAL_MATCH == 1:
                    mySubset = [i for i in myProjectListAll if myInputItem[1:].casefold() in i.casefold()]
                else:
                    mySubset = [i for i in myProjectListAll if myInputItem.casefold() in i.casefold()]
                #log (f"INPUT ITEM: {myInputItem}")
                
                # adding a complete project name if the user selects it from the list
                if mySubset:
                    
                    myInputElements.remove(myInputItem)
                    myInput = " ".join(myInputElements)
                    
                    for thisProj in mySubset:
                        if " " in thisProj:
                            thisProj_string = f"#({thisProj[1:]})"
                        else:
                            thisProj_string = thisProj
                        if myInput:
                            MY_ARG = f"{myInput} {thisProj_string} "
                        else:
                            MY_ARG = f"{thisProj_string} "
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
            
    
        # priority
        elif myInputItem.strip().casefold() in ['p1','p2','p3','p4']: #user trying to enter a priority
            #log (f"priority {myInputItem}")
            prioString = myInputItem.strip().casefold()


        elif myInputItem.startswith ('due:'):
            # check first if there is a due date already 
            
            matchDue = re.search(r'due:(\d+)d$', myInputItem)
            matchDueHour = re.search(r'due:(\d+)d(\d{2}:\d{2})$', myInputItem)
            matchINT = re.match(r'^due:(\d{4}-\d{2}-\d{2})$', myInputItem)
            matchINThour = re.match(r'^due:(\d{4}-\d{2}-\d{2}T\d{2}:\d{2})$',myInputItem)

                        
            if matchDue:
               dueString = getNewDate (int(matchDue.group(1)),F=False)
               
            elif matchDueHour:
                myTime = matchDueHour.group(2)
                dueString = getNewDate (int(matchDueHour.group(1)),F=False)
                if validateTime (myTime):
                    dueString = f"{dueString}T{myTime}"
                else:
                    log ("invalid time") 
                

            elif matchINT:
                dueString = matchINT.group(1)
                #log (dueString)
            elif matchINThour:
                dueString = matchINThour.group(1)
                #log (dueString)                            
            else:
                customDays = myInputItem.split(':',1)[1]
                myInput = " ".join(myInputElements)
                dueMenu(customDays,myInput)
                exit()
            

    MYOUTPUT = {"items": []}
    myTaskElements = parseInput (myInput)
    
    for xxx in myTaskElements[:]:
        if xxx.startswith('@') or xxx.startswith('#') or xxx.startswith('^') or xxx.startswith('due:') or xxx in ['p1','p2','p3','p4']:
            myTaskElements.remove(xxx)
        
    
    MY_TASK_TEXT = " ".join (myTaskElements)
    
    myTagString = ",,..,,".join(myTags) # added complicated delim so that commas are allowed. will fail if a user has that in their labels
    if myTagString:
        myTagStringK = ",".join(myTags)
        myTagStringF = f"🏷️{myTagStringK}"
    else:
        myTagStringF = ''
    
    # dueString
    try:
        dueStringF = f"🗓️ due:{dueString}"
    except NameError:
        dueString = ""
        dueStringF = ""

    # section string
    try:
        mySectStringF = f"🧩 section:{taskSectionName}"
    except NameError:
        mySectString = ""
        mySectStringF = ""
        taskSectionID = None
    
    # priority string
    try:
        if prioString == 'p1':
            prioStringF = 'p1️⃣'
            myPriority = 4  #the priority in the API is inverted (4= top priority)

        elif prioString == 'p2':
            prioStringF = 'p2️⃣'
            myPriority = 3  
        elif prioString == 'p3':
            prioStringF = 'p3️⃣'
            myPriority = 2
        elif prioString == 'p4':
            prioStringF = ''
            myPriority = 1
    except NameError:
        myPriority = 1
        prioStringF = ''

    if taskProjectName:
        if "/" in taskProjectName:
            myProj = taskProjectName.split("/")[0]
            mySect = taskProjectName.split("/")[0]
            try:
                myProjStringF = f"📋{myProj}"
                mySectStringF = f"🧩 section:{mySect}"
                
            except NameError:
                if mySectStringF:
                    taskProjectName = section_ParentProjects[taskSectionName[1:]]
                    myProjStringF = f"📋{taskProjectName}"
                    taskProjectID = get_project_id (myProjects,taskProjectName[1:])

        else:
            try:
                myProjStringF = f"📋{taskProjectName}"
            
            except NameError:
                if mySectStringF:
                    taskProjectName = section_ParentProjects[taskSectionName[1:]]
                    myProjStringF = f"📋{taskProjectName}"
                    taskProjectID = get_project_id (myProjects,taskProjectName[1:])
    else:
        taskProjectName = "#Inbox"
        myProjStringF = f"📋{taskProjectName}"
        taskProjectID = get_project_id (myProjects,taskProjectName[1:])


            

    MYOUTPUT["items"].append({
                "title": MY_TASK_TEXT,
                "subtitle": f"{myProjStringF} {mySectStringF} {myTagStringF} {prioStringF} {dueStringF} ⇧↩️ to create",
                "arg": myInput,
                "variables" : {
                "myTaskText": MY_TASK_TEXT,
                "myTagString": myTagString,
                "myProjectID": taskProjectID,
                "mySectionID": taskSectionID,
                "myDueDate": dueString,
                "myPriority": myPriority
                    },
                "icon": {
                        "path": f"icons/newTask.png"
                    }
                })
   
    print (json.dumps(MYOUTPUT))
    

def handleINTdate(dateString):
    """
    this function takes a date string in international format and returns 1) the number of days to that date, and 2) a string with the date in long format
    """
    # Convert the date string to a datetime object
    date = datetime.strptime(dateString, "%Y-%m-%d").date()

    # Get the current date
    today = datetime.now().date()

    # Calculate the number of days between the current date and the given date
    daysTo = (date - today).days

    # Format the date in the desired format
    formatted_date = date.strftime("%A, %B %d, %Y")

    return daysTo,formatted_date

def handleINTdateHour(dateString):
    """
    this function takes a date string in international format (plus time) and returns 1) the number of days to that date, and 2) a string with the date in long format
    """
    # Convert the date string to a datetime object
    date = datetime.strptime(dateString, "%Y-%m-%dT%H:%M")

    # Format the date in the desired format
    formatted_date = date.strftime("%A, %B %d, %Y, %H:%M")

    # Get the current date
    today = datetime.now()

    # Calculate the number of days between the current date and the given date
    daysTo = (date - today).days

    return daysTo,formatted_date


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
    # the F argument is for formatting (year at the end). If false, it will return an international date string format
    timeToday = date.today()
    delta = timedelta(days=myDays)
    finalTime = timeToday+delta
    
    if F == True:
        finalTimeF = finalTime.strftime("%A, %B %d, %Y")
        return finalTimeF
    else:
        finalTimeFS = finalTime.strftime("%Y-%m-%d")
        return finalTimeFS

def reschMenu(customDays):
    taskContent = os.getenv('myTaskContent')
    pattern = r'^(\d+)([wm]?)$'
    pattern = r'^(\d+)([wm]?)(?:([01]\d|2[0-3]):([0-5]\d))?$' 
    match = re.match(pattern, customDays)
    matchINT = re.match(r'^\d{4}-\d{2}-\d{2}$', customDays)
    matchINThour = re.match(r'^\d{4}-\d{2}-\d{2}T\d{2}:\d{2}$',customDays)


    
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
        num_str, letter, myHours,myMinutes = match.groups()
        if myHours and myMinutes:
            myTimeString = (f", {myHours}:{myMinutes}")
            myTimeStringSF = f"{myHours}:{myMinutes}"
        else:
            myTimeString = myTimeStringSF = ''
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
                        "title": f"Reschedule in {num:,} {dayString} 🗓️ {getNewDate(num)}{myTimeString}",
                        "subtitle": taskContent,
                        "arg": f"{num}d{myTimeStringSF}",
                        "variables" : {
                            },
                        "icon": {
                                "path": f"icons/today.png"
                            }
                        })
    elif matchINT:
        intDate = matchINT.group()
        daysTo, dateF = handleINTdate (intDate)
        if daysTo ==1:
            dayString = 'day'
        else:
            dayString = 'days'
        MYOUTPUT["items"].append({
                        "title": f"Reschedule in {daysTo:,} {dayString} 🗓️ {dateF}",
                        "subtitle": taskContent,
                        "arg": intDate,
                        "variables" : {
                            },
                        "icon": {
                                "path": f"icons/today.png"
                            }
                        })
    elif matchINThour:
        intDate = matchINThour.group()
        daysTo, dateF = handleINTdateHour (intDate)
        if daysTo ==1:
            dayString = 'day'
        else:
            dayString = 'days'
        MYOUTPUT["items"].append({
                        "title": f"Reschedule in {daysTo:,} {dayString} 🗓️ {dateF}",
                        "subtitle": taskContent,
                        "arg": intDate,
                        "variables" : {
                            },
                        "icon": {
                                "path": f"icons/today.png"
                            }
                        })
    else:
        MYOUTPUT["items"].append({
                        "title": f"Invalid format!",
                        "subtitle": "enter an integer (days) or add 'w' (weeks) or 'm' (months). Optional: time in 24h format",
                        "arg": '',
                        "variables" : {
                            },
                        "icon": {
                                "path": f"icons/warning.png"
                            }
                        })
    print (json.dumps(MYOUTPUT))



def validateTime(time_str):
    # Split the time string into hours and minutes

    hours, minutes = map(int, time_str.split(":"))
    # Check if hours and minutes are within valid ranges
    if 0 <= hours < 24 and 0 <= minutes < 60:
        return True
    else:
        return False



def dueMenu(customDays,inputThrough):
    """
    the goal of this function is to take the string after `due` and evaluate it to create a due date (and possibly time)
    Possibilities:
    1. single number: assumed days
    2. number plus d,w,m: calculate weeks, months
    3. #2, but with time after (e.g. 10d10:13)
    """
    
    pattern = r'^(\d+)([wm]?)(?:([01]\d|2[0-3]):([0-5]\d))?$' 
    # matches a number, with or without [wm], with or without time in 24h format
    
    #pattern = r'^(\d+)([wm]?)$'
    
    log (f"input:{customDays}")
    match = re.match(pattern, customDays)

    patternDue = r'(?<=due:)\d*[wm]?(?:([01]\d|2[0-3]):([0-5]\d))?'
    
    inputThroughF = re.sub(patternDue, '', inputThrough) #replace the due: string in the passthrough text
    
    
    MYOUTPUT = {"items": []}
    
    if customDays == '': #if no days are entered, show predefined menu options
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
        
    elif match: #if there is a match (one number at least)
        num_str, letter, myHours,myMinutes = match.groups()
        
        if myHours and myMinutes:
            myTimeString = (f", {myHours}:{myMinutes}")
            myTimeStringSF = f"{myHours}:{myMinutes}"
        else:
            myTimeString = myTimeStringSF = ''
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
                        "title": f"Due in {num:,} {dayString} 🗓️ {getNewDate(num)}{myTimeString}",
                        "subtitle": '',
                        "arg": f"{inputThroughF}{num}d{myTimeStringSF} ",
                        "variables" : {
                            },
                        "icon": {
                                "path": f"icons/today.png"
                            }
                        })
    
    else:
        MYOUTPUT["items"].append({
                        "title": f"Invalid format!",
                        "subtitle": "enter an integer (days) or add 'w' (weeks) or 'm' (months). Optional: time in 24h format",
                        "arg": '',
                        "variables" : {
                            },
                        "icon": {
                                "path": f"icons/warning.png"
                            }
                        })
    print (json.dumps(MYOUTPUT))


def rescheduleTask (days,taskID):
    if '-' in days: #if the full INT date was provided
        newDate = days
    elif ':' in days: #date and time were provided
        days,myTime = days.split('d',1)
        newDate = getNewDate(int(days),F=False)
        if myTime:
            newDate = f"{newDate}T{myTime}"
    else:
        
        newDate = getNewDate(int(days),F=False)
    
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
    mySource = os.getenv('mySource')
    
    if mySource == 'createLabel':
        # apparently it is not possible to reset an environment variable (tried os.environ['mySource'] = "") without restarting the script filter object
        # this would result in the filter trying to create the same new label multiple times. 
        # this workaround checks if the label is in the list of labels
        # might not be needed if there is a way to reset the environment variable mySource without restarting the script filter
        
        myNewLabel = os.getenv('myNewLabel')
        
        with open(MY_LABEL_COUNTS,'r') as myFile: # checking that the new label has not been already added
            label_counts = json.load(myFile)
            myLabelListAll = list (label_counts)
        
        if myNewLabel not in myLabelListAll:
            createLabel (myNewLabel)
        

    if MY_COMMAND == "complete":
        completeTask (MY_INPUT)

    if MY_COMMAND == "findDate":
        customDays = sys.argv[3] 
        reschMenu (customDays)

    if MY_COMMAND == "reschedule":
        daysReschedule = sys.argv[2]
        myTaskID = os.getenv('myTaskID')
        rescheduleTask(daysReschedule,myTaskID)

    if MY_COMMAND == "rebuild":
        forceRebuild()


    if MY_COMMAND == "newTask":
        taskText = os.getenv('myTaskText')
        taskLabels = os.getenv('myTagString')
        taskProjectID = os.getenv('myProjectID')
        taskSectionID = os.getenv('mySectionID')
        myDueDate = os.getenv('myDueDate')
        myPriority = os.getenv('myPriority')
        
        createNewTask (taskText,taskLabels,taskProjectID, taskSectionID, myDueDate, myPriority)

    if MY_COMMAND == "parse":
     
        parseNewTask (MY_INPUT)


if __name__ == '__main__':
    main ()
    