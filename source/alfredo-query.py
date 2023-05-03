#!/usr/bin/env python3

"""
# AlfreDo – a Todoist workflow
# using the todoist sync API
# Sunny ☀️   🌡️+18°F (feels +11°F, 59%) 🌬️↘7mph 🌗 2022-01-26 Wed 8:36AM
# restarted on 
# Partly cloudy ⛅️  🌡️+64°F (feels +64°F, 70%) 🌬️↖9mph 🌕&m Tue Apr  4 18:23:14 2023

"""

import json
from datetime import datetime
import sys
from alfredo_ops import log, get_project_id, get_project_name,fetchLabels,fetchProjects, checkingTime,readTodoistData
from config import SHOW_GOALS




def main():
    today = datetime.now().strftime("%Y-%m-%d")
    todayDate = datetime.today()
    checkingTime() #checking if the database needs to be rebuilt

    # reading data in
    allTasks, myProjects, myStats, myUser = readTodoistData()
    

    myFilterLabels = []
    myFilterProjects = []

    MYOUTPUT = {"items": []}
    countR=1

    
    MY_MODE = sys.argv[1]  # source: due today, all, overdue
    MY_INPUT = sys.argv[2]  # search string

    
    # prepare goals string if the user wants it
    if SHOW_GOALS == 1:
        # collecting stats
        DailyGoal = myUser['daily_goal']
        WeeklyGoal = myUser['weekly_goal']

        #tasks completed today
        todays = [item for item in myStats['days_items'] if item['date'] == today]
        SoFarCompleted = todays[0]['total_completed']
        
        #tasks completed this week
        TotalWeekCompleted = myStats['week_items'][0]['total_completed']

        if SoFarCompleted >= DailyGoal:
            statusDay = "✅"
        else:
            statusDay = "❌"

        if TotalWeekCompleted >= WeeklyGoal:
            statusWeek = "✅"
        else:
            statusWeek = "❌"
        GOALS_STRING = f" Daily: {SoFarCompleted}/{DailyGoal} {statusDay} Weekly: {TotalWeekCompleted}/{WeeklyGoal} {statusWeek} " 
    else:
        GOALS_STRING = ""
    
    

    if MY_MODE == "today":
        toShow = [task for task in allTasks if task['due'] and task['due']['date'] == today] # selecting tasks with due date = today
        if toShow:
            toShow = sorted(toShow, key = lambda i: i['due']['date']) #sorting by due date
        MYICON = 'icons/today.png'
        
        

    elif MY_MODE == "due":
        toShow = [task for task in allTasks if task['due'] and task['due']['date'] < today] # selecting tasks with due date before today
        if toShow:
            toShow = sorted(toShow, key = lambda i: i['due']['date']) #sorting by due date
        MYICON = 'icons/overdue.png'

    elif MY_MODE == "all":
        toShow = allTasks
        def get_due_date(item):
            if item['due'] is not None and 'date' in item['due']:
                return item['due']['date']
            else:
                return '9999-12-31' # a very late date
        # Sort the list using the get_due_date function as the key
        toShow = sorted(toShow, key=lambda i: get_due_date(i))

        MYICON = 'icons/bullet.png'

    label_counts, myLabelsAll = fetchLabels(toShow)
    project_counts,myProjectListAll = fetchProjects(toShow,myProjects)
    
    # evaluating the input string
    FINAL_INPUT = INPUT_ITEMS = MY_INPUT.split()
    
    
    mySearchStrings = []
    LABEL_FLAG = 0
    PROJECT_FLAG = 0
    
    
    for inputItem in INPUT_ITEMS:
        #log (inputItem)
        if inputItem.strip() in myLabelsAll: # is this a real tag? 
            #log (f"real tag: {inputItem}")
            myFilterLabels.append (inputItem[1:])
            
        elif inputItem in myProjectListAll: # is this a real project? :
            #log (f"real project: {inputItem}")
            idProj = get_project_id(myProjects, inputItem[1:])
            myFilterProjects.append (idProj)
        
        elif inputItem.startswith('@'): # user trying to enter a tag
            #log (f"tag fragment: {inputItem}")
            LABEL_FLAG = 1
            myTagFrag = inputItem
            FINAL_INPUT.remove(inputItem)
        
        elif inputItem.startswith('#'): # user trying to enter a project
            #log (f"project fragment: {inputItem}")
            PROJECT_FLAG = 1
            myProjFrag = inputItem
            FINAL_INPUT.remove(inputItem)

        else: # user trying to enter a search string
            #log (f"search string fragment: {inputItem}")
            mySearchStrings.append (inputItem)
            
    MY_INPUT = " ".join(FINAL_INPUT) #this is needed to allow multiple tags and projects in the input string

    # log (f"filterlabels: {myFilterLabels}")
    # log (f"filterprojects: {myFilterProjects}")
    # log (f"search strings: {mySearchStrings}")
   
    toShow = [item for item in toShow if (
        all(label in item.get('labels', []) for label in myFilterLabels) and 
        all(project in item.get('project_id', []) for project in myFilterProjects) and 
        all(substring.casefold() in item['content'].casefold() for substring in mySearchStrings)
        )]
    
    
    if LABEL_FLAG == 1:
        label_counts, myLabels = fetchLabels(toShow)
        mySubset = [i for i in myLabels if myTagFrag.casefold() in i.casefold()]
        
        # adding a complete tag if the user selects it from the list
        if mySubset:
            for thislabel in mySubset:
                if MY_INPUT:
                    MY_ARG = f"{MY_INPUT} {thislabel} "
                else:
                    MY_ARG = f"{thislabel} "
                MYOUTPUT["items"].append({
                "title": f"{thislabel} ({label_counts[thislabel[1:]]})",
                "subtitle": MY_INPUT,
                "arg": "",
                "variables" : {
                    "myIter": True,
                    "myArg": MY_ARG,
                    "myMode": MY_MODE
                    },
                "icon": {
                        "path": f"icons/label.png"
                    }
                })
        else:
            MYOUTPUT["items"].append({
            "title": "no labels matching",
            "subtitle": "try another query?",
            "variables" : {
                    
                    "myArg": MY_INPUT+" "
                    },
            "arg": "",
            "icon": {
                    "path": f"icons/Warning.png"
                }
            })
        print (json.dumps(MYOUTPUT))
        exit()

    if PROJECT_FLAG == 1:
            project_counts, myProjectList = fetchProjects(toShow,myProjects)
            mySubset = [i for i in myProjectList if myProjFrag.casefold() in i.casefold()]
            
            # adding a complete project name if the user selects it from the list
            if mySubset:
                for thisProj in mySubset:
                    if MY_INPUT:
                        MY_ARG = f"{MY_INPUT} {thisProj} "
                    else:
                        MY_ARG = f"{thisProj} "
                    MYOUTPUT["items"].append({
                    "title": f"{thisProj} ({project_counts[thisProj[1:]]})",
                    "subtitle": MY_INPUT,
                    "arg": '',
                    "variables" : {
                        "myIter": True,
                        "myArg": MY_ARG,
                        "myMode": MY_MODE
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
    
    if toShow:
        
        for task in toShow:
            
            myContent = task ['content'] 
            myMatchCount = len(toShow)
            if task ['due']:
                dueDate =datetime.strptime(task ['due']['date'] , '%Y-%m-%d')
                if dueDate:
                    dueDays = todayDate - dueDate
                    if abs(dueDays.days) == 1:
                        dayCountString = "day"
                    else:
                        dayCountString = "days"
                    if dueDays.days == 0:
                        dueString = "DUE TODAY"
                    elif dueDays.days < 0:
                        dueString = f"due in {abs(dueDays.days):,} {dayCountString} ⚠️"
                    elif dueDays.days > 0:        
                        dueString = f"{dueDays.days:,} {dayCountString} overdue❗"
                else: 
                    dueString = ""
            else: 
                dueString = ""
            
            if task['labels']:
                
                myLabelsString = f"🏷️ {','.join(task['labels'])}"
            else:
                myLabelsString = ""
            
            
            myProjectName = get_project_name(myProjects, task['project_id'])
    

            MYOUTPUT["items"].append({
            "title": f"{myContent} (#{myProjectName}) {dueString}",
            
            "subtitle": f"{countR}/{myMatchCount}.{GOALS_STRING}{myLabelsString}", 
             "variables" : {
                    "myIter": False,
                    "myURL": f"https://todoist.com/app/project/{task['project_id']}/task/{task['id']}",
                    "myTaskID": task['id'],
                    "myTaskContent": myContent,
                    "myArg": MY_INPUT,
                    "myMode": MY_MODE
                    },
            "icon": {
                    "path": MYICON
                },
            
            "arg": ""  
            })
            countR += 1
        

        print (json.dumps(MYOUTPUT))
    
    elif mySearchStrings or myFilterLabels:
        MYOUTPUT["items"].append({
                "title": "no tasks matching your query 🙁",
                "subtitle": "" ,
                
                "mods": {
                    "shift": {
                        
                        "arg": "",
                        "subtitle": "nothing to see here"
                                }
                    
                        },
                "arg": ""
                })
        print (json.dumps(MYOUTPUT))
    
    else: 
        
        MYOUTPUT["items"].append({
                "title": "no tasks left to do today! 🙌",
                "subtitle": f"{GOALS_STRING}",
                "mods": {
                    "shift": {
                        
                        "arg": "",
                        "subtitle": "nothing to see here"
                    }
        
                        },
                "arg": ""
                })
        log (MYOUTPUT)
        print (json.dumps(MYOUTPUT))

if __name__ == '__main__':
    main ()



""" OLDER CODE
# extracting any full tags from current input, adding them to the list to filter
    # fullTags = re.findall('@[^ ]+ ', MY_INPUT)
    # fullTags = [s.strip() for s in fullTags]
    # mySearchInput = MY_INPUT.strip()

    # if bool (fullTags):
    #     for currTag  in fullTags:
    #         if currTag.strip() in myLabels: #if it is a real tag
    #             myFilterLabels.append(currTag[1:].strip()) #adds to the list of tags to filter for
    #             mySearchInput = re.sub(currTag, '', mySearchInput).strip() #subtracts the tag from the search string
        
    # # check if the user is trying to enter a tag
    # MYMATCH = re.search(r'(?:^| )@[^ ]*$', MY_INPUT)
    # if (MYMATCH !=None):
    #     MYFLAG = MYMATCH.group(0).lstrip(' ')
    #     mySearchInput = re.sub(MYFLAG,'',MY_INPUT)
    #     # refining the subset to show based on tags and search string
    #     toShow = [item for item in toShow if (all(label in item.get('labels', []) for label in myFilterLabels)) and all(substring.casefold() in item['content'].casefold() for substring in mySearchInput.split())]
    # else:
    #     toShow = [item for item in toShow if (all(label in item.get('labels', []) for label in myFilterLabels)) and all(substring.casefold() in item['content'].casefold() for substring in mySearchInput.split())]

    # # check if the user is trying to enter a tag
    # MYMATCH = re.search(r'(?:^| )@[^ ]*$', MY_INPUT)
    # if (MYMATCH !=None):
        
    #     MYFLAG = MYMATCH.group(0).lstrip(' ')
    #     MY_INPUT = re.sub(MYFLAG,'',MY_INPUT)
    #     # refining the subset to show based on tags and search string
    #     toShow = [item for item in toShow if (all(label in item.get('labels', []) for label in myFilterLabels)) and all(substring.casefold() in item['content'].casefold() for substring in MY_INPUT.split())]
    #     log(f" search input: {mySearchInput}")
    #     log(f" input: {MY_INPUT}")
        # # getting all the tags (and counts) from the current subset of tasks
        # label_counts = {}
        # for item in toShow:
        #     for label in item.get('labels', []):
        #         if label in label_counts:
        #             label_counts[label] += 1
        #         else:
        #             label_counts[label] = 1
        # myLabels = list(label_counts.keys())
        # myLabels = ['@' + s for s in myLabels]
        # log (label_counts)
        # log (f"myLabels: {myLabels}")


        if MY_MODE == 'due':
                dueDate =datetime.strptime(task ['due']['date'] , '%Y-%m-%d')
                myDue = dueDate if task ['due'] else ""
                if myDue:
                    dueDays = todayDate - myDue
                    dueString = f"{dueDays.days:,} days overdue❗"
                else: 
                    dueString = ""
            elif MY_MODE == 'today':
                dueString = "DUE TODAY"
            else: 
                dueString = ""
            
            if task['labels']:
                
                myLabelsString = f"🏷️ {','.join(task['labels'])}"
            else:
                myLabelsString = ""
    
"""