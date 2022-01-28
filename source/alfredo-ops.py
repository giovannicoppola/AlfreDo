#!/usr/bin/env python3

import requests
import json
from requests.structures import CaseInsensitiveDict
from datetime import datetime
import sys

# AlfreDo – a Todoist workflow
# Chappaqua – Partly cloudy ⛅️  🌡️+31°F (feels +28°F, 82%) 🌬️↘4mph 🌗 2022-01-25 Tue 9:05AM

MY_TASK_ID = sys.argv[1]  


url = "https://api.todoist.com/rest/v1/tasks/"+MY_TASK_ID+"/close"
headers = CaseInsensitiveDict()
headers["Authorization"] = "Bearer f919fc636de42e7966cfa27c51742e6e0a1e4ef9"
resp = requests.post(url, headers=headers)
print(resp.status_code)


