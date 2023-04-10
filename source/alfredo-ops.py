#!/usr/bin/env python3

import requests
import json
from requests.structures import CaseInsensitiveDict
from datetime import datetime
import sys
from alfredo_fun import log

import uuid

def generate_uuid():
    return str(uuid.uuid4())

# AlfreDo – a Todoist workflow
# Partly cloudy ⛅️  🌡️+31°F (feels +28°F, 82%) 🌬️↘4mph 🌗 2022-01-25 Tue 9:05AM

MY_TASK_ID = sys.argv[1]  


# url = "https://api.todoist.com/rest/v1/tasks/"+MY_TASK_ID+"/close"
# headers = CaseInsensitiveDict()
# headers["Authorization"] = "Bearer f919fc636de42e7966cfa27c51742e6e0a1e4ef9"
# resp = requests.post(url, headers=headers)
# print(resp.status_code)


url = "https://api.todoist.com/sync/v9/sync"
MY_UUID = generate_uuid()
headers = {
    "Authorization": "Bearer f919fc636de42e7966cfa27c51742e6e0a1e4ef9",
}

data = {
    "commands": json.dumps([
        {
            "type": "item_complete",
            "uuid": MY_UUID,
            "args": {
                "id": MY_TASK_ID,
             #   "date_completed": "2017-01-02T01:00:00.000000Z"
            }
        }
    ])
}

response = requests.post(url, headers=headers, data=data)

log(response.content)
