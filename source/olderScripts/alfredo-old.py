#!/usr/bin/env python3

import requests
import json
from requests.structures import CaseInsensitiveDict
from datetime import datetime
import sys

# AlfreDo – a Todoist workflow
# Monday, January 24, 2022, 7:32 PM

#pip3 install todoist-api-python #tried the SDK but it creates a number of classes that I had hard time to convert to JSON objects I was familar with
#pip3 install todoist-python  #this is what I ended up using
#pip3 install attrs #was needed for the SDK

#pip3 install -U --target=. tld todoist-python #used to install a local copy of the dependencies, so that  Idon't rely on the user to upload the library
# might not need this at all 

# pip3 install -U --target=. tld requests

#################################
# OLDER VERSION USING THE REST API
# NEW VERSION USES THE SYNC API WHICH SEEMS TO BE MORE POWERFUL AND WITH MORE OPTIONS
# BUT EQUIVALENT TO THIS BELOW TO GET THE DUE TASKS
###########################


MY_MODE = sys.argv[1]  
url = "https://api.todoist.com/rest/v1/tasks"

headers = CaseInsensitiveDict()
headers["Authorization"] = "Bearer f919fc636de42e7966cfa27c51742e6e0a1e4ef9"


resp = requests.get(url, headers=headers)
myData = resp.json()

today = datetime.utcnow().strftime("%Y-%m-%d")
MYOUTPUT = {"items": []}
countR=1
myMatchCount=1

for task in myData:  #counting the total number of tasks due
    if 'due' in task and task['due']['date'] <= today:
        myMatchCount+=1

dueDateItems = [task for task in myData if 'due' in task] # selecting tasks with a due date

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
        "subtitle": myDue + "-"+ str(countR)+"/"+str(myMatchCount) + "-" + str(dueToday), 
        "arg": str(task['id']) + ";;" + str(dueToday)
        })
        countR += 1
    

print (json.dumps(MYOUTPUT))
