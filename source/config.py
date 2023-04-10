#!/usr/bin/env python3
# -*- coding: utf-8 -*-

#
# Tuesday, April 26, 2022
#



import os


TOKEN = os.path.expanduser(os.getenv('TOKEN', ''))

WF_BUNDLE = os.getenv('alfred_workflow_bundleid')
DATA_FOLDER = os.path.expanduser('~')+"/Library/Application Support/Alfred/Workflow Data/"+WF_BUNDLE
MY_DATABASE = f"{DATA_FOLDER}/allData.json"
RefRate = int(os.getenv('RefreshRate'))

if not os.path.exists(DATA_FOLDER):
    os.makedirs(DATA_FOLDER)

