#!/usr/bin/env python3

"""
CONFIG SCRIPT for the alfreDo Workflow for Todoist
Tuesday, April 26, 2022
"""



import os


TOKEN = os.path.expanduser(os.getenv('TOKEN', ''))

WF_BUNDLE = os.getenv('alfred_workflow_bundleid')
DATA_FOLDER = os.getenv('alfred_workflow_data')
MY_DATABASE = f"{DATA_FOLDER}/allData.json"
MY_LABEL_COUNTS = f"{DATA_FOLDER}/labelCounts.json"
MY_PROJECT_COUNTS = f"{DATA_FOLDER}/projectCounts.json"



SHOW_GOALS = int(os.getenv('SHOW_GOALS')) 
PARTIAL_MATCH = int(os.getenv('PARTIAL_MATCH')) 
RefRate = int(os.getenv('RefreshRate'))


if not os.path.exists(DATA_FOLDER):
    os.makedirs(DATA_FOLDER)

