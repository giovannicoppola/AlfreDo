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
from alfredo_ops import log, get_project_id, get_project_name,fetchLabels,fetchProjects, checkingTime,readTodoistData, fetchSections
from config import SHOW_GOALS
import re




def main():
    today = datetime.now().strftime("%Y-%m-%d")
    todayDate = datetime.today()
    checkingTime() #checking if the database needs to be rebuilt

    # reading data in
    allTasks, mySections, myProjects, myStats, myUser = readTodoistData()
    #log (f"===MY SECTIONS======={mySections}=========")
    

    myFilterLabels = []
    myFilterProjects = []
    myFilterSections = []

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
    
    
    #subsetting the tasks based on the mode
    if MY_MODE == "today":
        toShow = [task for task in allTasks if task['due'] and task['due']['date'].split("T")[0] == today] # selecting tasks with due date = today
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
    
    # getting all the tags, projects, sections (and counts) from the current subset of tasks
    label_counts, myLabelsAll = fetchLabels(toShow)
    project_counts,myProjectListAll = fetchProjects(toShow,myProjects,mySections)
    section_counts,mySectionListAll, section_ParentProjects = fetchSections(toShow,mySections,myProjects)
    #log (f"=========={section_counts}=========")
    #log (f"=========={mySectionListAll}=========")
    #log (f"=========={section_ParentProjects}=========")
    
    # evaluating the input string
    #FINAL_INPUT = INPUT_ITEMS = re.findall(r'\[[^\]]*\]|\([^)]*\)|"[^"]*"|#[^ ]*|@[^ ]*|\S+', MY_INPUT)
    #FINAL_INPUT = INPUT_ITEMS = MY_INPUT.split()
    

    def parseInput(MY_INPUT):
        
        pattern = r'\s*(#\([^)]+\)|\S+)\s*' #keeps together elements with space if they are in parenthese and preceded by #
        result = re.findall(pattern, MY_INPUT)
        
        return (result)

    FINAL_INPUT = INPUT_ITEMS = parseInput(MY_INPUT)
    mySearchStrings = []
    LABEL_FLAG = 0
    PROJECT_FLAG = 0
    #SECTION_FLAG = 0
    
    log (INPUT_ITEMS)
    
    for inputItem in INPUT_ITEMS:
        log (inputItem)
        if inputItem.strip() in myLabelsAll: # is this a real tag? 
            #log (f"real tag: {inputItem}")
            myFilterLabels.append (inputItem[1:])
            
        elif inputItem.startswith('#'): # user trying to enter a project 
            if "(" in inputItem: #there is a space and AlfreDO introduced parentheses (which are not allowed in project names)
                inputItem = inputItem.replace("(","")
                inputItem = inputItem.replace(")","")
                inputItem = inputItem.strip()
                
            
            #log (inputItem)
            if inputItem in myProjectListAll: # is this a real project? :
                if "/" in inputItem: #there is a section
                    inputProject = inputItem.split("/")[0]
                    inputSection = inputItem.split("/")[1]
                    idProj = get_project_id(myProjects, inputProject[1:])
                    idSect = get_project_id(mySections, inputSection) #this function should work for sections too
                    myFilterProjects.append (idProj)
                    myFilterSections.append (idSect)
            
                else:
                    #log (f"real project: {inputItem}")
                    idProj = get_project_id(myProjects, inputItem[1:])
                    #log (idProj)
                    myFilterProjects.append (idProj)
            
            else: # user trying to enter a project
                #log (f"project fragment: {inputItem}")
                PROJECT_FLAG = 1
                myProjFrag = inputItem
                #log (inputItem)
                FINAL_INPUT.remove(inputItem)
            
        
        elif inputItem.startswith('@'): # user trying to enter a tag
            #log (f"tag fragment: {inputItem}")
            LABEL_FLAG = 1
            myTagFrag = inputItem
            FINAL_INPUT.remove(inputItem)
        

        else: # user trying to enter a search string
            #log (f"search string fragment: {inputItem}")
            mySearchStrings.append (inputItem)

    log (myFilterProjects)
    log (myFilterSections)
    MY_INPUT = " ".join(FINAL_INPUT) #this is needed to allow multiple tags and projects in the input string

    # log (f"filterlabels: {myFilterLabels}")
    # log (f"filterprojects: {myFilterProjects}")
    # log (f"filtersections: {myFilterSections}")
    # log (f"search strings: {mySearchStrings}")
   
    
    toShow = [item for item in toShow if (
        all(label in item.get('labels', []) for label in myFilterLabels) and 
        all(project in item.get('project_id', '') for project in myFilterProjects) and 
        all(substring.casefold() in item['content'].casefold() for substring in mySearchStrings)
        )]
    

    if myFilterSections:
        toShow = [item for item in toShow if (
            all(label in item.get('labels', []) for label in myFilterLabels) and 
            all(project in item.get('project_id', '') for project in myFilterProjects) and 
            (item.get('section_id') is not None and all(section in item['section_id'] for section in myFilterSections)) and 
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
            project_counts, myProjectList = fetchProjects(toShow,myProjects,mySections)
            mySubset = [i for i in myProjectList if myProjFrag.casefold() in i.casefold()]
            
            # adding a complete project name if the user selects it from the list
            if mySubset:
                for thisProj in mySubset:
                    if " " in thisProj:
                        thisProj_string = f"#({thisProj[1:]})"
                    else:
                        thisProj_string = thisProj
                    if MY_INPUT:
                        MY_ARG = f"{MY_INPUT} {thisProj_string} "
                    else:
                        MY_ARG = f"{thisProj_string} "
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
    
    # if SECTION_FLAG == 1:
    #         section_counts,mySectionList, section_ParentProjects = fetchSections(toShow,mySections,myProjects)
    #         mySubset = [i for i in mySectionList if mySectFrag.casefold() in i.casefold()]
            
    #         # adding a complete project name if the user selects it from the list
    #         if mySubset:
    #             for thisSect in mySubset:
    #                 if MY_INPUT:
    #                     MY_ARG = f"{MY_INPUT} {thisSect} "
    #                 else:
    #                     MY_ARG = f"{thisSect} "
    #                 MYOUTPUT["items"].append({
    #                 "title": f"{thisSect} ({section_ParentProjects[thisSect[1:]]}, {section_counts[thisSect[1:]]})",
    #                 "subtitle": MY_INPUT,
    #                 "arg": '',
    #                 "variables" : {
    #                     "myIter": True,
    #                     "myArg": MY_ARG,
    #                     "myMode": MY_MODE
    #                     },
    #                 "icon": {
    #                         "path": f"icons/section.png"
    #                     }
    #                 })
    #         else:
    #             MYOUTPUT["items"].append({
    #             "title": "no sections matching",
    #             "subtitle": "try another query?",
    #             "arg": "",
    #              "variables" : {
                    
    #                 "myArg": MY_INPUT+" "
    #                 },
    #             "icon": {
    #                     "path": f"icons/Warning.png"
    #                 }
    #             })
    #         print (json.dumps(MYOUTPUT))
    #         exit()    
    
    if toShow:
        
        for task in toShow:
            
            myContent = task ['content'] 
            myMatchCount = len(toShow)
            if task ['due']:
                if 'T' in task ['due']['date']:
                    dueDate = datetime.strptime(task ['due']['date'], "%Y-%m-%dT%H:%M:%S")
                else:
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
                    "myAppURL": f"todoist://task?id={task['id']}",
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


