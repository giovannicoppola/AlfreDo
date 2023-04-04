#!/usr/bin/env python3

import requests
import json
from lib.requests.structures import CaseInsensitiveDict
from datetime import datetime
# AlfreDo – a Todoist workflow
# Monday, January 24, 2022, 7:32 PM

#pip3 install todoist-api-python #tried the SDK but it creates a number of classes that I had hard time to convert to JSON objects I was familar with
#pip3 install todoist-python  #this is what I ended up using
#pip3 install attrs #was needed for the SDK

#pip3 install -U --target=. tld todoist-python #used to install a local copy of the dependencies, so that  Idon't rely on the user to upload the library
# might not need this at all 

# pip3 install -U --target=. tld requests

"""import json

#from todoist_api_python.api import TodoistAPI 
from todoist.api import TodoistAPI
from datetime import datetime




api = TodoistAPI('f919fc636de42e7966cfa27c51742e6e0a1e4ef9')
api.sync()

#print(api.state['projects'])

#myProjects = api.state['projects']
#print (type(myProjects))

#print (myProjects[1])

today = datetime.utcnow().strftime("%Y-%m-%d")
MYOUTPUT = {"items": []}
countR=1

for task in api.state['items']:
    if task['due'] is not None and task['due']['date'] >= today:
        myContent = task ['content'] 
        MYOUTPUT["items"].append({
    "title": myContent,
    "subtitle": countR, 
    "arg": ""
    })
    countR += 1


print (json.dumps(MYOUTPUT))"""


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


        

for task in myData:
    if 'due' in task and task['due']['date'] <= today:
        myContent = task ['content'] 
        myDue = task ['due']['date']
        MYOUTPUT["items"].append({
        "title": myContent,
        "subtitle": myDue + "-"+ str(countR)+"/"+str(myMatchCount), 
        "arg": ""
        })
        countR += 1
    

print (json.dumps(MYOUTPUT))
