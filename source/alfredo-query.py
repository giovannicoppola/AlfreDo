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
from config import SHOW_GOALS, PARTIAL_MATCH
import re


def parseInput(MY_INPUT):
    
    pattern = r'\s*([@#]\([^)]+\)|\S+)\s*' #keeps together elements with space if they are in parentheses and preceded by # or @
    result = re.findall(pattern, MY_INPUT)
    
    return (result)

def main():
    today = datetime.now().strftime("%Y-%m-%d")
    todayDate = datetime.today()
    checkingTime() #checking if the database needs to be rebuilt

    # reading data in
    allTasks, mySections, myProjects, myStats, myUser = readTodoistData()
        

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
    

    FINAL_INPUT = INPUT_ITEMS = parseInput(MY_INPUT)
    mySearchStrings = []
    LABEL_FLAG = 0
    PROJECT_FLAG = 0
    
    log (f"input items: {INPUT_ITEMS}")
    
    for inputItem in INPUT_ITEMS:

        # an item starting with @ can be one of 5 things: 
            # 1. a complete, existing tag with spaces (will need parentheses)
            # 2. a complete, existing tag
            # 3. an incomplete tag the user is trying to add (will need drop down autocomplete)
            # 4. a non-existent tag (will be dealt with in the dropdown)
            # 5. a badly parsed tag (containing a space and a closed parenthesis)
        


        if inputItem.startswith('@'): # user trying to enter a tag
        #log (f"tag fragment: {inputItem}")
            
            # at the time of the first version of the workflow, parentheses and other special characters were not allowed in project names, and I used them to allow spaces in Alfred's window
            # in July 2023 they were allowed.
            if inputItem.startswith("@(") and inputItem.endswith(")") and " " in inputItem: #there is a space and AlfreDO introduced parentheses 
                inputItem = inputItem.replace("(","",1)
                inputItem = inputItem.replace(")","")
                inputItem = inputItem.strip()
                
            log (f"after checking: {inputItem}")
            if inputItem.strip() in myLabelsAll: # is this a real tag? 
            #log (f"real tag: {inputItem}")
            
                myFilterLabels.append (inputItem[1:]) # add a real tag to the list of labels
        
            else: #incomplete tag: needs dropdown
                
                LABEL_FLAG = 1
                myTagFrag = inputItem
                
                try:
                    FINAL_INPUT.remove(inputItem)
                except:
                    
                    MYOUTPUT["items"].append({
                    "title": "something is wrong with this label!",
                    "subtitle": "does it contain a space and a closed parenthesis, or other very special characters?",
                    "arg": "",
                        "variables" : {
                        "myIter": True,
                        "myArg": MY_INPUT+" "
                        },
                    "icon": {
                            "path": f"icons/Warning.png"
                        }
                    })
                    print (json.dumps(MYOUTPUT))
                    exit()
            
        elif inputItem.startswith('#'): # user trying to enter a project 
            if inputItem.startswith("#(") and inputItem.endswith(")") and " " in inputItem: # user trying to enter a project  and there is a space and AlfreDO introduced parentheses 
                inputItem = inputItem.replace("(","",1)
                inputItem = inputItem.replace(")","")
                inputItem = inputItem.strip()
                    
            
            log (f"after checking: {inputItem}")
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
            
            else:
                #log (f"project fragment: {inputItem}")
                #log (f"final input: {FINAL_INPUT}")
                PROJECT_FLAG = 1
                #myProjFrag = inputItem
                #log (inputItem)
                try:
                    FINAL_INPUT.remove(inputItem)
                except:
                
                    MYOUTPUT["items"].append({
                    "title": "something is wrong with this project name!",
                    "subtitle": "does it contain a space and a closed parenthesis, a forward slash, or other very special characters?",
                    "arg": "",
                    "variables" : {
                        "myIter": True,
                        "myArg": MY_INPUT+" "
                        },
                    "icon": {
                            "path": f"icons/Warning.png"
                        }
                    })
                    print (json.dumps(MYOUTPUT))
                    exit()
        

        else: # user trying to enter a search string
            #log (f"search string fragment: {inputItem}")
            mySearchStrings.append (inputItem)

    MY_INPUT = " ".join(FINAL_INPUT) #this is needed to allow multiple tags and projects in the input string

    #log (f"filterlabels: {myFilterLabels}")
    #log (f"filterprojects: {myFilterProjects}")
    # log (f"filtersections: {myFilterSections}")
    #log (f"search strings: {mySearchStrings}")
   
    
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
        #log (f"TAGfrag: {myTagFrag}")
        #log (f"myLabels: {myLabels}")
        
        if PARTIAL_MATCH == 1: #searches anywhere in the string
            mySubset = [i for i in myLabels if myTagFrag[1:].casefold() in i.casefold()]
        else: #searches from the start
            mySubset = [i for i in myLabels if myTagFrag.casefold() in i.casefold()]
        
        # adding a complete tag if the user selects it from the list
        if mySubset:
            for thislabel in mySubset:
                if " " in thislabel: #adding parentheses if there is a space in the project name
                    thisLabel_string = f"@({thislabel[1:]})"
                else:
                    thisLabel_string = thislabel
                    
                if MY_INPUT:
                    MY_ARG = f"{MY_INPUT} {thisLabel_string} "
                else:
                    MY_ARG = f"{thisLabel_string} "
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
                    "myIter": True,
                    "myArg": MY_INPUT+" ",
                    "myMode": MY_MODE
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
            
            if PARTIAL_MATCH == 1: #searches anywhere in the project string (default)
                mySubset = [i for i in myProjectList if myProjFrag[1:].casefold() in i.casefold()]
            else: #searches from the start
                mySubset = [i for i in myProjectList if myProjFrag.casefold() in i.casefold()]
            
            # adding a complete project name if the user selects it from the list
            if mySubset:
                for thisProj in mySubset:
                    if " " in thisProj: #adding parentheses if there is a space in the project name
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
                    "myIter": True,
                    "myArg": MY_INPUT+" ",
                    "myMode": MY_MODE
                    
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
                    if 'Z' in task ['due']['date']:
                        dueDate = datetime.strptime(task ['due']['date'], "%Y-%m-%dT%H:%M:%SZ")
                    else:
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
             "mods": {
                    "alt": {
                        
                        "arg": "",
                        "subtitle": ""
                                },
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
                 "variables" : {
                    "myIter": True,
                    "myArg": "",
                    "myMode": MY_MODE
                    
                    },
                "mods": {
                    "shift": {
                        
                        "arg": "",
                        "subtitle": "nothing to see here"
                                },
                    "cmd": {
                        
                        "arg": "",
                        "subtitle": "nothing to see here"
                                },         
                    "ctrl": {
                        
                        "arg": "",
                        "subtitle": "nothing to see here"
                                },         
                    "alt": {
                        
                        "arg": "",
                        "subtitle": "nothing to see here"
                                },         
                    
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


